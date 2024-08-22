package option

import "net/netip"

type WireGuardOutboundOptions struct {
	DialerOptions
	SystemInterface bool                   `json:"system_interface,omitempty"`
	GSO             bool                   `json:"gso,omitempty"`
	InterfaceName   string                 `json:"interface_name,omitempty"`
	LocalAddress    Listable[netip.Prefix] `json:"local_address"`
	PrivateKey      string                 `json:"private_key"`
	Peers           []WireGuardPeer        `json:"peers,omitempty"`
	ServerOptions
	PeerPublicKey    string            `json:"peer_public_key"`
	PreSharedKey     string            `json:"pre_shared_key,omitempty"`
	Reserved         []uint8           `json:"reserved,omitempty"`
	Workers          int               `json:"workers,omitempty"`
	MTU              uint32            `json:"mtu,omitempty"`
	Network          NetworkList       `json:"network,omitempty"`
	TurnRelay        *TurnRelayOptions `json:"turn_relay,omitempty"`         //hiddify
	FakePackets      string            `json:"fake_packets,omitempty"`       //hiddify
	FakePacketsSize  string            `json:"fake_packets_size,omitempty"`  //hiddify
	FakePacketsDelay string            `json:"fake_packets_delay,omitempty"` //hiddify
	FakePacketsMode  string            `json:"fake_packets_mode,omitempty"`  //hiddify
}

type WireGuardPeer struct {
	ServerOptions
	PublicKey    string           `json:"public_key,omitempty"`
	PreSharedKey string           `json:"pre_shared_key,omitempty"`
	AllowedIPs   Listable[string] `json:"allowed_ips,omitempty"`
	Reserved     []uint8          `json:"reserved,omitempty"`
}
