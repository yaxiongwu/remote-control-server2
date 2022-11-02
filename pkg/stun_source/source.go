package stun

import (
	"sync"

	"github.com/yaxiongwu/remote-control-server2/pkg/proto/rtc"
)

// Source represents a set of peers. Transports inside a SourceLocal
// are automatically subscribed to each other.
type Source interface {
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

type SourceLocal struct {
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

// NewSource creates a new SourceLocal
func NewSource(id string, sourceType rtc.SourceType) Source {
	s := &SourceLocal{
		id:                   id,
		clients:              make(map[string]*ClientLocal),
		firstDatachannelDesc: true,
		sourceType:           sourceType,
	}
	// log1.SetFlags(log1.Ldate | log1.Lshortfile | log1.Ltime)
	// go s.audioLevelObserver(cfg.Router.AudioLevelInterval)
	return s
}

// ID return SourceLocal id
func (s *SourceLocal) ID() string {
	return s.id
}

func (s *SourceLocal) IsFirstDatachannelDesc() bool {
	return s.firstDatachannelDesc
}
func (s *SourceLocal) SetFirstDatachannelDesc(b bool) {
	s.firstDatachannelDesc = b
}

func (s *SourceLocal) AddClient(client *ClientLocal) {
	s.mu.Lock()
	s.clients[client.GetID()] = client
	//Logger.Print("client:", client)
	s.mu.Unlock()
}

func (s *SourceLocal) GetClient(clientID string) *ClientLocal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients[clientID]
}

// RemovePeer removes Peer from the SourceLocal
func (s *SourceLocal) RemoveClient(c *ClientLocal) {
	cid := c.GetID()
	//Logger.V(0).Info("RemovePeer from SourceLocal", "peer_id", pid, "source_id", s.id)
	s.mu.Lock()
	if s.clients[cid].id == c.id {
		delete(s.clients, cid)
	}
	clientCount := len(s.clients)
	s.mu.Unlock()

	// Close SourceLocal if no peers
	if clientCount == 0 {
		s.Close()
	}
}

// Peers returns peers in this SourceLocal
func (s *SourceLocal) Clients() []*ClientLocal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c := make([]*ClientLocal, 0, len(s.clients))
	for _, client := range s.clients {
		c = append(c, client)
	}
	return c
}

// OnClose is called when the SourceLocal is closed
func (s *SourceLocal) OnClose(f func()) {
	s.onCloseHandler = f
}

func (s *SourceLocal) Close() {
	// if !s.closed.set(true) {
	// 	return
	// }
	s.closed = true
	if s.onCloseHandler != nil {
		s.onCloseHandler()
	}
}

func (s *SourceLocal) SetSourceClient(client *ClientLocal) {
	s.sourceClient = client
}

func (s *SourceLocal) GetSourceClient() *ClientLocal {
	return s.sourceClient
}

func (s *SourceLocal) GetSourceType() rtc.SourceType {
	return s.sourceType
}
