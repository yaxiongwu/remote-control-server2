package engine

import (
	"encoding/json"
	"time"

	"github.com/pion/ice/v2"
	log "github.com/pion/ion-log"
	"github.com/pion/webrtc/v3"
	rtclib "github.com/yaxiongwu/remote-control-client-go2/pkg/proto/rtc"
)

// Client is pub/sub transport
type Client struct {
	Id          string
	dataChannel *webrtc.DataChannel
	rtc         *RTC
	//pc          *webrtc.PeerConnection

	pubPc                         *webrtc.PeerConnection
	pubSendCandidates             []*webrtc.ICECandidate
	pubRecvCandidates             []webrtc.ICECandidateInit
	pubOnIceConnectionStateChange func(webrtc.ICEConnectionState, *webrtc.PeerConnection)

	subPc                         *webrtc.PeerConnection
	subSendCandidates             []*webrtc.ICECandidate
	subRecvCandidates             []webrtc.ICECandidateInit
	subOnIceConnectionStateChange func(webrtc.ICEConnectionState, *webrtc.PeerConnection)

	StartControlTime time.Time
	StartViewTime    time.Time
	ControlTimer     *time.Timer
	//role           Target
	//Role                       rtc.Role
	ConnectType rtclib.ConnectType

	config                     *RTCConfig
	OnIceConnectionStateChange func(webrtc.ICEConnectionState, *webrtc.PeerConnection)
	DataChannelEable           bool
}

