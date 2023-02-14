package engine

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/pion/ion-log"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	gst "github.com/yaxiongwu/remote-control-client-go2/pkg/gstreamer-src"
	"github.com/yaxiongwu/remote-control-client-go2/pkg/proto/rtc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// type Role int32

// const (
// 	Admin       Role = 0
// 	VideoSource Role = 1
// 	Control     Role = 2
// 	Observe     Role = 3
// 	Unknown     Role = 4
// )

const (
	API_CHANNEL = "ion-sfu"
)

// Call dc api
type Call struct {
	StreamID string `json:"streamId"`
	Video    string `json:"video"`
	Audio    bool   `json:"audio"`
}

type TrackInfo struct {
	Id        string
	Kind      string
	Muted     bool
	Type      MediaType
	StreamId  string
	Label     string
	Subscribe bool
	Layer     string
	Direction string
	Width     uint32
	Height    uint32
	FrameRate uint32
}

type Subscription struct {
	TrackId   string
	Mute      bool
	Subscribe bool
	Layer     string
}

type Target int32

const (
	Target_PUBLISHER  Target = 0
	Target_SUBSCRIBER Target = 1
)

type MediaType int32

const (
	MediaType_MediaUnknown  MediaType = 0
	MediaType_UserMedia     MediaType = 1
	MediaType_ScreenCapture MediaType = 2
	MediaType_Cavans        MediaType = 3
	MediaType_Streaming     MediaType = 4
	MediaType_VoIP          MediaType = 5
)

type TrackEvent_State int32

const (
	TrackEvent_ADD    TrackEvent_State = 0
	TrackEvent_UPDATE TrackEvent_State = 1
	TrackEvent_REMOVE TrackEvent_State = 2
)

// TrackEvent info
type TrackEvent struct {
	State  TrackEvent_State
	Uid    string
	Tracks []*TrackInfo
}

var (
	DefaultConfig = RTCConfig{
		WebRTC: WebRTCTransportConfig{
			Configuration: webrtc.Configuration{
				ICEServers: []webrtc.ICEServer{
					{
						URLs: []string{"stun:stun.stunprotocol.org:3478", "stun:stun.l.google.com:19302"},
					},
				},
			},
		},
	}
)

// WebRTCTransportConfig represents configuration options
type WebRTCTransportConfig struct {
	// if set, only this codec will be registered. leave unset to register all codecs.
	VideoMime     string
	Configuration webrtc.Configuration
	Setting       webrtc.SettingEngine
}

type RTCConfig struct {
	WebRTC WebRTCTransportConfig `mapstructure:"webrtc"`
}

// Signaller sends and receives signalling messages with peers.
// Signaller is derived from rtc.RTC_SignalClient, matching the
// exported API of the GRPC Signal Service.
// Signaller allows alternative signalling implementations
// if the GRPC Signal Service does not fit your use case.
type Signaller interface {
	Send(request *rtc.Request) error
	Recv() (*rtc.Reply, error)
	CloseSend() error
}

// Client a sdk client
type RTC struct {
	Service
	connected      bool
	MaxTimeControl int
	MaxTimeView    int
	MaxClientsNum  int

	config *RTCConfig

	uid     string
	pub     *Transport
	sub     *Transport
	clients []*Client
	//controlClient *Client
	//viewClients map[string]*Client
	//export to user
	OnTrack              func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver)
	OnDataChannel        func(*webrtc.DataChannel)
	OnError              func(error)
	OnTrackEvent         func(event TrackEvent)
	OnDataChannelMessage func(webrtc.DataChannelMessage)
	OnSpeaker            func(event []string)
	ControlFunc          func(control string, data float64)

	producer *WebMProducer
	recvByte int
	notify   chan struct{}

	//cache datachannel api operation before dr.OnOpen
	apiQueue []Call

	signaller Signaller

	ctx        context.Context
	cancel     context.CancelFunc
	handleOnce sync.Once
	sync.Mutex
	OnPubIceConnectionStateChange func(webrtc.ICEConnectionState)
	OnSubIceConnectionStateChange func(webrtc.ICEConnectionState)
	OnIceConnectionStateChange    func(webrtc.ICEConnectionState, *webrtc.PeerConnection)
	VedioTrack                    *webrtc.TrackLocalStaticSample
	AudioTrack                    *webrtc.TrackLocalStaticSample
}

type DataChannelMsgDataSdp struct {
	Type string `json:"type"`
	Sdp  string `json:"sdp"`
}
type DataChannelMsg struct {
	Cmd string `json:"cmd"`
	//Data string `json:"data"`
	Data interface{} `json:"data"`
}

func withConfig(config ...RTCConfig) *RTC {
	r := &RTC{
		notify: make(chan struct{}),
	}
	r.ctx, r.cancel = context.WithCancel(context.Background())

	if len(config) > 0 {
		r.config = &config[0]
	}

	return r
}

// NewRTC creates an RTC using the default GRPC signaller
func NewRTC(connector *Connector, config ...RTCConfig) (*RTC, error) {
	r := withConfig(config...)
	signaller, err := connector.Signal(r)
	r.start(signaller)
	audioSrc := " autoaudiosrc ! audio/x-raw"
	//omxh264enc可能需要设置长宽为32倍整数，否则会出现"green band"，一道偏色栏
	videoSrc := " autovideosrc ! video/x-raw, width=640, height=480 ! videoconvert ! queue"
	//videoSrc := flag.String("video-src", "videotestsrc", "GStreamer video src")
	flag.Parse()
	// Create a video track
	//videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion2")
	//var err error
	r.VedioTrack, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/H264"}, "video", "pion2")
	if err != nil {
		panic(err)
	}
	// Create a audio track
	r.AudioTrack, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		panic(err)
	}

	//rtmpudp := rtmpudp.Init("5000")
	//gst.CreatePipeline("vp8", []*webrtc.TrackLocalStaticSample{videoTrack}, videoSrc).Start()
	//gst.CreatePipeline("h264_x264", []*webrtc.TrackLocalStaticSample{videoTrack}, videoSrc, rtmpudp.GetConn()).Start()
	//gst.CreatePipeline("opus", []*webrtc.TrackLocalStaticSample{audioTrack}, audioSrc, rtmpudp.GetConn()).Start()
	gst.CreatePipeline("h264_x264", []*webrtc.TrackLocalStaticSample{r.VedioTrack}, videoSrc).Start()
	gst.CreatePipeline("opus", []*webrtc.TrackLocalStaticSample{r.AudioTrack}, audioSrc).Start()

	return r, err
}

