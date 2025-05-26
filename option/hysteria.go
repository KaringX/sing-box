package option

import (
	"fmt"
	"strings"

	"github.com/sagernet/sing/common/byteformats"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/json/badoption"
)

type HysteriaInboundOptions struct {
	ListenOptions
	Up                  *byteformats.NetworkBytesCompat `json:"up,omitempty"`
	UpMbps              int                             `json:"up_mbps,omitempty"`
	Down                *byteformats.NetworkBytesCompat `json:"down,omitempty"`
	DownMbps            int                             `json:"down_mbps,omitempty"`
	Obfs                string                          `json:"obfs,omitempty"`
	Users               []HysteriaUser                  `json:"users,omitempty"`
	ReceiveWindowConn   uint64                          `json:"recv_window_conn,omitempty"`
	ReceiveWindowClient uint64                          `json:"recv_window_client,omitempty"`
	MaxConnClient       int                             `json:"max_conn_client,omitempty"`
	DisableMTUDiscovery bool                            `json:"disable_mtu_discovery,omitempty"`
	InboundTLSOptionsContainer
}

type HysteriaUser struct {
	Name       string `json:"name,omitempty"`
	Auth       []byte `json:"auth,omitempty"`
	AuthString string `json:"auth_str,omitempty"`
}

type HysteriaOutboundOptions struct {
	DialerOptions
	ServerOptions
	ServerPorts         badoption.Listable[string]      `json:"server_ports,omitempty"`
	HopInterval         HopIntervalValue                `json:"hop_interval,omitempty"` //karing
	Up                  *byteformats.NetworkBytesCompat `json:"up,omitempty"`
	UpMbps              int                             `json:"up_mbps,omitempty"`
	Down                *byteformats.NetworkBytesCompat `json:"down,omitempty"`
	DownMbps            int                             `json:"down_mbps,omitempty"`
	Obfs                string                          `json:"obfs,omitempty"`
	Auth                []byte                          `json:"auth,omitempty"`
	AuthString          string                          `json:"auth_str,omitempty"`
	ReceiveWindowConn   uint64                          `json:"recv_window_conn,omitempty"`
	ReceiveWindow       uint64                          `json:"recv_window,omitempty"`
	DisableMTUDiscovery bool                            `json:"disable_mtu_discovery,omitempty"`
	Network             NetworkList                     `json:"network,omitempty"`
	OutboundTLSOptionsContainer
	TurnRelay *TurnRelayOptions `json:"turn_relay,omitempty"` //hiddify
	HopPorts  HopPortsValue     `json:"hop_ports,omitempty"`  //https://github.com/morgenanno/sing-box //"114,514,810-1919"
}

type _HysteriaOutboundOptions HysteriaOutboundOptions //karing
func (m *HysteriaOutboundOptions) UnmarshalJSON(bytes []byte) error { //karing
	err := json.Unmarshal(bytes, (*_HysteriaOutboundOptions)(m))
	if err != nil {
		return err
	}

	if len(m.ServerPorts) == 0 && len(m.HopPorts) > 0 {
		ports := strings.Split(string(m.HopPorts), ",")
		for i := 0; i < len(ports); i++ {
			ports[i] = strings.Replace(ports[i], "-", ":", -1)
			parts := strings.Split(ports[i], ":")
			if len(parts) == 1 {
				m.ServerPorts = append(m.ServerPorts, fmt.Sprintf("%s:%s", parts[0], parts[0]))
			} else if len(parts) == 2 {
				m.ServerPorts = append(m.ServerPorts, fmt.Sprintf("%s:%s", parts[0], parts[1]))
			} else {
				return E.New("invalid hop_ports format: ", string(m.HopPorts))
			}
		}
	}
	m.HopPorts = ""
	return nil
}
