// karing
package clashapi

import (
	"time"
)

func (s *Server) AddTick(tick *time.Ticker, onClose func()) {
	s.ticks.Store(tick, onClose)
}

func (s *Server) RemoveTick(tick *time.Ticker) {
	s.ticks.Delete(tick)
}

func (s *Server) RemoveTicks() {
	s.ticks.Clear()
}
