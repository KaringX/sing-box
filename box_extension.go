//karing

package box

import (
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/tun"
)

func (s *Box) Logger() log.ContextLogger {
	return s.logger
}

func (s *Box) InboundTunSuspend() (*tun.Inbound, *option.TunInboundOptions) {
	if s.inbound == nil || s.tunInboundOptions == nil || s.tunTag == "" {
		return nil, nil
	}
	inbound, _ := s.inbound.Get(s.tunTag)
	if inbound == nil {
		return nil, nil
	}
	tunInbound, _ := inbound.(*tun.Inbound)
	if tunInbound == nil {
		return nil, nil
	}
	tunInbound.Suspend()
	return tunInbound, s.tunInboundOptions
}
