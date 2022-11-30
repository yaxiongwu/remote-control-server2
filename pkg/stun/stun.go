package stun

import (
	"log"
	"os"
	"sync"

	"github.com/yaxiongwu/remote-control-server2/pkg/proto/rtc"
)

var LogInfo = log.New(os.Stdout, "[Info] ", log.LstdFlags|log.Lshortfile)
var LogDebug = log.New(os.Stdout, "[Debug] ", log.LstdFlags|log.Lshortfile)
var LogError = log.New(os.Stdout, "[Error] ", log.LstdFlags|log.Lshortfile)

// STUN represents an sfu instance
type STUN struct {
	sync.RWMutex

	sessions map[string]Session
	// datachannels []*Datachannel
	// withStats    bool
}

// NewSFU creates a new sfu instance
func NewSTUN() *STUN {
	stun := &STUN{
		sessions: make(map[string]Session),
	}
	return stun
}

// NewSession creates a new SessionLocal instance
func (s *STUN) NewSession(sid string, sourceType rtc.SourceType) Session {

	//session := NewSession(id, 0).(*SessionLocal) //NewSession(id)返回的是Session,类型是接口，这个 . (*SessionLocal)如何理解？
	session := &SessionLocal{
		id:                   sid,
		clients:              make(map[string]*ClientLocal),
		firstDatachannelDesc: true,
		sourceType:           sourceType,
	}

	session.OnClose(func() {
		s.Lock()
		delete(s.sessions, sid)
		s.Unlock()
	})

	s.Lock()
	s.sessions[sid] = session
	s.Unlock()
	return session
}

// GetSession by id

func (s *STUN) GetSession(sid string) Session {
	s.RLock()
	defer s.RUnlock()
	return s.sessions[sid]
}

// GetSessions return all sessions
func (s *STUN) GetSessions() map[string]Session {
	s.RLock()
	defer s.RUnlock()
	// sessions := make([]Session, 0, len(s.sessions))
	// for _, session := range s.sessions {
	// 	sessions = append(sessions, session)
	// }
	return s.sessions
}

func (s *STUN) AddSession(session Session) error {
	s.RLock()

	defer s.RUnlock()
	s.sessions[session.ID()] = session
	return nil
}
