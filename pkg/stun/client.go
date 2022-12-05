package stun

import (
	"errors"
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
	WantConnect(sid, uid string) *rtc.WantConnectReply
	GetRole() int8
	GetInfo() ClientInfo

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
	GetSession(string) Session
	GetSessions() map[string]Session
	AddSession(Session) error
	NewSession(string, rtc.SourceType) Session
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

type ClientInfo struct {
	Id   string
	Name string
}
type ClientLocal struct {
	sync.Mutex
	//id   string
	//name string
	info ClientInfo
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
	OnSessionDescription      func(*webrtc.SessionDescription, string, string)
	OnIceCandidate            func(*webrtc.ICECandidateInit, string, string)
	OnICEConnStateChange      func(webrtc.ICEConnectionState)
	OnJoinReply               func(*webrtc.SessionDescription)
	OnWantConnectRequest      func(*rtc.WantConnectRequest)
	OnWantConnectReply        func(*rtc.WantConnectReply)
	OnWantConnectRequestReply func(*rtc.WantConnectRequest)
}

// NewPeer creates a new PeerLocal for signaling with the given SFU
func NewClient(provider SessionProvider) *ClientLocal {
	log.SetFlags(log.Ldate | log.Lshortfile)
	return &ClientLocal{
		provider: provider,
	}
}

func (c *ClientLocal) Session() Session {
	if c.session != nil {
		return c.session
	} else {
		return nil
	}

}

// ID return the peer id
func (c *ClientLocal) GetID() string {
	return c.info.Id
}

// ID return the peer id
func (c *ClientLocal) SetID(cid string) error {
	c.info.Id = cid
	return nil
}

// Name return the peer name
func (c *ClientLocal) GetName() string {
	return c.info.Name
}

// ID return the peer id
func (c *ClientLocal) SetName(name string) error {
	c.info.Name = name
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
	//fmt.Println("Before RemoveClient")
	if c.session == nil {
		return nil //getonlineSources的时候没有加入到session中去
	}

	if c.session.GetClient(c.info.Id) != nil {
		c.session.RemoveClient(c)
	}
	// for i, c := range c.session.Clients() {
	// 	fmt.Println(i, ".", c.id)
	// }

	// if c.session != nil {
	// 	if c.session.Clients[c.id] != nil {
	// 		c.session.RemoveClient(c)
	// 	}
	// }
	// fmt.Println("After RemoveClient")
	// for i, c := range c.session.Clients() {
	// 	fmt.Println(i, ".", c.id)
	// }
	return nil
}

func (c *ClientLocal) CreateSession(sid string, sourceType rtc.SourceType) error {
	/*
		session只能在stun里创建，如果在这里创建，client实例释放的时候，是不是此处创建的session作为内部变量会释放回收？
		session自己一般管理clients，session无法获得上一级的stun，
		client通过provider，跳过session，直接获取stun
	*/

	seesions := c.provider.GetSessions()

	if seesions[sid] != nil {
		return errors.New("Session already exists")
	}

	newSeesion := c.provider.NewSession(sid, sourceType) //在stun里创建，但是在什么释放呢？
	c.session = newSeesion

	return nil
}

func (c *ClientLocal) Register(uid, name string, sourceType rtc.SourceType) error {
	if uid == "" {
		uid = cuid.New()
	}
	c.info.Id = uid
	c.info.Name = name
	c.CreateSession(uid, sourceType)
	s := c.provider.GetSession(uid)
	if s == nil {
		return errors.New("no seesion exists")
	}
	s.SetSourceClient(c)
	s.AddClient(c)
	c.session = s
	return nil
}

func (c *ClientLocal) Add(uid, name string) error {
	if uid == "" {
		uid = cuid.New()
	}
	c.info.Id = uid
	c.info.Name = name
	s := c.provider.GetSession(uid)
	if s == nil {
		return errors.New("no seesion exists")
	}
	s.AddClient(c)
	c.session = s
	return nil
}

// Join initializes this peer for a given sourceID
func (c *ClientLocal) WantConnect(from string, uid string) *rtc.WantConnectReply {

	// if c.source != nil {
	// 	//Logger.V(1).Info("peer already exists", "source_id", sid, "peer_id", p.id, "publisher_id", p.publisher.id)
	// 	return ErrTransportExists
	// }
	//println("WantConnect,from:%v,to:%v", from, uid)
	if from == "" {
		from = cuid.New()
	}
	c.info.Id = from
	c.role = CONTROL

	idleOrNot := true
	s := c.provider.GetSession(uid)
	if s == nil {
		return &rtc.WantConnectReply{
			Success: true,
			Error:   &rtc.Error{Reason: "no target"},
		}
	}
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

	return &rtc.WantConnectReply{
		Success:      true,
		IdleOrNot:    idleOrNot,
		RestTimeSecs: 101,
		NumOfWaiting: 121,
	}
}

func (c *ClientLocal) ProviderSessions() map[string]Session {
	p := c.provider
	return p.GetSessions()

}

func (c *ClientLocal) GetRole() int8 {
	return c.role
}

func (c *ClientLocal) GetInfo() ClientInfo {
	return c.info
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
