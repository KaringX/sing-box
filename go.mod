module github.com/sagernet/sing-box

go 1.23.2

toolchain go1.24.3

require (
	github.com/anytls/sing-anytls v0.0.8
	github.com/caddyserver/certmagic v0.23.0
	github.com/coder/websocket v1.8.13
	github.com/cretz/bine v0.2.0
	github.com/go-chi/chi/v5 v5.2.2
	github.com/go-chi/render v1.0.3
	github.com/godbus/dbus/v5 v5.1.1-0.20230522191255-76236955d466
	github.com/gofrs/uuid/v5 v5.3.2
	github.com/insomniacslk/dhcp v0.0.0-20250417080101-5f8cf70e8c5f
	github.com/libdns/alidns v1.0.5-libdns.v1.beta1
	github.com/libdns/cloudflare v0.2.2-0.20250708034226-c574dccb31a6
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/metacubex/tfo-go v0.0.0-20250516165257-e29c16ae41d4
	github.com/metacubex/utls v1.8.0
	github.com/mholt/acmez/v3 v3.1.2
	github.com/miekg/dns v1.1.67
	github.com/oschwald/maxminddb-golang v1.13.1
	github.com/sagernet/asc-go v0.0.0-20241217030726-d563060fe4e1
	github.com/sagernet/bbolt v0.0.0-20231014093535-ea5cb2fe9f0a
	github.com/sagernet/cors v1.2.1
	github.com/sagernet/fswatch v0.1.1
	github.com/sagernet/gomobile v0.1.8
	github.com/sagernet/gvisor v0.0.0-20250325023245-7a9c0f5725fb
	github.com/sagernet/quic-go v0.52.0-beta.1
	github.com/sagernet/sing v0.7.10
	github.com/sagernet/sing-mux v0.3.3
	github.com/sagernet/sing-quic v0.5.2-0.20250909083218-00a55617c0fb
	github.com/sagernet/sing-shadowsocks v0.2.8
	github.com/sagernet/sing-shadowsocks2 v0.2.1
	github.com/sagernet/sing-shadowtls v0.2.1-0.20250503051639-fcd445d33c11
	github.com/sagernet/sing-tun v0.7.2
	github.com/sagernet/sing-vmess v0.2.7
	github.com/sagernet/smux v1.5.34-mod.2
	github.com/sagernet/tailscale v1.80.3-sing-box-1.12-mod.1
	github.com/sagernet/wireguard-go v0.0.1-beta.7
	github.com/sagernet/ws v0.0.0-20231204124109-acfe8907c854
	github.com/spf13/cobra v1.9.1
	github.com/stretchr/testify v1.10.0
	github.com/vishvananda/netns v0.0.5
	go.uber.org/zap v1.27.0
	go4.org/netipx v0.0.0-20231129151722-fdeea329fbba
	golang.org/x/crypto v0.41.0
	golang.org/x/exp v0.0.0-20250506013437-ce4c2cf36ca6
	golang.org/x/mod v0.27.0
	golang.org/x/net v0.43.0
	golang.org/x/sys v0.35.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20241231184526-a9ab2273dd10
	google.golang.org/grpc v1.73.0
	google.golang.org/protobuf v1.36.6
	howett.net/plist v1.0.1
)

require ( //hiddify
	github.com/pion/logging v0.2.2
	github.com/pion/turn/v3 v3.0.3
	github.com/pires/go-proxyproto v0.8.0
	github.com/valyala/fastjson v1.6.4
//github.com/sagernet/tfo-go v0.0.0-20230816093905-5a5c285d44a6
)

