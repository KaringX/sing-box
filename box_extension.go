// karing
package box

import (
	"github.com/sagernet/sing-box/log"
)

func (s *Box) Logger() log.ContextLogger {
	return s.logger
}
