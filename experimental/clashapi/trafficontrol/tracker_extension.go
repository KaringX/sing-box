//go:build with_karing

package trafficontrol

import (
	"github.com/sagernet/sing-box/adapter"
)

func GetMatchRuleChain(outboundManager adapter.OutboundManager, matchOutboundTag string) ([]string, string, string) {
	var (
		chain        []string
		next         string
		outbound     string
		outboundType string
	)
	if outboundManager == nil {
		return chain, outbound, outboundType
	}
	if len(matchOutboundTag) != 0 {
		next = matchOutboundTag
	} else {
		next = outboundManager.Default().Tag()
	}
	for {
		detour, loaded := outboundManager.Outbound(next)
		if !loaded {
			break
		}
		chain = append(chain, next)
		outbound = detour.Tag()
		outboundType = detour.Type()
		group, isGroup := detour.(adapter.OutboundGroup)
		if !isGroup || group == nil {
			break
		}
		next = group.Now()
	}
	return chain, outbound, outboundType
}
