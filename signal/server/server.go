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
			sid := payload.Register.Sid
			uid := payload.Register.Uid
			sourceType := payload.Register.SourceType
			log.Infof("[C=>S] createSession: sid => %v, uid => %v", sid, uid, sourceType)
			//需要查找是否有重名
			err = client.CreateSession(sid, sourceType)

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
			client.Register(sid, uid)
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

			client.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, target int) {
				log.Debugf("[S=>C] peer.OnIceCandidate: target = %v, candidate = %v", target, candidate.Candidate)
				bytes, err := json.Marshal(candidate)
				if err != nil {
					log.Errorf("OnIceCandidate error: %v", err)
				}
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_Trickle{
						Trickle: &rtc.Trickle{
							Init:   string(bytes),
							Target: rtc.Target(target),
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
			client.OnWantControlReply = func(o *rtc.WantControlReply) {
				log.Debugf("[S=>C] client.OnWantControlReply: %v", o.Description)
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_WantControl{
						WantControl: &rtc.WantControlReply{
							Success:     true,
							Error:       nil,
							Description: o.Description,
						},
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
						Sid: s.ID(),
						Uid: s.GetSourceClient().GetID(),
					})
				}
			}
			onlineSources = append(onlineSources, &rtc.OnLineSources{
				Sid: "testSid",
				Uid: "testUid",
			})
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

		case *rtc.Request_WantControl:

			sid := payload.WantControl.Sid
			uid := payload.WantControl.Uid
			log.Infof("[C=>S] join: sid => %v, uid => %v", sid, uid)

			wantControlReply := client.WantControl(sid, uid)

			err = sig.Send(&rtc.Reply{
				Payload: &rtc.Reply_WantControl{
					WantControl: wantControlReply,
				},
			})
			if err != nil {
				log.Errorf("err:%v", err)
			}

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

			client.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, target int) {
				log.Debugf("[S=>C] peer.OnIceCandidate: target = %v, candidate = %v", target, candidate.Candidate)
				bytes, err := json.Marshal(candidate)
				if err != nil {
					log.Errorf("OnIceCandidate error: %v", err)
				}
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_Trickle{
						Trickle: &rtc.Trickle{
							Init:   string(bytes),
							Target: rtc.Target(target),
						},
					},
				})
				if err != nil {
					log.Errorf("OnIceCandidate send error: %v", err)
				}
			}
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
			//发sdp给视频源

			sourceClient := client.Session().GetSourceClient()

			if sourceClient != nil {
				log.Debugf("join desc from client %v to client:%v", client.GetID(), sourceClient.GetID())
				if sourceClient.OnWantControlReply != nil {
					sourceClient.OnWantControlReply(&rtc.WantControlReply{
						Success:     true,
						Uid:         payload.WantControl.Uid,
						Description: payload.WantControl.Description,
					})
				}
			}

		case *rtc.Request_Join:
			sid := payload.Join.Sid
			uid := payload.Join.Uid
			log.Infof("[C=>S] join: sid => %v, uid => %v", sid, uid)
			//fmt.Printf("contains videoSource: %v\r\n", strings.Contains(uid, "videoSource"))

			// if strings.Contains(uid, "videoSource") { //是视频源
			// 	client.CreateSession(sid)
			// 	rtcTarget = rtc.Target_SUBSCRIBER
			// 	err = sig.Send(&rtc.Reply{
			// 		Payload: &rtc.Reply_Join{
			// 			Join: &rtc.JoinReply{
			// 				Success: false,
			// 				Error: &rtc.Error{
			// 					Code:   int32(InternalError),
			// 					Reason: fmt.Sprintf("join error: %v", err),
			// 				},
			// 			},
			// 		},
			// 	})
			// } else { //后面加入的客户端
			// 	rtcTarget = rtc.Target_PUBLISHER
			// 	// bVideoSource = false
			// }
			/*
				视频源在Subscription里创建了session和join，但只是建立了一个
				grpc通道，没有跟wenrtc相关的内容，当客户端join时，才发datachannel 的offer
				给视频源，视频源会自动建立连接，在c.datachannel的回调函数中再publisher
				answer方是PUB，offer方是SUB

				原来的设计是各方join，提供一个datachannel的offer，服务器回复answer,在网页端
				会等待"join-reply"的回复；现在改成树莓派的视频源端发送subscription，保持一个grpc的
				通信通道，由网页客户端发送join，等待"join-reply"的回应，这个join所带的description会
				发给视频源端，回应的是description,如何界定这第一个des，把它当成网页客户端的'join-reply'
				回应给客户端，而且又让后面普通的description当成正常的des回应给客户端呢？
				在session中设置一个IsFirstDatachannelDesc，如果是网页来的join，树莓派回应des，当成第一个，
				以join-reply回应，如果不是，当普通的des。
				断线重连之后，要清空这个IsFirstDatachannelDesc。
			*/

			client.Join(sid, uid)

			// bHadJoinBefore := false
			// if client.Session() == nil {
			// 	log.Errorf("client has no session!")
			// 	continue
			// }
			// clients := client.Session().Clients()
			// for cl := range clients {
			// 	if clients[cl].GetID() == uid {
			// 		bHadJoinBefore = true
			// 	}
			// }
			// if !bHadJoinBefore {
			// 	client.Join(sid, uid)
			// }

			//log.Debugf("client.ProviderSessions(): %v", client.ProviderSessions())
			rtcTarget := rtc.Target_SUBSCRIBER
			// client.OnSessionDescription = func(o *webrtc.SessionDescription) {
			// 	//log.Debugf("[S=>C] Target_SUBSCRIBER: %v", o.SDP)
			// 	// if bVideoSource {
			// 	// 	rtcTarget = rtc.Target_SUBSCRIBER
			// 	// } else {
			// 	// 	rtcTarget = rtc.Target_PUBLISHER
			// 	// }
			// 	err = sig.Send(&rtc.Reply{
			// 		Payload: &rtc.Reply_Description{
			// 			Description: &rtc.SessionDescription{
			// 				Target: rtc.Target(rtcTarget), //需要特别注意SUB和PUB，answer方是PUB，offer方是SUB,视频源应该是pub，控制端主动发起，是Sub
			// 				Sdp:    o.SDP,
			// 				Type:   o.Type.String(),
			// 			},
			// 		},
			// 	})
			// 	if err != nil {
			// 		log.Errorf("negotiation error: %v", err)
			// 	}
			// }

			client.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, target int) {
				log.Debugf("[S=>C] peer.OnIceCandidate: target = %v, candidate = %v", target, candidate.Candidate)
				bytes, err := json.Marshal(candidate)
				if err != nil {
					log.Errorf("OnIceCandidate error: %v", err)
				}
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_Trickle{
						Trickle: &rtc.Trickle{
							Init:   string(bytes),
							Target: rtc.Target(target),
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
			// client.OnICEConnStateChange = func(c webrtc.ICEConnectionState) {
			// 	s.Lock()
			// 	err = sig.Send(&rtc.Reply{
			// 		Payload: &pb.SignalReply_IceConnectionState{
			// 			IceConnectionState: c.String(),
			// 		},
			// 	})
			// 	s.Unlock()

			// 	if err != nil {
			// 		stun.LogDebug.Print(err, "oniceconnectionstatechange error")
			// 	}
			// }

			// desc := webrtc.SessionDescription{
			// 	SDP:  payload.Join.Description.Sdp,
			// 	Type: webrtc.NewSDPType(payload.Join.Description.Type),
			// }

			//clients := client.Session().Clients()
			// for c := range clients {
			// 	// if clients[c].GetID() != client.GetID() {

			// 	// 	log.Debugf("join desc from client %v to client:%v", client.GetID(), clients[c].GetID())
			// 	// 	if clients[c].OnSessionDescription != nil {
			// 	// 		clients[c].OnSessionDescription(&desc)
			// 	// 	}

			// 	// 	if err != nil {
			// 	// 		log.Errorf("grpc send error: %v", err)
			// 	// 		return status.Errorf(codes.Internal, err.Error())
			// 	// 	}
			// 	// }
			// }

		case *rtc.Request_Description:

			// desc := webrtc.SessionDescription{
			// 	SDP:  payload.Description.Sdp,
			// 	Type: webrtc.NewSDPType(payload.Description.Type),
			// }
			//log.Debugf("description:%v", desc)
			clients := client.Session().Clients()
			for c := range clients {
				if clients[c].GetID() != client.GetID() {

					// 	log.Debugf("client.Session().IsFirstDatachannelDesc():%v", client.Session().IsFirstDatachannelDesc())
					// 	if client.Session().IsFirstDatachannelDesc() {
					// 		if clients[c].OnJoinReply != nil {
					// 			log.Debugf("clients[c].OnJoinReply(&desc)")
					// 			clients[c].OnJoinReply(&desc)
					// 		}
					// 		client.Session().SetFirstDatachannelDesc(false)

					// 	} else {
					// 		if clients[c].OnSessionDescription != nil {
					// 			log.Debugf("clients[c].OnSessionDescription(&desc)")
					// 			clients[c].OnSessionDescription(&desc)
					// 		}
					// 	}

					// 	if err != nil {
					// 		log.Errorf("grpc send error: %v", err)
					// 		return status.Errorf(codes.Internal, err.Error())
					// 	}
				}
			}

		case *rtc.Request_Trickle:
			var candidate webrtc.ICECandidateInit
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

			clients := client.Session().Clients()
			for c := range clients {
				if clients[c].GetID() != client.GetID() {
					log.Debugf("trickle from client %v to client:%v", client.GetID(), clients[c].GetID())
					log.Debugf("target %v, candidate %v", int(payload.Trickle.Target), candidate.Candidate)
					if clients[c].OnIceCandidate != nil {
						//pub和sub问题是相对的，具体问题还有待进一步优化
						tempTarget := 0
						if int(payload.Trickle.Target) == 0 {
							tempTarget = 1
						}
						//clients[c].OnIceCandidate(&candidate, int(payload.Trickle.Target))
						clients[c].OnIceCandidate(&candidate, tempTarget)
					}
				}
			}

		case *rtc.Request_Subscription:
			sid := payload.Subscription.Subscriptions[0].TrackId
			uid := payload.Subscription.Subscriptions[0].Layer
			log.Debugf("[C=>S] sid: %v,uid:%s", sid, uid) //, payload.Subscription.trackId, payload.Subscription.layer)

			//client.CreateSession(sid)
			//client.Join(sid, uid)
			rtcTarget = rtc.Target_PUBLISHER

			// client.OnSessionDescription = func(o *webrtc.SessionDescription) {
			// 	log.Debugf("[S=>C] client.OnSessionDescription: %v", o.SDP)
			// 	// if bVideoSource {
			// 	// 	rtcTarget = rtc.Target_SUBSCRIBER
			// 	// } else {
			// 	// 	rtcTarget = rtc.Target_PUBLISHER
			// 	// }
			// 	err = sig.Send(&rtc.Reply{
			// 		Payload: &rtc.Reply_Description{
			// 			Description: &rtc.SessionDescription{
			// 				Target: rtc.Target(rtcTarget), //需要特别注意SUB和PUB，视频源应该是pub，控制端主动发起，是Sub
			// 				Sdp:    o.SDP,
			// 				Type:   o.Type.String(),
			// 			},
			// 		},
			// 	})
			// 	if err != nil {
			// 		log.Errorf("negotiation error: %v", err)
			// 	}
			// }

			client.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, target int) {
				log.Debugf("[S=>C] peer.OnIceCandidate: target = %v, candidate = %v", target, candidate.Candidate)
				bytes, err := json.Marshal(candidate)
				if err != nil {
					log.Errorf("OnIceCandidate error: %v", err)
				}
				err = sig.Send(&rtc.Reply{
					Payload: &rtc.Reply_Trickle{
						Trickle: &rtc.Trickle{
							Init:   string(bytes),
							Target: rtc.Target(target),
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
			log.Debugf("client.GetID():%v,client.OnJoinReply: %v", client.GetID(), client.OnJoinReply)
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
