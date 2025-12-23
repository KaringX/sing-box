// karing
package dhcp

import (
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/dns/transport/local"
	M "github.com/sagernet/sing/common/metadata"
	"github.com/sagernet/sing/service"
)

var (
	cachedServers   []M.Socksaddr
	cachedUpdatedAt time.Time
)

func (t *Transport) getServersFromSystemDNS() []M.Socksaddr {
	systemConfig := local.GetSystemDNSConfig(t.ctx)
	if len(systemConfig.NameServer) > 0 {
		inboundManager := service.FromContext[adapter.InboundManager](t.ctx)
		if inboundManager != nil {
			tunAddressPrefix := inboundManager.GetTunAddressPrefix()
			var serverAddrs []M.Socksaddr
			for _, address := range systemConfig.NameServer {
				contains := false
				ip := M.ParseAddr(address)
				if len(tunAddressPrefix) > 0 {
					for _, tunAddress := range tunAddressPrefix {
						if tunAddress.Contains(ip) {
							contains = true
							break
						}
					}
				}
				if !contains {
					serverAddrs = append(serverAddrs, M.SocksaddrFrom(ip, 53))
				}
			}
			if len(serverAddrs) > 0 {
				return serverAddrs

			}
		}
	}
	return nil
}
