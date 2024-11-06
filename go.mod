module github.com/sagernet/sing-box

go 1.23.2

require (
	berty.tech/go-libtor v1.0.385
	github.com/caddyserver/certmagic v0.20.0
	github.com/cloudflare/circl v1.5.0
	github.com/cretz/bine v0.2.0
	github.com/fsnotify/fsnotify v1.8.0
	github.com/go-chi/chi/v5 v5.1.0
	github.com/go-chi/cors v1.2.1
	github.com/go-chi/render v1.0.3
	github.com/gofrs/uuid/v5 v5.3.0
	github.com/insomniacslk/dhcp v0.0.0-20231206064809-8c70d406f6d2
	github.com/libdns/alidns v1.0.3
	github.com/libdns/cloudflare v0.1.1
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/metacubex/tfo-go v0.0.0-20241006021335-daedaf0ca7aa
	github.com/mholt/acmez v1.2.0
	github.com/miekg/dns v1.1.62
	github.com/ooni/go-libtor v1.1.8
	github.com/oschwald/maxminddb-golang v1.13.1
	github.com/sagernet/bbolt v0.0.0-20231014093535-ea5cb2fe9f0a
	github.com/sagernet/cloudflare-tls v0.0.0-20231208171750-a4483c1b7cd1
	github.com/sagernet/gomobile v0.1.4
	github.com/sagernet/gvisor v0.0.0-20240428053021-e691de28565f
	github.com/sagernet/quic-go v0.47.0
	github.com/sagernet/reality v0.0.0-20230406110435-ee17307e7691
	github.com/sagernet/sing v0.4.3
	github.com/sagernet/sing-dns v0.2.3
	github.com/sagernet/sing-mux v0.2.0
	github.com/sagernet/sing-quic v0.2.2
	github.com/sagernet/sing-shadowsocks v0.2.7
	github.com/sagernet/sing-shadowsocks2 v0.2.0
	github.com/sagernet/sing-shadowtls v0.1.4
	github.com/sagernet/sing-tun v0.3.3
	github.com/sagernet/sing-vmess v0.1.12
	github.com/sagernet/smux v0.0.0-20231208180855-7041f6ea79e7
	github.com/sagernet/utls v1.6.7
	github.com/sagernet/wireguard-go v0.0.0-20231215174105-89dec3b2f3e8
	github.com/sagernet/ws v0.0.0-20231204124109-acfe8907c854
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.9.0
	go.uber.org/zap v1.27.0
	go4.org/netipx v0.0.0-20231129151722-fdeea329fbba
	golang.org/x/crypto v0.28.0
	golang.org/x/net v0.30.0
	golang.org/x/sys v0.26.0
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
	github.com/xtls/xray-core v1.8.24
)

require ( //karing
	github.com/Dreamacro/clash v1.18.0
	github.com/alitto/pond v1.9.2
	github.com/getsentry/sentry-go v0.29.1
	github.com/valyala/fastjson v1.6.4
)

require (
	github.com/dgryski/go-metro v0.0.0-20211217172704-adc40b04c140 // indirect
	github.com/francoispqt/gojay v1.2.13 // indirect
	github.com/ghodss/yaml v1.0.1-0.20220118164431-d8423dcdf344 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/quic-go/quic-go v0.46.0 // indirect
	github.com/refraction-networking/utls v1.6.7 // indirect
	github.com/riobard/go-bloom v0.0.0-20200614022211-cdc8013cb5b3 // indirect
	github.com/seiflotfy/cuckoofilter v0.0.0-20240715131351-a2f2c23f1771 // indirect
	github.com/v2fly/ss-bloomring v0.0.0-20210312155135-28617310f63e // indirect
	github.com/wlynxg/anet v0.0.3 // indirect
	github.com/xtls/reality v0.0.0-20240712055506-48f0b2d5ed6d // indirect
	go.uber.org/mock v0.4.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

//go get github.com/KaringX/sing-quic@de3999b0837577a99a891562c776914c8aed3bc1
//replace github.com/sagernet/sing => github.com/KaringX/sing v0.4.2-0.20240626045944-164e86c6147b
//replace github.com/sagernet/sing-tun => ../../KaringX/sing-tun
//replace github.com/sagernet/sing-quic => ../../KaringX/sing-quic
//replace github.com/sagernet/sing-dns => ../../KaringX/sing-dns
//replace github.com/sagernet/sing-tun => github.com/KaringX/sing-tun v0.3.3-0.20240808075023-e16a492e752c
//replace github.com/caddyserver/certmagic => github.com/caddyserver/certmagic v0.20.0
replace github.com/sagernet/sing => github.com/KaringX/sing v0.4.2-0.20241106071531-3533484124be
//replace github.com/sagernet/gvisor => github.com/sagernet/gvisor v0.0.0-20240428053021-e691de28565f

replace github.com/sagernet/sing-dns => github.com/KaringX/sing-dns v0.2.4-0.20240912090223-8c33bbae1bb5

replace github.com/sagernet/sing-quic => github.com/KaringX/sing-quic v0.2.0-beta.12.0.20240912075141-e76855c8573c

replace github.com/sagernet/wireguard-go => github.com/hiddify/wireguard-go v0.0.0-20240727191222-383c1da14ff1 //hiddify

replace github.com/xtls/xray-core => github.com/hiddify/xray-core v0.0.0-20240902024714-0fcb0895bb4b //hiddify

require (
	github.com/Dreamacro/protobytes v0.0.0-20230911123819-0bbf144b9b9a // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/pprof v0.0.0-20241029153458-d1b30febd7db // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/libdns/libdns v0.2.2 // indirect
	github.com/onsi/ginkgo/v2 v2.21.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pion/dtls/v2 v2.2.12 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/stun/v2 v2.0.0 // indirect
	github.com/pion/transport/v2 v2.2.10 // indirect
	github.com/pion/transport/v3 v3.0.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/quic-go/qtls-go1-20 v0.4.1 // indirect
	github.com/sagernet/netlink v0.0.0-20240916134442-83396419aa8b // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/u-root/uio v0.0.0-20230220225925-ffce2a382923 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20241009180824-f66d83c29e7c // indirect; indirectdirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241021214115-324edc3d5d38 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/blake3 v1.3.0 // indirect
)