// NewRTCWithSignaller creates an RTC with a specified signaller
func NewRTCWithSignaller(signaller Signaller, config ...RTCConfig) *RTC {
	r := withConfig(config...)
	r.start(signaller)
	return r
}

func (r *RTC) start(signaller Signaller) {
	r.signaller = signaller

	if !r.Connected() {
		r.Connect()
	}
}

// GetPubStats get pub stats
func (r *RTC) GetPubStats() webrtc.StatsReport {
	return r.pub.pc.GetStats()
}

// GetSubStats get sub stats
func (r *RTC) GetSubStats() webrtc.StatsReport {
	return r.sub.pc.GetStats()
}

func (r *RTC) GetPubTransport() *Transport {
	return r.pub
}

func (r *RTC) GetSubTransport() *Transport {
	return r.sub
}

// Publish local tracks
func (r *RTC) Publish(tracks ...webrtc.TrackLocal) ([]*webrtc.RTPSender, error) {
	var rtpSenders []*webrtc.RTPSender
	for _, t := range tracks {
		if rtpSender, err := r.pub.GetPeerConnection().AddTrack(t); err != nil {
			log.Errorf("AddTrack error: %v", err)
			return rtpSenders, err
		} else {
			rtpSenders = append(rtpSenders, rtpSender)
		}

	}
	r.onNegotiationNeeded()
	return rtpSenders, nil
}

// UnPublish local tracks by transceivers
func (r *RTC) UnPublish(senders ...*webrtc.RTPSender) error {
	for _, s := range senders {
		if err := r.pub.pc.RemoveTrack(s); err != nil {
			return err
		}
	}
	r.onNegotiationNeeded()
	return nil
}

// CreateDataChannel create a custom datachannel
func (r *RTC) CreateDataChannel(label string) (*webrtc.DataChannel, error) {
	log.Debugf("id=%v CreateDataChannel %v", r.uid, label)
	return r.pub.pc.CreateDataChannel(label, &webrtc.DataChannelInit{})
}

// trickle receive candidate from sfu and add to pc
func (r *RTC) trickle(candidate webrtc.ICECandidateInit, target Target) {
	log.Debugf("[S=>C] id=%v candidate=%v target=%v", r.uid, candidate, target)
	var t *Transport
	if target == Target_SUBSCRIBER {
		t = r.sub
	} else {
		t = r.pub
	}

	if t.pc.CurrentRemoteDescription() == nil {
		t.RecvCandidates = append(t.RecvCandidates, candidate)
	} else {
		err := t.pc.AddICECandidate(candidate)
		if err != nil {
			log.Errorf("id=%v err=%v", r.uid, err)
		}
	}
}

// receiveTrickle2 receive candidate from sfu and add to pc
func (r *RTC) receiveTrickle2(candidate webrtc.ICECandidateInit, from string) {
	log.Debugf("[S=>C] id= candidate=%v from=%v", candidate, from)
	for _, client := range r.clients {
		if client.Id == from {
			log.Debugf("candidates from :%v", from)
			if client.pubPc.CurrentRemoteDescription() == nil {
				log.Debugf("client.pubPc.CurrentRemoteDescription() == nil ")
				client.pubRecvCandidates = append(client.pubRecvCandidates, candidate)
			} else {
				log.Debugf("client.pubPc.AddICECandidate() candidate:%v", candidate)
				err := client.pubPc.AddICECandidate(candidate)
				if err != nil {
					log.Errorf("to=%v err=%v", from, err)
				}
			}
		}
	}

}

// negotiate sub negotiate
func (r *RTC) negotiate(sdp webrtc.SessionDescription) error {
	//log.Debugf("[S=>C] id=%v Negotiate sdp=%v", r.uid, sdp)
	// 1.sub set remote sdp
	err := r.sub.pc.SetRemoteDescription(sdp)
	if err != nil {
		log.Errorf("id=%v Negotiate r.sub.pc.SetRemoteDescription err=%v", r.uid, err)
		return err
	}

	// 2. safe to send candiate to sfu after join ok
	if len(r.sub.SendCandidates) > 0 {
		for _, cand := range r.sub.SendCandidates {
			log.Debugf("[C=>S] id=%v send sub.SendCandidates r.uid, r.rtc.trickle cand=%v", r.uid, cand)
			r.SendTrickle(cand, Target_SUBSCRIBER)
		}
		r.sub.SendCandidates = []*webrtc.ICECandidate{}
	}

	// 3. safe to add candidate after SetRemoteDescription
	if len(r.sub.RecvCandidates) > 0 {
		for _, candidate := range r.sub.RecvCandidates {
			log.Debugf("id=%v r.sub.pc.AddICECandidate candidate=%v", r.uid, candidate)
			_ = r.sub.pc.AddICECandidate(candidate)
		}
		r.sub.RecvCandidates = []webrtc.ICECandidateInit{}
	}

	// 4. create answer after add ice candidate
	answer, err := r.sub.pc.CreateAnswer(nil)
	if err != nil {
		log.Errorf("id=%v err=%v", r.uid, err)
		return err
	}

	// 5. set local sdp(answer)
	err = r.sub.pc.SetLocalDescription(answer)
	if err != nil {
		log.Errorf("id=%v err=%v", r.uid, err)
		return err
	}

	// 6. send answer to sfu
	err = r.SendAnswer(answer)
	if err != nil {
		log.Errorf("id=%v err=%v", r.uid, err)
		return err
	}
	return err
}

