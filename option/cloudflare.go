// https://github.com/shtorm-7/sing-box-extended
package option

type CloudflareProfile struct {
	ID         string `json:"id,omitempty"`
	AuthToken  string `json:"auth_token,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
	LicenseKey string `json:"license_key,omitempty"`
	Recreate   bool   `json:"recreate,omitempty"`
	Detour     string `json:"detour,omitempty"`
}