// NewTransport create a transport
func NewClient(uid string, rtc *RTC, connectType rtclib.ConnectType) *Client {
	c := &Client{
		Id:          uid,
		config:      &DefaultConfig,
		rtc:         rtc,
		ConnectType: connectType,
	}
	if rtc.config == nil {
		rtc.config = &DefaultConfig
	}
	// c.pub = NewTransport(uid, Target_PUBLISHER, rtc)
	// c.sub = NewTransport(uid, Target_PUBLISHER, rtc)
	var api *webrtc.API
	var me *webrtc.MediaEngine
	var err error
	c.config.WebRTC.Setting.SetICEMulticastDNSMode(ice.MulticastDNSModeDisabled)

	me, err = getPublisherMediaEngine(rtc.config.WebRTC.VideoMime)
	if err != nil {
		log.Errorf("getPublisherMediaEngine error: %v", err)
		return nil
	}

	api = webrtc.NewAPI(webrtc.WithMediaEngine(me), webrtc.WithSettingEngine(rtc.config.WebRTC.Setting))

	c.pubPc, err = api.NewPeerConnection(rtc.config.WebRTC.Configuration)

	me, err = getSubscriberMediaEngine()
	if err != nil {
		log.Errorf("getPublisherMediaEngine error: %v", err)
		return nil
	}
	api = webrtc.NewAPI(webrtc.WithMediaEngine(me), webrtc.WithSettingEngine(rtc.config.WebRTC.Setting))
	c.subPc, err = api.NewPeerConnection(rtc.config.WebRTC.Configuration)

	c.pubSendCandidates = []*webrtc.ICECandidate{}
	c.subSendCandidates = []*webrtc.ICECandidate{}

	// var err error
	// var api *webrtc.API
	// var me *webrtc.MediaEngine
	// c.config.WebRTC.Setting.SetICEMulticastDNSMode(ice.MulticastDNSModeDisabled)
	// // if role == Target_PUBLISHER {
	// // 	me, err = getPublisherMediaEngine(rtc.config.WebRTC.VideoMime)
	// // } else {
	// me, err = getSubscriberMediaEngine()
	// //}

	// if err != nil {
	// 	log.Errorf("getPublisherMediaEngine error: %v", err)
	// 	return nil
	// }

	// api = webrtc.NewAPI(webrtc.WithMediaEngine(me), webrtc.WithSettingEngine(c.config.WebRTC.Setting))
	// c.pc, err = api.NewPeerConnection(c.config.WebRTC.Configuration)

	// if err != nil {
	// 	log.Errorf("NewPeerConnection error: %v", err)
	// 	return nil
	// }

	// if role == Target_PUBLISHER {
	// 	log.Debugf("t.pc.CreateDataChannel(API_CHANNEL)")
	// 	_, err = t.pc.CreateDataChannel(API_CHANNEL, &webrtc.DataChannelInit{})

	// 	if err != nil {
	// 		log.Errorf("error creating data channel: %v", err)
	// 		return nil
	// 	}
	// }
	//pub从grpc走，sub从datachannel走
	c.pubPc.OnICECandidate(func(i *webrtc.ICECandidate) {
		//	log.Debugf("t.pc.OnICECandidate,myid:%v,%v", uid, i)
		if i == nil {
			// Gathering done
			log.Infof("gather candidate done")
			return
		}
		//append before join session success
		if c.pubPc.CurrentRemoteDescription() == nil {
			c.pubSendCandidates = append(c.pubSendCandidates, i)
		} else {
			//for _, cand := range c.pubSendCandidates {
			//c.rtc.SendTrickle2(cand, uid)

			//}
			c.pubSendCandidates = []*webrtc.ICECandidate{}
			c.rtc.SendTrickle2(i, uid)
		}
	})
	//pub从grpc走，sub从datachannel走
	c.subPc.OnICECandidate(func(i *webrtc.ICECandidate) {
		//	log.Debugf("subPc.OnICECandidate:%v", i)
		if i == nil {
			// Gathering done
			log.Infof("gather candidate done")
			return
		}
		//append before join session success
		if c.subPc.CurrentRemoteDescription() == nil {
			c.subSendCandidates = append(c.subSendCandidates, i)
			///log.Infof("c.subPc.CurrentRemoteDescription() == nil")
		} else {
			//log.Infof("c.subPc.CurrentRemoteDescription()  else")
			for _, cand := range c.subSendCandidates {
				//c.rtc.SendTrickle2(cand, uid)

				candJson, err := json.Marshal(&DataChannelMsg{
					Cmd:  "candi",
					Data: cand.ToJSON(),
				})
				if err != nil {
					log.Errorf("json.Marshal err=%v", err)
					break
				}
				//	log.Infof("c.subPc.OnICECandidate")
				c.dataChannel.SendText(string(candJson))

			}
			c.subSendCandidates = []*webrtc.ICECandidate{}
			candJson, err := json.Marshal(&DataChannelMsg{
				Cmd:  "candi",
				Data: i.ToJSON(),
			})
			if err != nil {
				log.Errorf("json.Marshal err=%v", err)
			} else {
				//log.Infof("c.subPc.OnICECandidate")
				c.dataChannel.SendText(string(candJson))
			}
		}
	})

	c.subPc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Debugf("IECConnectionStateChange to %v", state)
	})

	c.pubPc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		if rtc.OnIceConnectionStateChange != nil {
			rtc.OnIceConnectionStateChange(state, c.pubPc)
		}
		log.Debugf("IECConnectionStateChange to %v", state)
		if state == webrtc.ICEConnectionStateDisconnected || state == webrtc.ICEConnectionStateFailed || state == webrtc.ICEConnectionStateClosed {
			for i, client := range rtc.clients {
				if client.Id == c.Id {
					rtc.clients = append(rtc.clients[:i], rtc.clients[i+1:]...)
					break
				}
			}
		}

		//记录连接上的时间
		if state == webrtc.ICEConnectionStateConnected {
			c.StartControlTime = time.Now()
			c.StartViewTime = time.Now()
			if c.ConnectType == rtclib.ConnectType_Control {
				c.ControlTimer = time.NewTimer(time.Duration(rtc.MaxTimeControl) * time.Second)
			} else if c.ConnectType == rtclib.ConnectType_View {
				c.ControlTimer = time.NewTimer(time.Duration(rtc.MaxTimeView) * time.Second)
			}
			go func() {
				for {
					select {
					case <-c.ControlTimer.C:
						c.subPc.Close()
					}
				}
			}()
			log.Debugf("startTime: %v", c.StartControlTime)
		}
	})

	return c
}

// func (t *Client) GetPeerConnection() *webrtc.PeerConnection {
// 	return t.pc
// }
