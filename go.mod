module github.com/sagernet/sing-box

go 1.23.2

require (
	github.com/caddyserver/certmagic v0.20.0
	github.com/cloudflare/circl v1.5.0
	github.com/cretz/bine v0.2.0
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/render v1.0.3
	github.com/gofrs/uuid/v5 v5.3.0
	github.com/insomniacslk/dhcp v0.0.0-20250109001534-8abf58130905
	github.com/libdns/alidns v1.0.3
	github.com/libdns/cloudflare v0.1.1
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/metacubex/tfo-go v0.0.0-20241231083714-66613d49c422
	github.com/mholt/acmez v1.2.0
	github.com/miekg/dns v1.1.63
	github.com/oschwald/maxminddb-golang v1.12.0
	github.com/sagernet/asc-go v0.0.0-20241217030726-d563060fe4e1
	github.com/sagernet/bbolt v0.0.0-20231014093535-ea5cb2fe9f0a
	github.com/sagernet/cloudflare-tls v0.0.0-20231208171750-a4483c1b7cd1
	github.com/sagernet/cors v1.2.1
	github.com/sagernet/fswatch v0.1.1
	github.com/sagernet/gomobile v0.1.4
	github.com/sagernet/gvisor v0.0.0-20241123041152-536d05261cff
	github.com/sagernet/quic-go v0.49.0-beta.1
	github.com/sagernet/reality v0.0.0-20230406110435-ee17307e7691
	github.com/sagernet/sing v0.6.1
	github.com/sagernet/sing-dns v0.4.0
	github.com/sagernet/sing-mux v0.3.1
	github.com/sagernet/sing-quic v0.4.0
	github.com/sagernet/sing-shadowsocks v0.2.7
	github.com/sagernet/sing-shadowsocks2 v0.2.0
	github.com/sagernet/sing-shadowtls v0.2.0
	github.com/sagernet/sing-tun v0.6.1
	github.com/sagernet/sing-vmess v0.2.0
	github.com/sagernet/smux v0.0.0-20231208180855-7041f6ea79e7
	github.com/sagernet/utls v1.6.7
	github.com/sagernet/wireguard-go v0.0.1-beta.5
	github.com/sagernet/ws v0.0.0-20231204124109-acfe8907c854
	github.com/spf13/cobra v1.8.1
	github.com/stretchr/testify v1.10.0
	go.uber.org/zap v1.27.0
	go4.org/netipx v0.0.0-20231129151722-fdeea329fbba
	golang.org/x/crypto v0.32.0
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56
	golang.org/x/mod v0.20.0
	golang.org/x/net v0.34.0
	golang.org/x/sys v0.30.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20230429144221-925a1e7659e6
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.35.1
	howett.net/plist v1.0.1
)

require ( //hiddify
	github.com/pion/logging v0.2.2
	github.com/pion/turn/v3 v3.0.3
	github.com/pires/go-proxyproto v0.8.0
//github.com/sagernet/tfo-go v0.0.0-20230816093905-5a5c285d44a6
)

require ( //karing
	github.com/Dreamacro/clash v1.18.0
	github.com/alitto/pond v1.9.2
	github.com/getsentry/sentry-go v0.29.1
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/valyala/fastjson v1.6.4
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

replace github.com/Dreamacro/clash => github.com/KaringX/clash v0.0.0-20241101044429-f71df89d4091

//go get github.com/KaringX/sing-quic@de3999b0837577a99a891562c776914c8aed3bc1
replace github.com/sagernet/sing => github.com/KaringX/sing v0.5.0-alpha.11.0.20250223034922-ddf34fb62560 //karing_v0.6.1

//replace github.com/sagernet/sing => ../../KaringX/sing

replace github.com/sagernet/sing-dns => github.com/KaringX/sing-dns v0.3.0-beta.14.0.20250223034424-bed5e3efc69b //karing_v0.4.0

//replace github.com/sagernet/sing-dns => ../../KaringX/sing-dns

replace github.com/sagernet/sing-quic => github.com/KaringX/sing-quic v0.2.0-beta.12.0.20250211065213-b48dcc4767ef //karing_v0.4.0

//replace github.com/sagernet/sing-quic => ../../KaringX/sing-quic

replace github.com/sagernet/sing-tun => github.com/KaringX/sing-tun v0.6.2-0.20250224122437-2421005a71b2 //karing_v0.6.1

//replace github.com/sagernet/sing-tun => ../../KaringX/sing-tun

replace github.com/sagernet/wireguard-go => github.com/KaringX/wireguard-go v0.0.1-beta.5.0.20250223050301-4b516c388c7f //karing_v0.0.1-beta.5

//replace github.com/sagernet/wireguard-go => ../../KaringX/wireguard-go

require (
	github.com/Dreamacro/protobytes v0.0.0-20230911123819-0bbf144b9b9a // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/libdns/libdns v0.2.2 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mdlayher/netlink v1.7.2 // indirect
	github.com/mdlayher/socket v0.4.1 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pierrec/lz4/v4 v4.1.14 // indirect
	github.com/pion/dtls/v2 v2.2.7 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/stun/v2 v2.0.0 // indirect
	github.com/pion/transport/v2 v2.2.1 // indirect
	github.com/pion/transport/v3 v3.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/quic-go/qtls-go1-20 v0.4.1 // indirect
	github.com/sagernet/netlink v0.0.0-20240612041022-b9a21c07ac6a // indirect
	github.com/sagernet/nftables v0.3.0-beta.4 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/u-root/uio v0.0.0-20230220225925-ffce2a382923 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/blake3 v0.2.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/blake3 v1.3.0 // indirect
)