func (r *RTC) Reconnect() {
	r.onNegotiationNeeded()
}

// onNegotiationNeeded will be called when add/remove track, but never trigger, call by hand
func (r *RTC) onNegotiationNeeded() {
	// 1. pub create offer
	offer, err := r.pub.pc.CreateOffer(nil)
	if err != nil {
		log.Errorf("id=%v err=%v", r.uid, err)
	}

	// 2. pub set local sdp(offer)
	err = r.pub.pc.SetLocalDescription(offer)
	if err != nil {
		log.Errorf("id=%v err=%v", r.uid, err)
	}

	//3. send offer to sfu
	err = r.SendOffer(offer)
	if err != nil {
		log.Errorf("id=%v err=%v", r.uid, err)
	}
}

// selectRemote select remote video/audio
func (r *RTC) selectRemote(streamId, video string, audio bool) error {
	log.Debugf("id=%v streamId=%v video=%v audio=%v", r.uid, streamId, video, audio)
	call := Call{
		StreamID: streamId,
		Video:    video,
		Audio:    audio,
	}

	// cache cmd when dc not ready
	if r.sub.api == nil || r.sub.api.ReadyState() != webrtc.DataChannelStateOpen {
		log.Debugf("id=%v append to r.apiQueue call=%v", r.uid, call)
		r.apiQueue = append(r.apiQueue, call)
		return nil
	}

	// send cached cmd
	if len(r.apiQueue) > 0 {
		for _, cmd := range r.apiQueue {
			log.Debugf("[C=>S] id=%v r.sub.api.Send cmd=%v", r.uid, cmd)
			marshalled, err := json.Marshal(cmd)
			if err != nil {
				continue
			}
			err = r.sub.api.Send(marshalled)
			if err != nil {
				log.Errorf("error: %v", err)
			}
			time.Sleep(time.Millisecond * 10)
		}
		r.apiQueue = []Call{}
	}

	// send this cmd
	log.Debugf("[C=>S] id=%v r.sub.api.Send call=%v", r.uid, call)
	marshalled, err := json.Marshal(call)
	if err != nil {
		return err
	}
	err = r.sub.api.Send(marshalled)
	if err != nil {
		log.Errorf("id=%v err=%v", r.uid, err)
	}
	return err
}

// PublishWebm publish a webm producer
func (r *RTC) PublishFile(file string, video, audio bool) error {
	if !FileExist(file) {
		return os.ErrNotExist
	}
	ext := filepath.Ext(file)
	switch ext {
	case ".webm":
		r.producer = NewWebMProducer(file, 0)
	default:
		return errInvalidFile
	}
	if video {
		videoTrack, err := r.producer.GetVideoTrack()
		if err != nil {
			log.Debugf("error: %v", err)
			return err
		}
		_, err = r.pub.pc.AddTrack(videoTrack)
		if err != nil {
			log.Debugf("error: %v", err)
			return err
		}
	}
	if audio {
		audioTrack, err := r.producer.GetAudioTrack()
		if err != nil {
			log.Debugf("error: %v", err)
			return err
		}
		_, err = r.pub.pc.AddTrack(audioTrack)
		if err != nil {
			log.Debugf("error: %v", err)
			return err
		}
	}
	r.producer.Start()
	//trigger by hand
	r.onNegotiationNeeded()
	return nil
}

func (r *RTC) trackEvent(event TrackEvent) {
	if r.OnTrackEvent == nil {
		log.Errorf("r.OnTrackEvent == nil")
		return
	}
	r.OnTrackEvent(event)
}

func (r *RTC) speaker(event []string) {
	if r.OnSpeaker == nil {
		log.Errorf("r.OnSpeaker == nil")
		return
	}
	r.OnSpeaker(event)
}

// setRemoteSDP pub SetRemoteDescription and send cadidate to sfu
func (r *RTC) setRemoteSDP(sdp webrtc.SessionDescription) error {
	err := r.pub.pc.SetRemoteDescription(sdp)
	if err != nil {
		log.Errorf("id=%v err=%v", r.uid, err)
		return err
	}

	// it's safe to add cand now after SetRemoteDescription
	if len(r.pub.RecvCandidates) > 0 {
		for _, candidate := range r.pub.RecvCandidates {
			log.Debugf("id=%v r.pub.pc.AddICECandidate candidate=%v", r.uid, candidate)
			err = r.pub.pc.AddICECandidate(candidate)
			if err != nil {
				log.Errorf("id=%v r.pub.pc.AddICECandidate err=%v", r.uid, err)
			}
		}
		r.pub.RecvCandidates = []webrtc.ICECandidateInit{}
	}

	// it's safe to send cand now after join ok
	if len(r.pub.SendCandidates) > 0 {
		for _, cand := range r.pub.SendCandidates {
			log.Debugf("id=%v r.rtc.trickle cand=%v", r.uid, cand)
			r.SendTrickle(cand, Target_PUBLISHER)
		}
		r.pub.SendCandidates = []*webrtc.ICECandidate{}
	}
	return nil
}