require ( //karing
	github.com/Dreamacro/clash v1.18.0
	github.com/alitto/pond v1.9.2
	github.com/enfein/mieru/v3 v3.19.1
	github.com/getsentry/sentry-go v0.33.0
	github.com/klauspost/compress v1.17.11
	github.com/shirou/gopsutil/v3 v3.24.5
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

replace github.com/Dreamacro/clash => github.com/KaringX/clash v0.0.0-20241101044429-f71df89d4091

//go get github.com/KaringX/sing@50a9c2efd67210376bb718ef766d53710768d5ae
replace github.com/sagernet/sing => github.com/KaringX/sing v0.5.0-alpha.11.0.20250904033655-da7ae65a160b //karing_v0.7.6

//replace github.com/sagernet/sing => ../../KaringX/sing

replace github.com/sagernet/sing-dns => github.com/KaringX/sing-dns v0.3.0-beta.14.0.20250530212309-a733af0f5878 //karing_v0.4.5

//replace github.com/sagernet/sing-dns => ../../KaringX/sing-dns

replace github.com/sagernet/sing-quic => github.com/KaringX/sing-quic v0.2.0-beta.12.0.20250826053421-b88d178eb45c //karing_v0.5.0

//replace github.com/sagernet/sing-quic => ../../KaringX/sing-quic

replace github.com/sagernet/quic-go => github.com/KaringX/quic-go v0.49.0-beta.1.0.20250528041323-b571ce42d479 //karing_v0.52.0-beta1

//replace github.com/sagernet/quic-go => ../../KaringX/quic-go

replace github.com/sagernet/sing-tun => github.com/KaringX/sing-tun v0.6.2-0.20250904033100-bd6aec0397b1 //karing_v0.7.0

//replace github.com/sagernet/sing-tun => ../../KaringX/sing-tun

replace github.com/sagernet/ws => github.com/KaringX/ws v0.0.0-20250301004503-082f57b5c845 //karing_fork

//replace github.com/sagernet/ws => ../../KaringX/ws

replace github.com/sagernet/wireguard-go => github.com/KaringX/wireguard-go v0.0.1-beta.5.0.20250531024932-ebfb2f0492e2 //karing_v0.0.1-beta.7-2

//replace github.com/sagernet/wireguard-go => ../../KaringX/wireguard-go

replace github.com/sagernet/tailscale => github.com/KaringX/tailscale v1.80.3-mod.5.0.20250531025012-b525f2db9b5f //karing_v1.80.3-mod.5

//replace github.com/sagernet/tailscale => ../../KaringX/tailscale

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Dreamacro/protobytes v0.0.0-20230617041236-6500a9f4f158 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/akutz/memconn v0.1.0 // indirect
	github.com/alexbrainman/sspi v0.0.0-20231016080023-1a75b4708caa // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/bits-and-blooms/bitset v1.13.0 // indirect
	github.com/caddyserver/zerossl v0.1.3 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/coreos/go-iptables v0.7.1-0.20240112124308-65c67c9f46e6 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dblohm7/wingoes v0.0.0-20240119213807-a09d6be7affa // indirect
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1 // indirect
	github.com/digitalocean/go-smbios v0.0.0-20180907143718-390a4f403a8e // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/gaissmai/bart v0.11.1 // indirect
	github.com/go-json-experiment/json v0.0.0-20250103232110-6a9a0fde9288 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/nftables v0.2.1-0.20240414091927-5e242ec57806 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/csrf v1.7.3-0.20250123201450-9dd6af1f6d30 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/hdevalence/ed25519consensus v0.2.0 // indirect
	github.com/illarion/gonotify/v2 v2.0.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jsimonetti/rtnetlink v1.4.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/kortschak/wol v0.0.0-20200729010619-da482cc4850a // indirect
	github.com/libdns/libdns v1.1.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mdlayher/genetlink v1.3.2 // indirect
	github.com/mdlayher/netlink v1.7.3-0.20250113171957-fbb4dce95f42 // indirect
	github.com/mdlayher/sdnotify v1.0.0 // indirect
	github.com/mdlayher/socket v0.5.1 // indirect
	github.com/mitchellh/go-ps v1.0.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pion/dtls/v2 v2.2.7 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/stun/v2 v2.0.0 // indirect
	github.com/pion/transport/v2 v2.2.1 // indirect
	github.com/pion/transport/v3 v3.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus-community/pro-bing v0.4.0 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/safchain/ethtool v0.3.0 // indirect
	github.com/sagernet/netlink v0.0.0-20240612041022-b9a21c07ac6a // indirect
	github.com/sagernet/nftables v0.3.0-beta.4 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/tailscale/certstore v0.1.1-0.20231202035212-d3fa0460f47e // indirect
	github.com/tailscale/go-winio v0.0.0-20231025203758-c4f33415bf55 // indirect
	github.com/tailscale/golang-x-crypto v0.0.0-20240604161659-3fde5e568aa4 // indirect
	github.com/tailscale/goupnp v1.0.1-0.20210804011211-c64d0f06ea05 // indirect
	github.com/tailscale/hujson v0.0.0-20221223112325-20486734a56a // indirect
	github.com/tailscale/netlink v1.1.1-0.20240822203006-4d49adab4de7 // indirect
	github.com/tailscale/peercred v0.0.0-20250107143737-35a0c7bd7edc // indirect
	github.com/tailscale/web-client-prebuilt v0.0.0-20250124233751-d4cd19a26976 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/u-root/uio v0.0.0-20240224005618-d2acac8f3701 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap/exp v0.3.0 // indirect
	go4.org/mem v0.0.0-20240501181205-ae6ca9944745 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/term v0.34.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	golang.zx2c4.com/wireguard/windows v0.5.3 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/blake3 v1.3.0 // indirect
)
