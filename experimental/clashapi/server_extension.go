// karing
package clashapi

import (
	"time"
)

func (s *Server) AddTick(tick *time.Ticker, onClose func()) {
	s.access.RLock()
	defer s.access.RUnlock()
	s.ticks.Store(tick, onClose)
}

func (s *Server) RemoveTick(tick *time.Ticker) {
	s.access.RLock()
	defer s.access.RUnlock()
	s.ticks.Delete(tick)
}

func (s *Server) RemoveTicks() {
	s.access.RLock()
	defer s.access.RUnlock()
	s.ticks.Clear()
}