// setRemoteSDP pub SetRemoteDescription and send cadidate to sfu
func (r *RTC) setClientRemoteSDP(sdp webrtc.SessionDescription, from string) error {
	for _, client := range r.clients {
		if client.Id == from {
			log.Infof("get remote description from:%v", from)
			err := client.pubPc.SetRemoteDescription(sdp)
			if err != nil {
				log.Errorf("id=%v err=%v", r.uid, err)
				return err
			}
			//log.Infof("set remote description:%v", sdp)

			// it's safe to add cand now after SetRemoteDescription
			if client.pubRecvCandidates != nil {
				if len(client.pubRecvCandidates) > 0 {
					for _, candidate := range client.pubRecvCandidates {
						log.Debugf("id=%v client.pubPc..AddICECandidate candidate=%v", r.uid, candidate)
						err = client.pubPc.AddICECandidate(candidate)
						if err != nil {
							log.Errorf("id=%v r.pub.pc.AddICECandidate err=%v", r.uid, err)
						}
					}
					client.pubRecvCandidates = []webrtc.ICECandidateInit{}
				}
			}

			// it's safe to send cand now after join ok
			if client.pubSendCandidates != nil {
				if len(client.pubSendCandidates) > 0 {
					for _, cand := range client.pubSendCandidates {
						log.Debugf("[C=>S] id=%v send sub.SendCandidates r.uid, r.rtc.trickle cand=%v", from, cand)
						r.SendTrickle2(cand, from)
					}
					client.pubSendCandidates = []*webrtc.ICECandidate{}
				}
			}
		}
	}

	return nil
}

// GetBandWidth call this api cyclely
func (r *RTC) GetBandWidth(cycle int) (int, int) {
	var recvBW, sendBW int
	if r.producer != nil {
		sendBW = r.producer.GetSendBandwidth(cycle)
	}

	recvBW = r.recvByte / cycle / 1000
	r.recvByte = 0
	return recvBW, sendBW
}

func (r *RTC) Name() string {
	return "Room"
}

func (r *RTC) Connect() {
	go r.onSingalHandleOnce()
	r.connected = true
}

func (r *RTC) Connected() bool {
	return r.connected
}

func (r *RTC) onSingalHandleOnce() {
	// onSingalHandle is wrapped in a once and only started after another public
	// method is called to ensure the user has the opportunity to register handlers
	r.handleOnce.Do(func() {
		err := r.onSingalHandle()
		if r.OnError != nil {
			r.OnError(err)
		}
	})
}

func (r *RTC) SendJoin(sid string, uid string, offer webrtc.SessionDescription, config map[string]string) error {
	log.Infof("[C=>S] [%v] sid=%v", r.uid, sid)
	go r.onSingalHandleOnce()
	r.Lock()
	err := r.signaller.Send(
		&rtc.Request{
			Payload: &rtc.Request_Join{
				Join: &rtc.JoinRequest{
					Sid:    sid,
					Uid:    uid,
					Config: config,
					Description: &rtc.SessionDescription{
						Target: rtc.Target_PUBLISHER,
						Type:   "offer",
						Sdp:    offer.SDP,
					},
				},
			},
		},
	)
	r.Unlock()
	if err != nil {
		log.Errorf("[C=>S] [%v] err=%v", r.uid, err)
	}
	return err
}

func (r *RTC) SendTrickle(candidate *webrtc.ICECandidate, target Target) {
	log.Debugf("[C=>S] [%v] candidate=%v target=%v", r.uid, candidate, target)
	bytes, err := json.Marshal(candidate.ToJSON())
	if err != nil {
		log.Errorf("error: %v", err)
		return
	}
	go r.onSingalHandleOnce()
	r.Lock()
	err = r.signaller.Send(
		&rtc.Request{
			Payload: &rtc.Request_Trickle{
				Trickle: &rtc.Trickle{
					Target: rtc.Target(target),
					Init:   string(bytes),
				},
			},
		},
	)
	r.Unlock()
	if err != nil {
		log.Errorf("[%v] err=%v", r.uid, err)
	}
}

func (r *RTC) SendTrickle2(candidate *webrtc.ICECandidate, destination string) {
	//log.Debugf("[C=>S] [%v] candidate=%v uid=%v", r.uid, candidate, destination)
	log.Debugf("[C=>S] candidate=%v destination=%v", candidate, destination)
	bytes, err := json.Marshal(candidate.ToJSON())
	if err != nil {
		log.Errorf("error: %v", err)
		return
	}

	go r.onSingalHandleOnce()
	r.Lock()
	err = r.signaller.Send(
		&rtc.Request{
			Payload: &rtc.Request_Trickle{
				Trickle: &rtc.Trickle{
					To:   destination,
					Init: string(bytes),
				},
			},
		},
	)
	r.Unlock()
	if err != nil {
		log.Errorf("[%v] err=%v", r.uid, err)
	}
}

func (r *RTC) SendOffer(sdp webrtc.SessionDescription) error {
	log.Infof("[C=>S] [%v] sdp=%v", r.uid, sdp)
	go r.onSingalHandleOnce()
	r.Lock()
	err := r.signaller.Send(
		&rtc.Request{
			Payload: &rtc.Request_Description{
				Description: &rtc.SessionDescription{
					Target: rtc.Target_PUBLISHER,
					Type:   "offer",
					Sdp:    sdp.SDP,
				},
			},
		},
	)
	r.Unlock()
	if err != nil {
		log.Errorf("[%v] err=%v", r.uid, err)
		return err
	}
	return nil
}

func (r *RTC) SendAnswer(sdp webrtc.SessionDescription) error {
	log.Infof("[C=>S] [%v] sdp=%v", r.uid, sdp)
	r.Lock()
	err := r.signaller.Send(
		&rtc.Request{
			Payload: &rtc.Request_Description{
				Description: &rtc.SessionDescription{
					Target: rtc.Target_SUBSCRIBER,
					Type:   "answer",
					Sdp:    sdp.SDP,
				},
			},
		},
	)
	r.Unlock()
	if err != nil {
		log.Errorf("[%v] err=%v", r.uid, err)
		return err
	}
	return nil
}

