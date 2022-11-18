package stun

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/lucsky/cuid"
	"github.com/pion/webrtc/v3"
	"github.com/yaxiongwu/remote-control-server2/pkg/proto/rtc"
)

type Client interface {
	GetID() string
	GetName() string
	SetID(cid string) error
	SetName(name string) error
	CreateSession(sid string) error
	Session() Session
	Close() error
	WantControl(sid, uid string) *rtc.WantControlReply
	GetRole() int8

	// InterOnOfferJSON2GRCP(*webrtc.SessionDescription) func()
	// InterOnOfferGRCP2JSON(*webrtc.SessionDescription) func()
	// InterOnIceCandidateJSON2GRCP(*webrtc.ICECandidateInit, int) func()
	// InterOnIceCandidateGRCP2JSON(*webrtc.ICECandidateInit, int) func()
	// InterOnICEConnStateChangeJSON2GRCP(webrtc.ICEConnectionState) func()
	// InterOnICEConnStateChangeGRCP2JSON(webrtc.ICEConnectionState) func()

	//SendDCMessage(label string, msg []byte) error
}

var (
	// ErrTransportExists join is called after a peerconnection is established
	ErrTransportExists = errors.New("rtc transport already exists for this connection")
	// ErrNoTransportEstablished cannot signal before join
	ErrNoTransportEstablished = errors.New("no rtc transport exists for this Peer")
	// ErrOfferIgnored if offer received in unstable state
	ErrOfferIgnored = errors.New("offered ignored")
)

// JoinConfig allow adding more control to the peers joining a SessionLocal.
type JoinConfig struct {
	// If true the peer will not be allowed to publish tracks to SessionLocal.
	NoPublish bool
	// If true the peer will not be allowed to subscribe to other peers in SessionLocal.
	NoSubscribe bool
	// If true the peer will not automatically subscribe all tracks,
	// and then the peer can use peer.Subscriber().AddDownTrack/RemoveDownTrack
	// to customize the subscrbe stream combination as needed.
	// this parameter depends on NoSubscribe=false.
	NoAutoSubscribe          bool
	VideoSourceToBeControled bool //标识是否是视频源，视频被控制端
}

// SessionProvider provides the SessionLocal to the sfu.Peer
// This allows the sfu.SFU{} implementation to be customized / wrapped by another package
type SessionProvider interface {
	GetSession(sid string) Session
	GetSessions() []Session
}

/*
  由于有的使用grpc，有的使用websocket+json，所以还需要数据装换，无法直接将受到的数据转发

*/
// SFU represents an sfu instance
const (
	SOURCE  = 0
	CONTROL = 1
	VIEW    = 2
	UNKNOWN = 3
)

type ClientLocal struct {
	sync.Mutex
	id   string
	name string
	//closed   bool
	session          Session
	provider         SessionProvider //这个provider有什么用？提供source之上的STUN?
	role             int8
	timeBeginControl time.Time
	timeEndControl   time.Time
	timeBeginView    time.Time
	timeEndView      time.Time
	VideoSourceID    string
	// OnOfferJSON2GRCP              func(*webrtc.SessionDescription)
	// OnOfferGRCP2JSONReply         func(*webrtc.SessionDescription)
	// OnOfferGRCP2JSONNotify        func(*webrtc.SessionDescription)
	// OnIceCandidateJSON2GRCP       func(*webrtc.ICECandidateInit, int)
	// OnIceCandidateGRCP2JSON       func(*webrtc.ICECandidateInit, int)
	// OnICEConnStateChangeJSON2GRCP func(webrtc.ICEConnectionState)
	// OnICEConnStateChangeGRCP2JSON func(webrtc.ICEConnectionState)
	OnSessionDescription func(*webrtc.SessionDescription, string, string)
	OnIceCandidate       func(*webrtc.ICECandidateInit, string, string)
	OnICEConnStateChange func(webrtc.ICEConnectionState)
	OnJoinReply          func(*webrtc.SessionDescription)
	OnWantControlRequest func(*rtc.WantControlRequest)
	OnWantControlReply   func(*rtc.WantControlReply)
}

// NewPeer creates a new PeerLocal for signaling with the given SFU
func NewClient(provider SessionProvider) *ClientLocal {
	log.SetFlags(log.Ldate | log.Lshortfile)
	return &ClientLocal{
		provider: provider,
	}
}

func (c *ClientLocal) Session() Session {
	return c.session
}

func (c *ClientLocal) CreateSession(sid string, sourceType rtc.SourceType) error {

	s := c.provider.GetSessions()

	for _, session := range s {
		if session.ID() == sid {
			return errors.New("Session already exists")
		}
	}

	if c.session != nil {
		if c.session.ID() == sid {
			return nil
		}
	}
	c.session = NewSession(sid, sourceType)
	return nil
}

// ID return the peer id
func (c *ClientLocal) GetID() string {
	return c.id
}

// ID return the peer id
func (c *ClientLocal) SetID(cid string) error {
	c.id = cid
	return nil
}

// Name return the peer name
func (c *ClientLocal) GetName() string {
	return c.name
}

// ID return the peer id
func (c *ClientLocal) SetName(name string) error {
	c.name = name
	return nil
}

