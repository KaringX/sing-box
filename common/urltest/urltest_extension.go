// karing
package urltest

import (
	"github.com/sagernet/sing-box/adapter"
)

func (s *HistoryStorage) GetURLTestHistory() map[string]*adapter.URLTestHistory {
	history := make(map[string]*adapter.URLTestHistory)
	s.access.Lock()
	for k, v := range s.delayHistory {
		history[k] = v
	}
	s.access.Unlock()
	return history
}
