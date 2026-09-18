package constant

const (
	TypeTun                = "tun"
	TypeRedirect           = "redirect"
	TypeTProxy             = "tproxy"
	TypeDirect             = "direct"
	TypeBridge             = "bridge"
	TypeBlock              = "block"
	TypeDNS                = "dns"
	TypeSOCKS              = "socks"
	TypeHTTP               = "http"
	TypeMixed              = "mixed"
	TypeShadowsocks        = "shadowsocks"
	TypeSnell              = "snell"
	TypeVMess              = "vmess"
	TypeTrojan             = "trojan"
	TypeTrustTunnel        = "trusttunnel" // https://github.com/shtorm-7/sing-box-extended
	TypeNaive              = "naive"
	TypeWireGuard          = "wireguard"
	TypeWARP               = "warp"   // https://github.com/shtorm-7/sing-box-extended
	TypeMASQUE             = "masque" // https://github.com/shtorm-7/sing-box-extended
	TypeHysteria           = "hysteria"
	TypeTor                = "tor"
	TypeSSH                = "ssh"
	TypeShadowTLS          = "shadowtls"
	TypeMieru              = "mieru" //karing https://github.com/enfein/mbox
	TypeAnyTLS             = "anytls"
	TypeSudoku             = "sudoku" // https://github.com/shtorm-7/sing-box-extended
	TypeShadowsocksR       = "shadowsocksr"
	TypeVLESS              = "vless"
	TypeTUIC               = "tuic"
	TypeHysteria2          = "hysteria2"
	TypeOpenConnect        = "openconnect"
	TypeOpenVPNClient      = "openvpn-client"
	TypeOpenVPNServer      = "openvpn-server"
	TypeVPNServer          = "vpn-server" // https://github.com/shtorm-7/sing-box-extended
	TypeVPNClient          = "vpn-client" // https://github.com/shtorm-7/sing-box-extended
	TypeTailscale          = "tailscale"
	TypeCloudflared        = "cloudflared"
	TypeDERP               = "derp"
	TypeResolved           = "resolved"
	TypeSSMAPI             = "ssm-api"
	TypeAPI                = "api"
	TypeCCM                = "ccm"
	TypeOCM                = "ocm"
	TypeOOMKiller          = "oom-killer"
	TypeUSBIPServer        = "usbip-server"
	TypeUSBIPClient        = "usbip-client"
	TypeHysteriaRealm      = "hysteria-realm"
	TypeACME               = "acme"
	TypeCloudflareOriginCA = "cloudflare-origin-ca"
)

const (
	TypeSelector = "selector"
	TypeURLTest  = "urltest"
)

func ProxyDisplayName(proxyType string) string {
	switch proxyType {
	case TypeTun:
		return "TUN"
	case TypeRedirect:
		return "Redirect"
	case TypeTProxy:
		return "TProxy"
	case TypeDirect:
		return "Direct"
	case TypeBridge:
		return "Bridge"
	case TypeBlock:
		return "Block"
	case TypeDNS:
		return "DNS"
	case TypeSOCKS:
		return "SOCKS"
	case TypeHTTP:
		return "HTTP"
	case TypeMixed:
		return "Mixed"
	case TypeShadowsocks:
		return "Shadowsocks"
	case TypeSnell:
		return "Snell"
	case TypeVMess:
		return "VMess"
	case TypeTrojan:
		return "Trojan"
	case TypeTrustTunnel: // https://github.com/shtorm-7/sing-box-extended
		return "TrustTunnel"
	case TypeNaive:
		return "Naive"
	case TypeWireGuard:
		return "WireGuard"
	case TypeWARP: // https://github.com/shtorm-7/sing-box-extended
		return "WARP"
	case TypeMASQUE: // https://github.com/shtorm-7/sing-box-extended
		return "MASQUE"
	case TypeHysteria:
		return "Hysteria"
	case TypeTor:
		return "Tor"
	case TypeSSH:
		return "SSH"
	case TypeShadowTLS:
		return "ShadowTLS"
	case TypeShadowsocksR:
		return "ShadowsocksR"
	case TypeVLESS:
		return "VLESS"
	case TypeTUIC:
		return "TUIC"
	case TypeHysteria2:
		return "Hysteria2"
	case TypeMieru: //karing https://github.com/enfein/mbox
		return "Mieru"
	case TypeAnyTLS:
		return "AnyTLS"
	case TypeSudoku: // https://github.com/shtorm-7/sing-box-extended
		return "Sudoku"
	case TypeOpenConnect:
		return "OpenConnect"
	case TypeOpenVPNClient:
		return "OpenVPN Client"
	case TypeOpenVPNServer:
		return "OpenVPN Server"
	case TypeTailscale:
		return "Tailscale"
	case TypeCloudflared:
		return "Cloudflared"
	case TypeSelector:
		return "Selector"
	case TypeURLTest:
		return "URLTest"
	case TypeVPNClient: // https://github.com/shtorm-7/sing-box-extended
		return "VPN Client"
	case TypeVPNServer: // https://github.com/shtorm-7/sing-box-extended
		return "VPN Server"
	default:
		return "Unknown"
	}
}