// Close shuts down the peer connection and sends true to the done channel
func (c *ClientLocal) Close() error {
	c.Lock()
	defer c.Unlock()

	// if !c.closed.set(true) {
	// 	return nil
	// }
	//c.closed = true
	fmt.Println("Before RemoveClient")
	if c.session == nil {
		return nil //getonlineSources的时候没有加入到session中去
	}
	for i, c := range c.session.Clients() {
		fmt.Println(i, ".", c.id)
	}

	if c.session != nil {
		c.session.RemoveClient(c)
		c.session.SetFirstDatachannelDesc(true)
	}
	fmt.Println("After RemoveClient")
	for i, c := range c.session.Clients() {
		fmt.Println(i, ".", c.id)
	}
	return nil
}

func (c *ClientLocal) Register(uid, name string) error {
	if uid == "" {
		uid = cuid.New()
	}
	c.id = uid
	c.name = name
	s := c.provider.GetSession(uid)
	s.SetSourceClient(c)
	s.AddClient(c)
	c.session = s
	return nil
}

func (c *ClientLocal) Add(uid, name string) error {
	if uid == "" {
		uid = cuid.New()
	}
	c.id = uid
	c.name = name
	s := c.provider.GetSession(uid)
	s.AddClient(c)
	c.session = s
	return nil
}

// Join initializes this peer for a given sessionID
func (c *ClientLocal) Join(sid, uid string) error {

	// if c.session != nil {
	// 	//Logger.V(1).Info("peer already exists", "session_id", sid, "peer_id", p.id, "publisher_id", p.publisher.id)
	// 	return ErrTransportExists
	// }
	println("peer_Join,uid:%v", uid)
	if uid == "" {
		uid = cuid.New()
	}
	c.id = uid

	s := c.provider.GetSession(sid)
	//Logger.Printf("join,*c:%v,c:%v,&c:%v", *c, c, &c)
	//需要处理断线、第二个用户登录等问题
	clients := s.Clients()
	for _, client := range clients {
		/*
			!!!!!!!!!!!!!!注意
			断线之后没有清空数据，有两种方法管理连接：
			1. grpc可以监听连接状态，在grpc.NewServe()函数中传入statsHandler就可以监听各状态和标签，但是跟client signal很难联系起来。
			2. 往client中发数据，如果连接中断会报错“transport is closing”

		*/
		//每一个手机连进来的UID是唯一的，且不变的，如果有存这个手机的ID，可能是断网后重连的
		println("clientID:%s", client.GetID())
		if client.GetID() == uid {
			s.RemoveClient(client)
			s.SetFirstDatachannelDesc(true)
			print("s.SetFirstDatachannelDesc:%s", client.GetID())
		}
	}

	s.AddClient(c)
	c.session = s

	return nil
}

// Join initializes this peer for a given sourceID
func (c *ClientLocal) WantControl(from string, to string) *rtc.WantControlReply {

	// if c.source != nil {
	// 	//Logger.V(1).Info("peer already exists", "source_id", sid, "peer_id", p.id, "publisher_id", p.publisher.id)
	// 	return ErrTransportExists
	// }
	println("peer_Join,from:%v,to:%v", from, to)
	if from == "" {
		from = cuid.New()
	}
	c.id = from
	c.role = CONTROL

	idleOrNot := true

	s := c.provider.GetSession(to)
	//Logger.Printf("join,*c:%v,c:%v,&c:%v", *c, c, &c)
	//需要处理断线、第二个用户登录等问题
	clients := s.Clients()
	for _, client := range clients {
		println("clientID:%s", client.GetID())
		if client.GetRole() == CONTROL {
			idleOrNot = false
		}
	}

	s.AddClient(c)
	c.session = s

	return &rtc.WantControlReply{
		Success:                 true,
		IdleOrNot:               idleOrNot,
		RestTimeSecOfControling: 101,
		NumOfWaiting:            121,
	}
}

func (c *ClientLocal) ProviderSessions() []Session {
	p := c.provider
	return p.GetSessions()

}

func (c *ClientLocal) GetRole() int8 {
	return c.role
}

// func (c *ClientLocal) OnOfferJSON2GRCP(*webrtc.SessionDescription) {
// 	return c.OnOfferJSON2GRCP
// }
// func (c *ClientLocal) OnOfferGRCP2JSONReply(*webrtc.SessionDescription) {
// 	return c.OnOfferGRCP2JSONReply
// }
// func (c *ClientLocal) OnIceCandidateJSON2GRCP(*webrtc.ICECandidateInit, int) {
// 	return c.OnIceCandidateJSON2GRCP
// }
// func (c *ClientLocal) OnIceCandidateGRCP2JSON(*webrtc.ICECandidateInit, int) {
// 	return c.OnIceCandidateGRCP2JSON
// }
// func (c *ClientLocal) OnICEConnStateChangeJSON2GRCP(webrtc.ICEConnectionState) {
// 	return c.OnICEConnStateChangeJSON2GRCP
// }
// func (c *ClientLocal) OnICEConnStateChangeGRCP2JSON(webrtc.ICEConnectionState) {
// 	return c.OnICEConnStateChangeGRCP2JSON
// }
