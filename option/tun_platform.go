package option

import "github.com/sagernet/sing/common/json/badoption"

type TunPlatformOptions struct {
	HTTPProxy                      *HTTPProxyOptions `json:"http_proxy,omitempty"`
	AllowBypass                    bool              `json:"allow_bypass,omitempty"`                         //karing
	ExcludeApns                    bool              `json:"exclude_apns,omitempty"`                         //karing
	AutoRouteUseSubRangesByDefault bool              `json:"auto_route_use_sub_ranges_by_default,omitempty"` //karing
}

type HTTPProxyOptions struct {
	Enabled bool `json:"enabled,omitempty"`
	ServerOptions
	BypassDomain badoption.Listable[string] `json:"bypass_domain,omitempty"`
	MatchDomain  badoption.Listable[string] `json:"match_domain,omitempty"`
}
