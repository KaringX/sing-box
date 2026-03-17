// karing
package route

import (
	"context"
	"net/netip"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/process"
	"github.com/sagernet/sing-box/experimental/clashapi/trafficontrol"
	R "github.com/sagernet/sing-box/route/rule"
	E "github.com/sagernet/sing/common/exceptions"
)

var _ adapter.Router = (*Router)(nil)

func (r *Router) ResetOutboundNetwork(tags []string) {
	if r.network != nil {
		r.network.ResetOutboundNetwork(tags)
	}
}

func (r *Router) GetRemoteRuleSetRulesCount() map[string]int {
	counts := make(map[string]int)
	for _, ruleSet := range r.ruleSets {
		if ruleset, isRemote := ruleSet.(*R.RemoteRuleSet); isRemote {
			counts[ruleset.Url()] = ruleset.RulesCount()
		}
	}
	return counts
}

func (r *Router) FindProcessInfo(ctx context.Context, network string, source netip.AddrPort) (*adapter.ConnectionOwner, error) {
	if r.processSearcher != nil {
		var originDestination netip.AddrPort
		return process.FindProcessInfo(r.processSearcher, ctx, network, source, originDestination)
	}
	return nil, E.New("processSearcher not impl")
}

func (r *Router) GetMatchRuleChain(outboundManager adapter.OutboundManager, matchOutboundTag string) ([]string, string, string) {
	return trafficontrol.GetMatchRuleChain(outboundManager, matchOutboundTag)
}

func (r *Router) GetMatchRule(ctx context.Context, metadata *adapter.InboundContext) (adapter.Rule, error) {
	rule, _, _, _, err := r.matchRule(ctx, metadata, false, nil, nil)
	if err != nil {
		return nil, err
	}

	return rule, err
}

func (r *Router) GetAssetContent(path string) ([]byte, error) {
	if r.platformInterface == nil {
		return nil, E.New("platform interface not set")
	}
	return r.platformInterface.GetAssetContent(path)
}
