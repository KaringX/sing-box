// https://github.com/shtorm-7/sing-box-extended
package option

type CloudflareProfile struct {
	ID         string `json:"id,omitempty"`
	AuthToken  string `json:"auth_token,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
	LicenseKey string `json:"license_key,omitempty"`
	Recreate   bool   `json:"recreate,omitempty"`
	Detour     string `json:"detour,omitempty"`

	PublicKey   string `json:"public_key,omitempty"` // karing
	IPv4Address string `json:"ipv4_address"`         // karing
	IPv6Address string `json:"ipv6_address"`         // karing
}
