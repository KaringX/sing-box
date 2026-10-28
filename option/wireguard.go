package option

import (
	"net/netip"

	Xbadoption "github.com/sagernet/sing-box/common/xray/json/badoption"
	"github.com/sagernet/sing/common/json/badoption"
)

type WireGuardEndpointOptions struct {
	System       bool                             `json:"system,omitempty"`
	Name         string                           `json:"name,omitempty"`
	MTU          uint32                           `json:"mtu,omitempty"`
	Address      badoption.Listable[netip.Prefix] `json:"address"`
	PrivateKey   string                           `json:"private_key"`
	ListenPort   uint16                           `json:"listen_port,omitempty"`
	Peers        []WireGuardPeer                  `json:"peers,omitempty"`
	UDPTimeout   badoption.Duration               `json:"udp_timeout,omitempty"`
	UDPMapping   UDPNATBehavior                   `json:"udp_mapping,omitempty"`
	UDPFiltering UDPNATBehavior                   `json:"udp_filtering,omitempty"`
	UDPNATMax    uint32                           `json:"udp_nat_max,omitempty"`
	Workers      int                              `json:"workers,omitempty"`
	DialerOptions
	PreallocatedBuffersPerPool uint32            `json:"preallocated_buffers_per_pool,omitempty"` // https://github.com/shtorm-7/sing-box-extended
	DisablePauses              bool              `json:"disable_pauses,omitempty"`                // https://github.com/shtorm-7/sing-box-extended
	Amnezia                    *WireGuardAmnezia `json:"amnezia,omitempty"`                       // https://github.com/shtorm-7/sing-box-extended
}

type WireGuardPeer struct {
	Address                     string                           `json:"address,omitempty"`
	Port                        uint16                           `json:"port,omitempty"`
	PublicKey                   string                           `json:"public_key,omitempty"`
	PreSharedKey                string                           `json:"pre_shared_key,omitempty"`
	AllowedIPs                  badoption.Listable[netip.Prefix] `json:"allowed_ips,omitempty"`
	PersistentKeepaliveInterval uint16                           `json:"persistent_keepalive_interval,omitempty"`
	Reserved                    []uint8                          `json:"reserved,omitempty"`
}

type WireGuardAmnezia struct { // https://github.com/shtorm-7/sing-box-extended
	JC                     int               `json:"jc,omitempty"`
	JMin                   int               `json:"jmin,omitempty"`
	JMax                   int               `json:"jmax,omitempty"`
	S1                     int               `json:"s1,omitempty"`
	S2                     int               `json:"s2,omitempty"`
	S3                     int               `json:"s3,omitempty"`
	S4                     int               `json:"s4,omitempty"`
	H1                     *Xbadoption.Range `json:"h1,omitempty"`
	H2                     *Xbadoption.Range `json:"h2,omitempty"`
	H3                     *Xbadoption.Range `json:"h3,omitempty"`
	H4                     *Xbadoption.Range `json:"h4,omitempty"`
	I1                     string            `json:"i1,omitempty"`
	I2                     string            `json:"i2,omitempty"`
	I3                     string            `json:"i3,omitempty"`
	I4                     string            `json:"i4,omitempty"`
	I5                     string            `json:"i5,omitempty"`
	HeaderProtectionKey    string            `json:"header_protection_key,omitempty"`
	ContentPaddingAddition *Xbadoption.Range `json:"content_padding_addition,omitempty"`
	RekeyAfterTime         *Xbadoption.Range `json:"rekey_after_time,omitempty"`
	RekeyTimeout           *Xbadoption.Range `json:"rekey_timeout,omitempty"`
	RejectAfterTime        *Xbadoption.Range `json:"reject_after_time,omitempty"`
	KeepaliveTimeout       *Xbadoption.Range `json:"keepalive_timeout,omitempty"`
	MaxHandshakeAttempts   *Xbadoption.Range `json:"max_handshake_attempts,omitempty"`
}
