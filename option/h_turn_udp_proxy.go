package option

//hiddify
type TurnRelayOptions struct {
	ServerOptions
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Realm    string `json:"realm,omitempty"`
}
