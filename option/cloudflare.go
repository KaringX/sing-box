// https://github.com/shtorm-7/sing-box-extended
package option

type CloudflareProfile struct {
	//ID         string `json:"id,omitempty"`// karing
	//AuthToken  string `json:"auth_token,omitempty"`// karing
	PrivateKey string `json:"private_key,omitempty"`
	//LicenseKey string `json:"license_key,omitempty"`// karing
	//Recreate   bool   `json:"recreate,omitempty"`// karing
	//Detour     string `json:"detour,omitempty"`// karing

	PublicKey   string `json:"public_key,omitempty"` // karing
	IPv4Address string `json:"ipv4_address"`         // karing
	IPv6Address string `json:"ipv6_address"`         // karing
}
