//go:build with_karing

package box

import (
	"github.com/sagernet/sing-box/log"
)

var contextId int

func (s *Box) Logger() log.ContextLogger {
	return s.logger
}