func (r *RTC) SendAnswer2(sdp webrtc.SessionDescription, to string) error {
	log.Infof("[C=>S] [%v] to [%v] sdp=%v", r.uid, to, sdp)
	r.Lock()
	err := r.signaller.Send(
		&rtc.Request{
			Payload: &rtc.Request_Description{
				Description: &rtc.SessionDescription{
					Target: rtc.Target_SUBSCRIBER,
					Type:   "answer",
					Sdp:    sdp.SDP,
					To:     to,
				},
			},
		},
	)
	r.Unlock()
	if err != nil {
		log.Errorf("[%v] err=%v", r.uid, err)
		return err
	}
	return nil
}

func (r *RTC) SendWantConnectReply(sdp webrtc.SessionDescription, to string) error {
	log.Infof("[C=>S] [%v] to [%v] sdp=%v", r.uid, to, sdp)
	r.Lock()
	err := r.signaller.Send(
		&rtc.Request{
			Payload: &rtc.Request_WantConnectReply{
				WantConnectReply: &rtc.WantConnectReply{
					Sdp:       sdp.SDP,
					SdpType:   "offer",
					To:        to,
					IdleOrNot: true,
					Success:   true,
				},
			},
		},
	)
	r.Unlock()
	if err != nil {
		log.Errorf("[%v] err=%v", r.uid, err)
		return err
	}
	return nil
}

// Subscribe to tracks
func (r *RTC) Subscribe(trackInfos []*Subscription) error {
	if len(trackInfos) == 0 {
		return errors.New("track id is empty")
	}
	var infos []*rtc.Subscription
	for _, t := range trackInfos {
		infos = append(infos, &rtc.Subscription{
			TrackId:   t.TrackId,
			Mute:      t.Mute,
			Subscribe: t.Subscribe,
			Layer:     t.Layer,
		})
	}

	log.Infof("[C=>S] infos: %v", infos)
	err := r.signaller.Send(
		&rtc.Request{
			Payload: &rtc.Request_Subscription{
				Subscription: &rtc.SubscriptionRequest{
					Subscriptions: infos,
				},
			},
		},
	)
	r.sub.pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		log.Debugf("[S=>C] id=%v [r.sub.pc.OnDataChannel] got dc %v", r.uid, dc.Label())
		if dc.Label() == API_CHANNEL {
			log.Debugf("%v got dc %v", r.uid, dc.Label())
			r.sub.api = dc
			// send cmd after open
			r.sub.api.OnOpen(func() {
				log.Debugf("r.sub.api.OnOpen,state:%v,%v dc %v", r.sub.api.ReadyState(), r.uid, dc.Label())
				r.sub.api.SendText("wuyaxiong")
				if len(r.apiQueue) > 0 {
					for _, cmd := range r.apiQueue {
						log.Debugf("%v r.sub.api.OnOpen send cmd=%v", r.uid, cmd)
						marshalled, err := json.Marshal(cmd)
						if err != nil {
							continue
						}
						err = r.sub.api.Send(marshalled)
						if err != nil {
							log.Errorf("id=%v err=%v", r.uid, err)
						}
						time.Sleep(time.Millisecond * 10)
					}
					r.apiQueue = []Call{}
				}
			})
			return
		}
		log.Debugf("%v got dc %v", r.uid, dc.Label())
		if r.OnDataChannel != nil {
			r.OnDataChannel(dc)
		}
	})

	return err
}
func saveToDisk(i media.Writer, track *webrtc.TrackRemote) {
	defer func() {
		if err := i.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		packet, _, err := track.ReadRTP()
		if err != nil {
			panic(err)
		}

		if err := i.WriteRTP(packet); err != nil {
			panic(err)
		}
	}
}

// Creata new session
func (r *RTC) RegisterNewVideoSource(uid, name string, sourceType rtc.SourceType, config ...*JoinConfig) error {
	r.uid = uid
	log.Infof("[C=>S] sid: %v,uid:%v", uid, name)
	err := r.signaller.Send(
		&rtc.Request{
			Payload: &rtc.Request_Register{
				Register: &rtc.RegisterRequest{
					//Sid: sid,
					Uid:  uid,
					Name: name,
					//Role: rtc.Role_Admin,
					SourceType: sourceType,
				},
			},
		},
	)
	return err
}

// SubscribeFromEvent will parse event and subscribe what you want
func (r *RTC) SubscribeFromEvent(event TrackEvent, audio, video bool, layer string) error {
	log.Infof("event=%+v audio=%v video=%v layer=%v", event, audio, video, layer)
	if event.State == TrackEvent_UPDATE {
		return nil
	}

	var sub bool
	if event.State == TrackEvent_ADD {
		sub = true
	}

	var infos []*Subscription
	for _, t := range event.Tracks {
		// sub audio or not
		if audio && t.Kind == "audio" {
			infos = append(infos, &Subscription{
				TrackId:   t.Id,
				Mute:      t.Muted,
				Subscribe: sub,
				Layer:     t.Layer,
			})
			continue
		}
		// sub one layer
		if layer != "" && t.Kind == "video" && t.Layer == layer {
			infos = append(infos, &Subscription{
				TrackId:   t.Id,
				Mute:      t.Muted,
				Subscribe: sub,
				Layer:     t.Layer,
			})
			continue
		}
		// sub all if not set simulcast
		if t.Kind == "video" && layer == "" {
			infos = append(infos, &Subscription{
				TrackId:   t.Id,
				Mute:      t.Muted,
				Subscribe: sub,
				Layer:     t.Layer,
			})
		}
	}
	// sub video if publisher event not setting simulcast layer
	if len(infos) == 1 {
		for _, t := range event.Tracks {
			if t.Kind == "video" {
				infos = append(infos, &Subscription{
					TrackId:   t.Id,
					Mute:      t.Muted,
					Subscribe: sub,
					Layer:     t.Layer,
				})
			}
		}
	}
	for _, i := range infos {
		log.Infof("Subscribe/UnSubscribe infos=%+v", i)
	}
	return r.Subscribe(infos)
}

