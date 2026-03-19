// karing
package local

import (
	"context"

	"github.com/sagernet/sing-box/adapter"
	M "github.com/sagernet/sing/common/metadata"
	"github.com/sagernet/sing/service"
)

func GetSystemDNSConfig(ctx context.Context) *dnsConfig {
	return getSystemDNSConfig(ctx)
}

func getServersFromSystemDNS(ctx context.Context) []M.Socksaddr {
	systemConfig := getSystemDNSConfig(ctx)
	if len(systemConfig.NameServer) > 0 {
		inboundManager := service.FromContext[adapter.InboundManager](ctx)
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
