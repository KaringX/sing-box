// karing
package box

import (
	"github.com/sagernet/sing-box/log"
)

var contextId int //karing

func (s *Box) Logger() log.ContextLogger {
	return s.logger
}
