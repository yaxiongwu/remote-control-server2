package stun

import (
	"log"
	"os"
	"sync"
)

var LogInfo = log.New(os.Stdout, "[Info] ", log.LstdFlags|log.Lshortfile)
var LogDebug = log.New(os.Stdout, "[Debug] ", log.LstdFlags|log.Lshortfile)
var LogError = log.New(os.Stdout, "[Error] ", log.LstdFlags|log.Lshortfile)

// STUN represents an sfu instance
type STUN struct {
	sync.RWMutex

	sources map[string]Source
	// datachannels []*Datachannel
	// withStats    bool
}

// NewSFU creates a new sfu instance
func NewSTUN() *STUN {
	stun := &STUN{
		sources: make(map[string]Source),
	}
	return stun
}

// NewSource creates a new SourceLocal instance
func (s *STUN) newSource(id string) Source {

	source := NewSource(id, 0).(*SourceLocal) //NewSource(id)返回的是Source,类型是接口，这个 . (*SourceLocal)如何理解？
	source.OnClose(func() {
		s.Lock()
		delete(s.sources, id)
		s.Unlock()
	})
	s.Lock()
	s.sources[id] = source
	s.Unlock()
	return source
}

// GetSource by id
func (s *STUN) getSource(id string) Source {
	s.RLock()
	defer s.RUnlock()
	return s.sources[id]
}

func (s *STUN) GetSource(sid string) Source {
	source := s.getSource(sid)
	if source == nil {
		source = s.newSource(sid)
	}
	return source
}

// GetSources return all sources
func (s *STUN) GetSources() []Source {
	s.RLock()
	defer s.RUnlock()
	sources := make([]Source, 0, len(s.sources))
	for _, source := range s.sources {
		sources = append(sources, source)
	}
	return sources
}