// Close client close
func (r *RTC) Close() {
	log.Infof("id=%v", r.uid)
	close(r.notify)
	if r.pub != nil {
		r.pub.pc.Close()
	}
	if r.sub != nil {
		r.sub.pc.Close()
	}
	r.cancel()
}

func (r *RTC) onSingalHandle() error {
	for {
		//only one goroutine for recving from stream, no need to lock
		stream, err := r.signaller.Recv()
		if err != nil {
			if err == io.EOF {
				log.Infof("[%v] WebRTC Transport Closed", r.uid)
				if err := r.signaller.CloseSend(); err != nil {
					log.Errorf("[%v] error sending close: %s", r.uid, err)
				}
				return err
			}

			errStatus, _ := status.FromError(err)
			if errStatus.Code() == codes.Canceled {
				if err := r.signaller.CloseSend(); err != nil {
					log.Errorf("[%v] error sending close: %s", r.uid, err)
				}
				return err
			}

			log.Errorf("[%v] Error receiving RTC response: %v", r.uid, err)
			if r.OnError != nil {
				r.OnError(err)
			}
			return err
		}

		switch payload := stream.Payload.(type) {
		case *rtc.Reply_Register:
			//处理RegisterReply的结果，如重名等
		//case *rtc.Request_Join:
		//网页或APP的Request_join可直接发到这里来
		case *rtc.Reply_Join:
			success := payload.Join.Success
			err := errors.New(payload.Join.Error.String())

			if !success {
				log.Errorf("[%v] [join] failed error: %v", r.uid, err)
				return err
			}
			log.Infof("[%v] [join] success", r.uid)
			log.Infof("payload.Reply.Description=%v", payload.Join.Description)
			sdp := webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  payload.Join.Description.Sdp,
			}

			if err = r.setRemoteSDP(sdp); err != nil {
				log.Errorf("[%v] [join] error %s", r.uid, err)
				return err
			}
		case *rtc.Reply_Description:
			var sdpType webrtc.SDPType
			if payload.Description.Type == "offer" {
				sdpType = webrtc.SDPTypeOffer
			} else {
				sdpType = webrtc.SDPTypeAnswer
			}
			sdp := webrtc.SessionDescription{
				SDP:  payload.Description.Sdp,
				Type: sdpType,
			}
			if sdp.Type == webrtc.SDPTypeOffer {
				log.Infof("[%v] [description] got offer call s.OnNegotiate sdp=%+v", r.uid, sdp)
				err := r.negotiate(sdp)
				if err != nil {
					log.Errorf("error: %v", err)
				}
			} else if sdp.Type == webrtc.SDPTypeAnswer {
				log.Infof("[%v] [description] got answer call sdp=%+v", r.uid, sdp)
				//err = r.setRemoteSDP(sdp)
				err = r.setClientRemoteSDP(sdp, payload.Description.From)
				if err != nil {
					log.Errorf("[%v] [description] setRemoteSDP err=%s", r.uid, err)
				}
			}
		case *rtc.Reply_WantConnectRequest:
			log.Infof("rtc.Reply_WantConnect :%v,type:%v", payload.WantConnectRequest, payload.WantConnectRequest.ConnectType)
			// if payload.WantConnect.SdpType == "offer" {
			// 	sdp := webrtc.SessionDescription{
			// 		SDP:  payload.WantConnect.Sdp,
			// 		Type: webrtc.SDPTypeOffer,
			// 	}
			//log.Infof("wantConnect from [%v] ", payload.WantConnectRequest.From)
			err := r.getWantConnectRequest(payload.WantConnectRequest.From, payload.WantConnectRequest.ConnectType)
			if err != nil {
				log.Errorf("error: %v", err)
			}
			//}

		case *rtc.Reply_Trickle:
			var candidate webrtc.ICECandidateInit
			_ = json.Unmarshal([]byte(payload.Trickle.Init), &candidate)
			//log.Infof("[%v] [trickle] from=%v to=%v candidate=%v", r.uid, payload.Trickle.From, payload.Trickle.To, candidate)
			r.receiveTrickle2(candidate, payload.Trickle.From)
		case *rtc.Reply_TrackEvent:
			if r.OnTrackEvent == nil {
				log.Errorf("s.OnTrackEvent == nil")
				continue
			}
			var TrackInfos []*TrackInfo
			for _, v := range payload.TrackEvent.Tracks {
				TrackInfos = append(TrackInfos, &TrackInfo{
					Id:        v.Id,
					Kind:      v.Kind,
					Muted:     v.Muted,
					Type:      MediaType(v.Type),
					StreamId:  v.StreamId,
					Label:     v.Label,
					Width:     v.Width,
					Height:    v.Height,
					FrameRate: v.FrameRate,
					Layer:     v.Layer,
				})
			}
			trackEvent := TrackEvent{
				State:  TrackEvent_State(payload.TrackEvent.State),
				Uid:    payload.TrackEvent.Uid,
				Tracks: TrackInfos,
			}

			log.Infof("s.OnTrackEvent trackEvent=%+v", trackEvent)
			r.OnTrackEvent(trackEvent)
		case *rtc.Reply_Subscription:
			if !payload.Subscription.Success {
				log.Errorf("suscription error: %v", payload.Subscription.Error)
			}
		case *rtc.Reply_Error:
			log.Errorf("Request error: %v", payload.Error)
		default:
			log.Errorf("Unknown RTC type!!!!%v", payload)
		}
	}
}

