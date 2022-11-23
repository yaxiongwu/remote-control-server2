package server

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	log "github.com/pion/ion-log"
	"github.com/pion/webrtc/v3"
	"github.com/yaxiongwu/remote-control-server2/pkg/stun"

	//rtc "github.com/pion/ion/proto/rtc"
	rtc "github.com/yaxiongwu/remote-control-server2/pkg/proto/rtc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Code int32

const (
	Ok                     Code = 200
	BadRequest             Code = 400
	Forbidden              Code = 403
	NotFound               Code = 404
	RequestTimeout         Code = 408
	UnsupportedMediaType   Code = 415
	BusyHere               Code = 486
	TemporarilyUnavailable Code = 480
	InternalError          Code = 500
	NotImplemented         Code = 501
	ServiceUnavailable     Code = 503
)

type STUNServer struct {
	rtc.UnimplementedRTCServer
	sync.Mutex
	//SFU  *sfu.SFU
	STUN *stun.STUN
	sigs map[string]rtc.RTC_SignalServer
}

func NewServer(stun *stun.STUN) *STUNServer {
	return &STUNServer{STUN: stun}
}

// func NewSFUServer(sfu *sfu.SFU) *SFUServer {
// 	return &SFUServer{
// 		SFU:  sfu,
// 		sigs: make(map[string]rtc.RTC_SignalServer),
// 	}
// }

func (s *STUNServer) BroadcastTrackEvent(uid string, tracks []*rtc.TrackInfo, state rtc.TrackEvent_State) {

	s.Lock()
	defer s.Unlock()
	for id, sig := range s.sigs {
		if id == uid {
			continue
		}
		err := sig.Send(&rtc.Reply{
			Payload: &rtc.Reply_TrackEvent{
				TrackEvent: &rtc.TrackEvent{
					Uid:    uid,
					Tracks: tracks,
					State:  state,
				},
			},
		})
		if err != nil {
			log.Errorf("signal send error: %v", err)
		}
	}
}

