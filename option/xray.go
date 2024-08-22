package option
//hiddify
import "github.com/xtls/xray-core/infra/conf"

type XrayOutboundOptions struct {
	DialerOptions
	Network          NetworkList        `json:"network,omitempty"`
	UDPOverTCP       *UDPOverTCPOptions `json:"udp_over_tcp,omitempty"`
	XrayOutboundJson *map[string]any    `json:"xray_outbound_raw"`
	Fragment         *conf.Fragment     `json:"xray_fragment"`
	LogLevel         string             `json:"xray_loglevel"`
}