func (r *RTC) refuseWantConnect(refuse *rtc.WantConnectReply) error {
	r.Lock()
	//如果有一个client的Role是controler，就回应忙，并附带时间信息
	err := r.signaller.Send(
		&rtc.Request{
			Payload: &rtc.Request_WantConnectReply{
				WantConnectReply: refuse,
			},
		},
	)
	if err != nil {
		log.Errorf("[%v] err=%v", r.uid, err)
		return err
	}
	r.Unlock()
	return nil

}

// func (r *RTC) getWantConnectRequest(uid string, sdp webrtc.SessionDescription) error {
func (r *RTC) getWantConnectRequest(uid string, connectType rtc.ConnectType) error {
	log.Debugf("getWantConnectRequest from %v ,type:%v", uid, connectType)
	/*如果有一个client的Role是controler，就回应忙，并附带时间信息
	  如果Role是view,需要统计个数，超过的回应忙
	*/
	var client *Client
	if connectType == rtc.ConnectType_Control {
		//如果有一个client的Role是controler，就回应忙，并附带时间信息
		for _, c := range r.clients {
			if c.ConnectType == rtc.ConnectType_Control {
				return r.refuseWantConnect(&rtc.WantConnectReply{
					To:           uid,
					Success:      true,
					IdleOrNot:    false,
					RestTimeSecs: 100,
					NumOfWaiting: 1,
				})
			}
		}
		client = NewClient(uid, r, rtc.ConnectType_Control)
		client.DataChannelEable = true
		r.clients = append(r.clients, client)
	} else if connectType == rtc.ConnectType_View {
		//如果Role是view,需要统计个数，超过的回应忙
		log.Debugf("len(r.clients):%d", len(r.clients))
		for _, c := range r.clients {
			log.Debugf("client: %v", c.Id)
		}
		num := len(r.clients)
		if num >= r.MaxClientsNum { //这里没有考虑control个数，一般为1个，即有1个的误差
			return r.refuseWantConnect(&rtc.WantConnectReply{
				To:           uid,
				Success:      true,
				IdleOrNot:    false,
				RestTimeSecs: 100,
				NumOfWaiting: uint32(num),
			})
		}
		client = NewClient(uid, r, rtc.ConnectType_View)
		client.DataChannelEable = false
		r.clients = append(r.clients, client)
	}

	if _, err := client.pubPc.AddTrack(r.VedioTrack); err != nil {
		log.Errorf("AddTrack error: %v", err)
	}
	if _, err := client.pubPc.AddTrack(r.AudioTrack); err != nil {
		log.Errorf("AddTrack error: %v", err)
	}

	client.dataChannel, _ = client.pubPc.CreateDataChannel(API_CHANNEL, &webrtc.DataChannelInit{})

	client.pubPc.OnDataChannel(func(dc *webrtc.DataChannel) {
		client.dataChannel = dc
		log.Errorf("test")
		if r.OnDataChannel != nil {
			r.OnDataChannel(dc)
		}
	})
	log.Errorf("test")
	client.subPc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		r.OnTrack(track, receiver)
	})

	client.dataChannel.OnOpen(func() {
		log.Debugf("client data channel opened")
		//client.dataChannel.SendText("wuyaxiong nb")
	})

	client.dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		//datachannel always receive message,but only controler can run r.OnDataChannelMessage
		//recvData := make(map[string]interface{})
		var dataChannelMsg DataChannelMsg
		log.Debugf("received msg.Data:%s", msg.Data)
		err := json.Unmarshal(msg.Data, &dataChannelMsg)
		if err != nil {
			log.Errorf("Unmarshal:err %v", err)
			return
		}
		log.Debugf("received message:%v,data type:%T,data:%v", dataChannelMsg.Cmd, dataChannelMsg.Data, dataChannelMsg.Data)
		switch dataChannelMsg.Cmd {
		case "offer":
			var offer webrtc.SessionDescription
			var sdpType webrtc.SDPType
			var recvSdp DataChannelMsgDataSdp
			sdpString, ok := dataChannelMsg.Data.(string)
			if ok {
				err = json.Unmarshal([]byte(sdpString), &recvSdp)

				if err != nil {
					log.Errorf("Unmarshal sdp:err %v", err)
					return
				}
			}
			if recvSdp.Type == "offer" {
				sdpType = webrtc.SDPTypeOffer
			} else {
				sdpType = webrtc.SDPTypeAnswer
			}
			offer = webrtc.SessionDescription{
				Type: sdpType,
				SDP:  recvSdp.Sdp,
			}
			err := client.subPc.SetRemoteDescription(offer)
			if err != nil {
				log.Errorf("id=%v Negotiate r.sub.pc.SetRemoteDescription err=%v", r.uid, err)
				break
			}

			// 2. safe to send candiate to sfu after join ok
			if len(client.subSendCandidates) > 0 {
				for _, cand := range client.subSendCandidates {
					log.Debugf("[C=>S] id=%v send sub.SendCandidates r.uid, r.rtc.trickle cand=%v", r.uid, cand)
					//send from data channel
					//r.SendTrickle(cand, Target_SUBSCRIBER)
					candJson, err := json.Marshal(&DataChannelMsg{
						Cmd:  "candi",
						Data: cand.ToJSON(),
					})
					if err != nil {
						log.Errorf("json.Marshal err=%v", err)
						break
					}
					client.dataChannel.SendText(string(candJson))
				}
				client.subSendCandidates = []*webrtc.ICECandidate{}
			}

			// 3. safe to add candidate after SetRemoteDescription
			if len(client.subRecvCandidates) > 0 {
				for _, candidate := range client.subRecvCandidates {
					log.Debugf("id=%v r.sub.pc.AddICECandidate candidate=%v", r.uid, candidate)
					_ = client.subPc.AddICECandidate(candidate)
				}
				client.subRecvCandidates = []webrtc.ICECandidateInit{}
			}

			// 4. create answer after add ice candidate
			answer, err := client.subPc.CreateAnswer(nil)
			if err != nil {
				log.Errorf("id=%v err=%v", r.uid, err)
				return
			}

			// 5. set local sdp(answer)
			err = client.subPc.SetLocalDescription(answer)
			if err != nil {
				log.Errorf("id=%v err=%v", r.uid, err)
				return
			}

			// 6. send answer to sfu
			//err = r.SendAnswer(answer)
			//send from data channel
			answerData := DataChannelMsgDataSdp{
				Sdp:  answer.SDP,
				Type: "answer",
			}
			answerDataJson, err := json.Marshal(&answerData)
			answerJson, err := json.Marshal(&DataChannelMsg{
				Cmd:  "answer",
				Data: answerDataJson,
			})
			if err != nil {
				log.Errorf("json.Marshal err=%v", err)
				break
			}
			log.Infof("send anwer:%v", answer)
			client.dataChannel.SendText(string(answerJson))
			if err != nil {
				log.Errorf("id=%v err=%v", r.uid, err)
				return
			}

		case "candi":
			candiString, ok := dataChannelMsg.Data.(string)
			var resvCandidate webrtc.ICECandidateInit
			if ok {
				err = json.Unmarshal([]byte(candiString), &resvCandidate)

				if err != nil {
					log.Errorf("Unmarshal sdp:err %v", err)
					return
				}
			}

			if client.subPc.CurrentRemoteDescription() == nil {
				client.subRecvCandidates = append(client.subRecvCandidates, resvCandidate)
				log.Debugf("resvCandidate:%v", resvCandidate)
			} else {
				err := client.subPc.AddICECandidate(resvCandidate)
				log.Debugf("resvCandidate:%v", resvCandidate)
				if err != nil {
					log.Errorf("id=%v err=%v", r.uid, err)
				}
			}
		case "speed":
			recvSpeed, ok := dataChannelMsg.Data.(float64)
			if ok && client.DataChannelEable && r.ControlFunc != nil {
				r.ControlFunc("speed", recvSpeed)
				//log.Debugf("recvSpeed,%v", recvSpeed)
			}
		case "turn":
			recvTurn, ok := dataChannelMsg.Data.(float64)
			if ok && client.DataChannelEable && r.ControlFunc != nil {
				r.ControlFunc("turn", recvTurn)
				//log.Debugf("recvTurn,%v", recvTurn)
			}
		case "control":
			recvControl, ok := dataChannelMsg.Data.(float64)
			log.Infof("recvControl,%v,%v", recvControl, ok)
			if ok {
				if recvControl == 1 {
					hasController := false
					var controlRes string
					//如果有一个client的Role是controler，就回应忙，并附带时间信息
					for _, c := range r.clients {
						if c.ConnectType == rtc.ConnectType_Control {
							hasController = true
						}
					}
					if !hasController {
						client.ConnectType = rtc.ConnectType_Control
						client.DataChannelEable = true
						controlRes = "ok"
					} else {
						client.ConnectType = rtc.ConnectType_View
						client.DataChannelEable = false
						controlRes = "终端已有人正在控制，请稍后再试"
					}

					respJson, err := json.Marshal(&DataChannelMsg{
						Cmd:  "controlRes",
						Data: controlRes,
					})
					if err != nil {
						log.Errorf("json.Marshal err=%v", err)
					} else {
						log.Infof("c.subPc.OnICECandidate")
						client.dataChannel.SendText(string(respJson))
					}

				}
				log.Debugf("recvControl,%v", recvControl)
			}
		default:
		}
		// _, have_d := recvData["control"]
		// if have_d {
		// 	if recvData["control"] == 1 { //want control
		// 		client.dataChannel.SendText("idle")
		// 		//如果有一个client的Role是controler，return ,else send message to pi from datachannel
		// 		for _, c := range r.clients {
		// 			if c.ConnectType == rtc.ConnectType_Control {
		// 				client.dataChannel.SendText("busy-controled")
		// 			} //if c.ConnectType ==
		// 		} //for
		// 		client.DataChannelEable = true
		// 	} //if(recvData["control"] ==1)
		// } //if have_d {

		if client.DataChannelEable == true {
			//log.Debugf("client data channel messages:%s", msg)
			if r.OnDataChannelMessage != nil {
				r.OnDataChannelMessage(msg)
			}
		}
	})

	offer, err := client.pubPc.CreateOffer(nil)
	//log.Infof("offer: %v", offer)
	if err != nil {
		log.Errorf("id=%v err=%v", r.uid, err)
	}

	// 2. pub set local sdp(offer)
	err = client.pubPc.SetLocalDescription(offer)
	if err != nil {
		log.Errorf("id=%v err=%v", r.uid, err)
	}

	if len(client.pubSendCandidates) > 0 {
		for _, cand := range client.pubSendCandidates {
			log.Debugf("[C=>S] id=%v send sub.SendCandidates r.uid, r.rtc.trickle cand=%v", uid, cand)
			r.SendTrickle2(cand, uid)
		}
		client.pubSendCandidates = []*webrtc.ICECandidate{}
	}

	// 3. safe to add candidate after SetRemoteDescription
	if len(client.pubRecvCandidates) > 0 {
		for _, candidate := range client.pubRecvCandidates {
			log.Debugf("uid=%v r.sub.pc.AddICECandidate candidate=%v", uid, candidate)
			_ = client.pubPc.AddICECandidate(candidate)
		}
		client.pubRecvCandidates = []webrtc.ICECandidateInit{}
	}

	err = r.SendWantConnectReply(offer, uid)
	if err != nil {
		log.Errorf("id=%v err=%v", r.uid, err)
		return err
	}
	return err
}
