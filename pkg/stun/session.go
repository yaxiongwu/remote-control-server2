package stun

import (
	"sync"

	"github.com/yaxiongwu/remote-control-server2/pkg/proto/rtc"
)

// Session represents a set of peers. Transports inside a SessionLocal
// are automatically subscribed to each other.
type Session interface {
	ID() string
	AddClient(client *ClientLocal)
	GetClient(ClientID string) *ClientLocal
	RemoveClient(client *ClientLocal)
	//Clients() []Client
	Clients() []*ClientLocal
	IsFirstDatachannelDesc() bool
	SetFirstDatachannelDesc(b bool)
	SetSourceClient(sourceClient *ClientLocal)
	GetSourceClient() *ClientLocal
	GetSourceType() rtc.SourceType
}

type SessionLocal struct {
	id string
	mu sync.RWMutex
	//config         WebRTCTransportConfig
	//clients map[string]Client
	clients map[string]*ClientLocal
	//relayPeers map[string]*RelayPeer
	closed bool
	// audioObs       *AudioObserver
	// fanOutDCs      []string
	// datachannels   []*Datachannel
	onCloseHandler       func()
	firstDatachannelDesc bool
	sourceType           rtc.SourceType
	sourceClient         *ClientLocal
}

// NewSession creates a new SessionLocal
func NewSession(id string, sourceType rtc.SourceType) Session {
	s := &SessionLocal{
		id:                   id,
		clients:              make(map[string]*ClientLocal),
		firstDatachannelDesc: true,
		sourceType:           sourceType,
	}
	// log1.SetFlags(log1.Ldate | log1.Lshortfile | log1.Ltime)
	// go s.audioLevelObserver(cfg.Router.AudioLevelInterval)
	return s
}

// ID return SessionLocal id
func (s *SessionLocal) ID() string {
	return s.id
}

func (s *SessionLocal) IsFirstDatachannelDesc() bool {
	return s.firstDatachannelDesc
}
func (s *SessionLocal) SetFirstDatachannelDesc(b bool) {
	s.firstDatachannelDesc = b
}

func (s *SessionLocal) AddClient(client *ClientLocal) {
	s.mu.Lock()
	s.clients[client.GetID()] = client
	//Logger.Print("client:", client)
	s.mu.Unlock()
}

func (s *SessionLocal) GetClient(clientID string) *ClientLocal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients[clientID]
}

// RemovePeer removes Peer from the SessionLocal
func (s *SessionLocal) RemoveClient(c *ClientLocal) {
	cid := c.GetID()
	//Logger.V(0).Info("RemovePeer from SessionLocal", "peer_id", pid, "session_id", s.id)
	s.mu.Lock()
	if s.clients[cid].id == c.id {
		delete(s.clients, cid)
	}
	clientCount := len(s.clients)
	s.mu.Unlock()

	// Close SessionLocal if no peers
	if clientCount == 0 {
		s.Close()
	}
}

// Peers returns peers in this SessionLocal
func (s *SessionLocal) Clients() []*ClientLocal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c := make([]*ClientLocal, 0, len(s.clients))
	for _, client := range s.clients {
		c = append(c, client)
	}
	return c
}

// OnClose is called when the SessionLocal is closed
func (s *SessionLocal) OnClose(f func()) {
	s.onCloseHandler = f
}

func (s *SessionLocal) Close() {
	// if !s.closed.set(true) {
	// 	return
	// }
	s.closed = true
	if s.onCloseHandler != nil {
		s.onCloseHandler()
	}
}

func (s *SessionLocal) SetSourceClient(client *ClientLocal) {
	s.sourceClient = client
}

func (s *SessionLocal) GetSourceClient() *ClientLocal {
	return s.sourceClient
}

func (s *SessionLocal) GetSourceType() rtc.SourceType {
	return s.sourceType
}
