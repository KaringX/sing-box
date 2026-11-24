//karing

package box

import (
	"github.com/sagernet/sing-box/dns/transport/fakeip"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/tun"
)

func (s *Box) Logger() log.ContextLogger {
	return s.logger
}

func (s *Box) InboundTunSuspend() (*tun.Inbound, *option.TunInboundOptions) {
	//tun reuse
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
	if s.dnsTransport != nil {
		fakeIP := s.dnsTransport.FakeIP()
		if fakeIP != nil && fakeIP.Store() != nil {
			if store, isStore := fakeIP.Store().(*fakeip.Store); isStore {
				store.Suspend()
			}
		}
	}
	return tunInbound, s.tunInboundOptions
}