func (s *STUNServer) Signal(sig rtc.RTC_SignalServer) error {

	// var tracksMutex sync.RWMutex
	// var tracksInfo []*rtc.TrackInfo

	client := stun.NewClient(s.STUN)
	defer client.Close()
	for {

		in, err := sig.Recv()

		if err != nil {

			if err == io.EOF {
				return nil
			}

			errStatus, _ := status.FromError(err)
			if errStatus.Code() == codes.Canceled {
				return nil
			}

			log.Errorf("%v signal error %d", fmt.Errorf(errStatus.Message()), errStatus.Code())
			return err
		}
		//log.Infof("in.Payload.(type):%v\r\n", in.Payload)
		// bVideoSource := true

		rtcTarget := rtc.Target_SUBSCRIBER
		switch payload := in.Payload.(type) {
		case *rtc.Request_Register:
			name := payload.Register.Name
			uid := payload.Register.Uid
			sourceType := payload.Register.SourceType
			log.Infof("[C=>S] createSession: name => %v, uid => %v", name, uid, sourceType)
			//需要查找是否有重名
			err = client.CreateSession(uid, sourceType)

			if err != nil {
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_Register{
						Register: &rtc.RegisterReply{
							Error: &rtc.Error{
								Code:   int32(1),
								Reason: fmt.Sprintf("create seesion error: %v", err),
							},
						},
					},
				})
				if err != nil {
					log.Errorf("create seesion error: %v", err)
				}
				break
			}

			//client.Join(sid, uid)
			client.Register(uid, name)
			log.Debugf("client.GetID():%v", client.GetID())

			rtcTarget = rtc.Target_PUBLISHER

			client.OnSessionDescription = func(o *webrtc.SessionDescription, from string, to string) {
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_Description{
						Description: &rtc.SessionDescription{
							From:   from,
							To:     to,
							Target: rtc.Target(rtcTarget), //需要特别注意SUB和PUB，视频源应该是pub，控制端主动发起，是Sub
							Sdp:    o.SDP,
							Type:   o.Type.String(),
						},
					},
				})
				if err != nil {
					log.Errorf("negotiation error: %v", err)
				}
			}

			client.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, from string, to string) {
				log.Debugf("[S=>C] peer.OnIceCandidate:from= %v, to= %v,candidate = %v", from, to, candidate.Candidate)
				bytes, err := json.Marshal(candidate)
				if err != nil {
					log.Errorf("OnIceCandidate error: %v", err)
				}
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_Trickle{
						Trickle: &rtc.Trickle{
							Init: string(bytes),
							From: from,
							To:   to,
						},
					},
				})
				if err != nil {
					log.Errorf("OnIceCandidate send error: %v", err)
				}
			}

			client.OnJoinReply = func(o *webrtc.SessionDescription) {
				log.Debugf("[S=>C] client.OnJoinReply: %v", o.SDP)
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_Join{
						Join: &rtc.JoinReply{
							Success: true,
							Error:   nil,
							Description: &rtc.SessionDescription{
								Target: rtc.Target(rtcTarget),
								Sdp:    o.SDP,
								Type:   o.Type.String(),
							},
						},
					},
				})
			}
			//服务器需要把web端发过来的wantconnectRequest以reply的方式发给视频源
			client.OnWantConnectRequestReply = func(o *rtc.WantConnectRequest) {
				log.Debugf("[S=>C] client.OnWantConnectReply: %v", o)
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_WantConnectRequest{
						WantConnectRequest: o,
					},
				})
			}

			//log.Debugf("client.GetID():%v,client.OnJoinReply: %v", client.GetID(), client.OnJoinReply)
		case *rtc.Request_OnLineSource:

			sessions := s.STUN.GetSessions()

			log.Debugf("[S=>C] client.OnLineSource: %v,sessions:%v", payload.OnLineSource.SourceType, sessions)
			var onlineSources []*rtc.OnLineSources

			for _, s := range sessions {
				log.Debugf("Sid:%v,Uid:%v", s.ID(), client.GetID())
				if s.GetSourceType() == payload.OnLineSource.SourceType {
					onlineSources = append(onlineSources, &rtc.OnLineSources{
						//Sid: s.ID(),
						Uid:  s.GetSourceClient().GetID(),
						Name: s.GetSourceClient().GetName(),
					})
				}
			}
			// onlineSources = append(onlineSources, &rtc.OnLineSources{
			// 	Sid: "testSid",
			// 	Uid: "testUid",241
			// })
			log.Debugf("onlineSources:%v", onlineSources)
			err = sig.Send(&rtc.Reply{
				Payload: &rtc.Reply_OnLineSource{
					OnLineSource: &rtc.OnLineSourceReply{
						Success:       true,
						Error:         nil,
						OnLineSources: onlineSources,
					},
				},
			})
			if err != nil {
				log.Errorf("err:%v", err)
			}

		case *rtc.Request_WantConnect:

			from := payload.WantConnect.From
			to := payload.WantConnect.To
			//name := payload.WantConnect.Name
			log.Infof("[C=>S] Request_WantConnect:  from :%v,to:%v,type:%v", from, to, payload.WantConnect.ConnectType)
			//WantConnect只带了目的地址，比如网页上带了PiVedioSource，到了Pi那端，需要知道的是网页的id
			//对连接用户的管理交给视频源，这里不回应WantConnectReply，
			// WantConnectReply := client.WantConnect(from, to)

			// err = sig.Send(&rtc.Reply{
			// 	Payload: &rtc.Reply_WantConnect{
			// 		WantConnect: WantConnectReply,
			// 	},
			// })
			// if err != nil {
			// 	log.Errorf("err:%v", err)
			// }
			client.WantConnect(from, to)
			rtcTarget := rtc.Target_SUBSCRIBER
			client.OnSessionDescription = func(o *webrtc.SessionDescription, from string, to string) {
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_Description{
						Description: &rtc.SessionDescription{
							From:   from,
							To:     to,
							Target: rtc.Target(rtcTarget), //需要特别注意SUB和PUB，视频源应该是pub，控制端主动发起，是Sub
							Sdp:    o.SDP,
							Type:   o.Type.String(),
						},
					},
				})
				if err != nil {
					log.Errorf("negotiation error: %v", err)
				}
			}

			client.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, from string, to string) {
				log.Debugf("[S=>C] peer.OnIceCandidate:from= %v, to= %v,candidate = %v", from, to, candidate.Candidate)
				bytes, err := json.Marshal(candidate)
				if err != nil {
					log.Errorf("OnIceCandidate error: %v", err)
				}
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_Trickle{
						Trickle: &rtc.Trickle{
							Init: string(bytes),
							From: from,
							To:   to,
						},
					},
				})
				if err != nil {
					log.Errorf("OnIceCandidate send error: %v", err)
				}
			}

			client.OnWantConnectReply = func(o *rtc.WantConnectReply) {
				//log.Debugf("[S=>C] client.OnWantConnectReply: %v", o)
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_WantConnect{
						WantConnect: o,
					},
				})
			}

			sourceClient := client.Session().GetSourceClient()
			if sourceClient != nil {
				log.Debugf("WantConnect desc from client %v to client:%v,Uid:%v", client.GetID(), sourceClient.GetID(), payload.WantConnect.From)
				if sourceClient.OnWantConnectRequestReply != nil {
					// sourceClient.OnWantConnectRequestReply(&rtc.WantConnectReply{
					// 	Success:   true,
					// 	From:      payload.WantConnect.From,
					// 	Sdp:       payload.WantConnect.Sdp,
					// 	SdpType:   payload.WantConnect.SdpType,
					// 	ConnectTyp
					// })
					sourceClient.OnWantConnectRequestReply(payload.WantConnect)
				}
			}

			//网页向服务器发Request_WantConnect,服务器向视频源发WantConnectReply，视频源根据实际情况回复Request_WantConnectReply,服务器收到后，转发WantConnectReply给网页
		case *rtc.Request_WantConnectReply:

			from := payload.WantConnectReply.From
			to := payload.WantConnectReply.To
			//name := payload.WantConnect.Name
			log.Infof("[C=>S] Request_WantConnectReply:  from :%v,to:%v", from, to)
			//log.Debugf("Request_WantConnectReply:%v", payload.WantConnectReply)
			//clients := client.Session().Clients()

			c := client.Session().GetClient(payload.WantConnectReply.To)
			// for _, c := range clients {
			// 	if c.GetID() == payload.WantConnectReply.To {

			if c.OnWantConnectReply != nil && c != nil {

				// c.OnWantConnectReply(&rtc.WantConnectReply{
				// 	Success:      true,
				// 	From:         payload.WantConnectReply.From,
				// 	Sdp:          payload.WantConnectReply.Sdp,
				// 	SdpType:      payload.WantConnectReply.SdpType,
				// 	IdleOrNot:    payload.WantConnectReply.IdleOrNot,
				// 	RestTimeSecs: payload.WantConnectReply.RestTimeSecs,
				// 	NumOfWaiting: payload.WantConnectReply.NumOfWaiting,
				// })
				c.OnWantConnectReply(payload.WantConnectReply)
				// 	}
				// }
			}

		case *rtc.Request_Description:

			desc := webrtc.SessionDescription{
				SDP:  payload.Description.Sdp,
				Type: webrtc.NewSDPType(payload.Description.Type),
			}
			log.Debugf("description:%v,from:%v,to:%v", payload.Description.Sdp, client.GetID(), payload.Description.To)
			// clients := client.Session().Clients()
			// for _, c := range clients {
			//if c.GetID() == payload.Description.To {
			c := client.Session().GetClient(payload.Description.To)
			if c.OnSessionDescription != nil && c != nil {
				c.OnSessionDescription(&desc, client.GetID(), payload.Description.To)
			}
			// 	}
			// }

		case *rtc.Request_Trickle:
			var candidate webrtc.ICECandidateInit
			log.Infof("Trickle.Init [%v] to [%v]", payload.Trickle.Init, payload.Trickle.To)
			err := json.Unmarshal([]byte(payload.Trickle.Init), &candidate)
			if err != nil {
				log.Errorf("error parsing ice candidate, error -> %v", err)
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_Error{
						Error: &rtc.Error{
							Code:   int32(InternalError),
							Reason: fmt.Sprintf("unmarshal ice candidate error:  %v", err),
						},
					},
				})
				if err != nil {
					log.Errorf("grpc send error: %v", err)
					return status.Errorf(codes.Internal, err.Error())
				}
				continue
			}

			//clients := client.Session().Clients()
			// for _, c := range clients {
			// 	if c.GetID() == payload.Trickle.To {
			c := client.Session().GetClient(payload.Trickle.To)
			if c.OnIceCandidate != nil && c != nil {
				log.Debugf("client.GetID():%v,trickle from %v to %v,candidate:%v", client.GetID(), payload.Trickle.From, payload.Trickle.To, candidate.Candidate)
				if c.OnIceCandidate != nil {
					//pub和sub问题是相对的，具体问题还有待进一步优化
					//clients[c].OnIceCandidate(&candidate, int(payload.Trickle.Target))
					c.OnIceCandidate(&candidate, client.GetID(), payload.Trickle.To)
				}
				//}
			}

		case *rtc.Request_Subscription:
			// sid := payload.Subscription.Subscriptions[0].TrackId
			// uid := payload.Subscription.Subscriptions[0].Layer
			// log.Debugf("[C=>S] sid: %v,uid:%s", sid, uid) //, payload.Subscription.trackId, payload.Subscription.layer)

			// //client.CreateSession(sid)
			// //client.Join(sid, uid)
			// rtcTarget = rtc.Target_PUBLISHER

			// // client.OnSessionDescription = func(o *webrtc.SessionDescription) {
			// // 	log.Debugf("[S=>C] client.OnSessionDescription: %v", o.SDP)
			// // 	// if bVideoSource {
			// // 	// 	rtcTarget = rtc.Target_SUBSCRIBER
			// // 	// } else {
			// // 	// 	rtcTarget = rtc.Target_PUBLISHER
			// // 	// }
			// // 	err = sig.Send(&rtc.Reply{
			// // 		Payload: &rtc.Reply_Description{
			// // 			Description: &rtc.SessionDescription{
			// // 				Target: rtc.Target(rtcTarget), //需要特别注意SUB和PUB，视频源应该是pub，控制端主动发起，是Sub
			// // 				Sdp:    o.SDP,
			// // 				Type:   o.Type.String(),
			// // 			},
			// // 		},
			// // 	})
			// // 	if err != nil {
			// // 		log.Errorf("negotiation error: %v", err)
			// // 	}
			// // }

			// client.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, target int) {
			// 	log.Debugf("[S=>C] peer.OnIceCandidate: target = %v, candidate = %v", target, candidate.Candidate)
			// 	bytes, err := json.Marshal(candidate)
			// 	if err != nil {
			// 		log.Errorf("OnIceCandidate error: %v", err)
			// 	}
			// 	err = sig.Send(&rtc.Reply{
			// 		Payload: &rtc.Reply_Trickle{
			// 			Trickle: &rtc.Trickle{
			// 				Init:   string(bytes),
			// 				Target: rtc.Target(target),
			// 			},
			// 		},
			// 	})
			// 	if err != nil {
			// 		log.Errorf("OnIceCandidate send error: %v", err)
			// 	}
			// }

			// client.OnJoinReply = func(o *webrtc.SessionDescription) {
			// 	log.Debugf("[S=>C] client.OnJoinReply: %v", o.SDP)
			// 	err = sig.Send(&rtc.Reply{
			// 		Payload: &rtc.Reply_Join{
			// 			Join: &rtc.JoinReply{
			// 				Success: true,
			// 				Error:   nil,
			// 				Description: &rtc.SessionDescription{
			// 					Target: rtc.Target(rtcTarget),
			// 					Sdp:    o.SDP,
			// 					Type:   o.Type.String(),
			// 				},
			// 			},
			// 		},
			// 	})
			// }
			// log.Debugf("client.GetID():%v,client.OnJoinReply: %v", client.GetID(), client.OnJoinReply)
			// 	subscription := payload.Subscription
			// 	needNegotiate := false
			// 	for _, trackInfo := range subscription.Subscriptions {
			// 		if trackInfo.Subscribe {
			// 			// Add down tracks
			// 			for _, p := range peer.Session().Peers() {
			// 				if p.ID() != peer.ID() {
			// 					for _, track := range p.Publisher().PublisherTracks() {
			// 						if track.Receiver.TrackID() == trackInfo.TrackId && track.Track.RID() == trackInfo.Layer {
			// 							log.Infof("Add RemoteTrack: %v to peer %v %v %v", trackInfo.TrackId, peer.ID(), track.Track.Kind(), track.Track.RID())
			// 							dt, err := peer.Publisher().GetRouter().AddDownTrack(peer.Subscriber(), track.Receiver)
			// 							if err != nil {
			// 								log.Errorf("AddDownTrack error: %v", err)
			// 							}
			// 							// switchlayer
			// 							switch trackInfo.Layer {
			// 							case "f":
			// 								dt.Mute(false)
			// 								_ = dt.SwitchSpatialLayer(2, true)
			// 								log.Infof("%v SwitchSpatialLayer:  2", trackInfo.TrackId)
			// 							case "h":
			// 								dt.Mute(false)
			// 								_ = dt.SwitchSpatialLayer(1, true)
			// 								log.Infof("%v SwitchSpatialLayer:  1", trackInfo.TrackId)
			// 							case "q":
			// 								dt.Mute(false)
			// 								_ = dt.SwitchSpatialLayer(0, true)
			// 								log.Infof("%v SwitchSpatialLayer:  0", trackInfo.TrackId)
			// 							}
			// 							needNegotiate = true
			// 						}
			// 					}
			// 				}
			// 			}
			// 		} else {
			// 			// Remove down tracks
			// 			for _, downTrack := range peer.Subscriber().DownTracks() {
			// 				streamID := downTrack.StreamID()
			// 				if downTrack != nil && downTrack.ID() == trackInfo.TrackId {
			// 					peer.Subscriber().RemoveDownTrack(streamID, downTrack)
			// 					_ = downTrack.Stop()
			// 					needNegotiate = true
			// 				}
			// 			}
			// 		}
			// 	}
			// 	if needNegotiate {
			// 		peer.Subscriber().Negotiate()
			// 	}

			// 	_ = sig.Send(&rtc.Reply{
			// 		Payload: &rtc.Reply_Subscription{
			// 			Subscription: &rtc.SubscriptionReply{
			// 				Success: true,
			// 				Error:   nil,
			// 			},
			// 		},
			// 	})
		}
	}
}
