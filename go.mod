module github.com/sagernet/sing-box

go 1.25.5

require (
	filippo.io/age v1.3.1
	github.com/anthropics/anthropic-sdk-go v1.26.0
	github.com/anytls/sing-anytls v0.0.11
	github.com/caddyserver/certmagic v0.25.3-0.20260421143802-60d9d8b415d6
	github.com/caddyserver/zerossl v0.1.5
	github.com/coder/websocket v1.8.14
	github.com/creack/pty v1.1.24
	github.com/cretz/bine v0.2.0
	github.com/database64128/tfo-go/v2 v2.3.2
	github.com/dblohm7/wingoes v0.0.0-20240119213807-a09d6be7affa
	github.com/go-chi/chi/v5 v5.2.5
	github.com/go-chi/render v1.0.3
	github.com/godbus/dbus/v5 v5.2.2
	github.com/gofrs/uuid/v5 v5.5.1
	github.com/insomniacslk/dhcp v0.0.0-20260220084031-5adc3eb26f91
	github.com/jsimonetti/rtnetlink v1.4.1
	github.com/keybase/go-keychain v0.0.1
	github.com/libdns/acmedns v0.5.0
	github.com/libdns/alidns v1.0.6
	github.com/libdns/cloudflare v0.2.2
	github.com/libdns/libdns v1.1.1
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mattn/go-runewidth v0.0.27
	github.com/mdlayher/netlink v1.11.2
	github.com/metacubex/utls v1.8.7
	github.com/mholt/acmez/v3 v3.1.6
	github.com/miekg/dns v1.1.72
	github.com/openai/openai-go/v3 v3.26.0
	github.com/oschwald/maxminddb-golang v1.13.1
	github.com/pkg/sftp v1.13.10
	github.com/sagernet/asc-go v0.0.0-20260827085112-8f8226245b0d
	github.com/sagernet/bbolt v0.0.0-20260823094646-e24805439c9c
	github.com/sagernet/cors v1.2.1
	github.com/sagernet/cronet-go v0.0.0-20260831031307-45832ab07484
	github.com/sagernet/cronet-go/all v0.0.0-20260831031307-45832ab07484
	github.com/sagernet/fswatch v0.1.2
	github.com/sagernet/gliderssh v0.3.4-0.20260531100337-2194faca5648
	github.com/sagernet/gomobile v0.1.12
	github.com/sagernet/gvisor v0.0.0-20260727.0-sing-box-mod.1
	github.com/sagernet/netlink v0.0.0-20260814022025-64455d367bbf
	github.com/sagernet/nftables v0.3.0-mod.4
	github.com/sagernet/quic-go v0.61.0-sing-box-mod.7
	github.com/sagernet/sing v0.9.0-beta.4
	github.com/sagernet/sing-cloudflared v0.1.3-0.20260706062323-d9787e794aa3
	github.com/sagernet/sing-mux v0.3.5
	github.com/sagernet/sing-openconnect v0.0.0-20260810065514-53aa8058f8df
	github.com/sagernet/sing-openvpn v0.0.0-20260729104525-103eb5fe5eb6
	github.com/sagernet/sing-quic v0.7.0-beta.4
	github.com/sagernet/sing-shadowsocks v0.2.8
	github.com/sagernet/sing-shadowsocks2 v0.2.1
	github.com/sagernet/sing-shadowtls v0.2.1
	github.com/sagernet/sing-snell v0.0.0-20260829071736-20f2eaec77c3
	github.com/sagernet/sing-tun v0.9.0-beta.4
	github.com/sagernet/sing-usbip v0.0.0-20260817040617-28bd42667eca
	github.com/sagernet/sing-vmess v0.2.8-0.20250909125414-3aed155119a1
	github.com/sagernet/smux v1.5.50-sing-box-mod.1
	github.com/sagernet/tailscale v1.102.1-sing-box-1.14-mod.4
	github.com/sagernet/wireguard-go v0.0.5-0.20260823125007-8bd032a91a30
	github.com/sagernet/ws v0.0.0-20231204124109-acfe8907c854
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.12.0
	github.com/tailscale/go-winio v0.0.0-20231025203758-c4f33415bf55
	github.com/vishvananda/netns v0.0.5
	go.uber.org/zap v1.27.1
	go4.org/netipx v0.0.0-20231129151722-fdeea329fbba
	golang.org/x/crypto v0.54.0
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f
	golang.org/x/mod v0.37.0
	golang.org/x/net v0.57.0
	golang.org/x/sync v0.22.0
	golang.org/x/sys v0.47.0
	golang.org/x/term v0.45.0
	golang.org/x/text v0.40.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20241231184526-a9ab2273dd10
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.11
	howett.net/plist v1.0.1
)

require ( //hiddify
	github.com/pion/logging v0.2.4
	github.com/pion/turn/v3 v3.0.3
	github.com/pires/go-proxyproto v0.11.0
)

require ( //karing
	github.com/alitto/pond v1.9.2
	github.com/enfein/mieru/v3 v3.35.0
	github.com/getsentry/sentry-go v0.48.0
	github.com/klauspost/compress v1.18.6
	github.com/klauspost/cpuid/v2 v2.3.0
	github.com/mattn/go-sqlite3 v1.14.49
	github.com/metacubex/mihomo v1.19.29
	github.com/metacubex/sing v0.5.7
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/valyala/fastjson v1.6.10
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	lukechampine.com/blake3 v1.3.0
)

//replace github.com/metacubex/mihomo => github.com/KaringX/clash v0.0.0-20241101044429-f71df89d4091

//go get github.com/KaringX/sing@a5ccca51fe8a39390591cf9480ffebb9f0c3fc61
replace github.com/sagernet/sing => github.com/KaringX/sing v0.5.0-alpha.11.0.20260819025303-12e7cea73174 //karing_v0.8.13

//replace github.com/sagernet/sing => ../../KaringX/sing

replace github.com/sagernet/sing-dns => github.com/KaringX/sing-dns v0.3.0-beta.14.0.20250530212309-a733af0f5878 //karing_v0.4.5

//replace github.com/sagernet/sing-dns => ../../KaringX/sing-dns

replace github.com/sagernet/sing-quic => github.com/KaringX/sing-quic v0.2.0-beta.12.0.20260805050642-0122883b056b //karing_v0.6.4

//replace github.com/sagernet/sing-quic => ../../KaringX/sing-quic

replace github.com/sagernet/quic-go => github.com/KaringX/quic-go v0.49.0-beta.1.0.20260319073656-e535214ac88c //karing_v0.59.0-sing-box-mod.4

//replace github.com/sagernet/quic-go => ../../KaringX/quic-go

replace github.com/sagernet/sing-tun => github.com/KaringX/sing-tun v0.6.2-0.20260819025838-60cad1147202 //karing_v0.8.12

//replace github.com/sagernet/sing-tun => ../../KaringX/sing-tun

replace github.com/sagernet/ws => github.com/KaringX/ws v0.0.0-20250301004503-082f57b5c845 //karing_fork

//replace github.com/sagernet/ws => ../../KaringX/ws

replace github.com/sagernet/wireguard-go => github.com/KaringX/wireguard-go v0.0.1-beta.5.0.20260812025805-7d2d140c5113 //karing_v0.0.4

//replace github.com/sagernet/wireguard-go => ../../KaringX/wireguard-go

replace github.com/sagernet/tailscale => github.com/KaringX/tailscale v1.80.3-mod.5.0.20260812023457-7007033a39f9 //v1.92.4-sing-box-1.13-mod.9

//replace github.com/sagernet/tailscale => ../../KaringX/tailscale

replace github.com/sagernet/sing-vmess => github.com/KaringX/sing-vmess v0.2.8-0.20260526023644-9a9b48357eaf //v0.2.8-0

//replace github.com/sagernet/sing-vmess => ../../KaringX/sing-vmess

replace github.com/metacubex/sing-shadowsocks2 => github.com/Restia-Ashbell/sing-shadowsocks2 v0.0.0-20250606153500-31ef7b14aaa6

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/ajg/form v1.7.1 // indirect
	github.com/akutz/memconn v0.1.0 // indirect
	github.com/alexbrainman/sspi v0.0.0-20231016080023-1a75b4708caa // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/caddyserver/zerossl v0.1.5 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/coreos/go-iptables v0.8.0 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	filippo.io/hpke v0.4.0 // indirect
	github.com/RyuaNerin/go-krypto v1.3.0 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/akutz/memconn v0.1.0 // indirect
	github.com/alexbrainman/sspi v0.0.0-20231016080023-1a75b4708caa // indirect
	github.com/anchore/go-lzo v0.1.0 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/axiomhq/hyperloglog v0.0.0-20240319100328-84253e514e02 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.2.0 // indirect
	github.com/coreos/go-iptables v0.7.1-0.20240112124308-65c67c9f46e6 // indirect
	github.com/coreos/go-oidc/v3 v3.17.0 // indirect
	github.com/database64128/netx-go v0.1.1 // indirect
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1 // indirect
	github.com/dgryski/go-camellia v0.0.0-20191119043421-69a8a13fb23d // indirect
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc // indirect
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/florianl/go-nfqueue/v2 v2.1.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/gaissmai/bart v0.18.0 // indirect
	github.com/go-json-experiment/json v0.0.0-20250813024750-ebf49471dced // indirect

	github.com/gaissmai/bart v0.26.1 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect

	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go4org/hashtriemap v0.0.0-20251130024219-545ba229f689 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/certificate-transparency-go v1.3.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/google/nftables v0.2.1-0.20240414091927-5e242ec57806 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/hdevalence/ed25519consensus v0.2.0 // indirect
	github.com/huin/goupnp v1.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect

	github.com/jsimonetti/rtnetlink v1.4.0 // indirect
	github.com/libdns/libdns v1.1.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mdlayher/netlink v1.9.0 // indirect
	github.com/mdlayher/socket v0.5.1 // indirect
	github.com/metacubex/chacha v0.1.5 // indirect
	github.com/metacubex/randv2 v0.2.0 // indirect
	github.com/mitchellh/go-ps v1.0.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.27 // indirect
	github.com/pion/dtls/v2 v2.2.7 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/stun/v2 v2.0.0 // indirect
	github.com/pion/transport/v2 v2.2.1 // indirect
	github.com/pion/transport/v3 v3.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus-community/pro-bing v0.4.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/safchain/ethtool v0.3.0 // indirect
	github.com/sagernet/cronet-go/lib/android_386 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/android_amd64 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/android_arm v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/android_arm64 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/darwin_amd64 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/darwin_arm64 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/ios_amd64_simulator v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/ios_arm64 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/ios_arm64_simulator v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_386 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_386_musl v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_amd64 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_amd64_musl v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_arm v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_arm64 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_arm64_musl v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_arm_musl v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_loong64 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_loong64_musl v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_mips64le v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_mipsle v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_mipsle_musl v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_riscv64 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/linux_riscv64_musl v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/tvos_amd64_simulator v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/tvos_arm64 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/tvos_arm64_simulator v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/windows_amd64 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/cronet-go/lib/windows_arm64 v0.0.0-20260807161529-8d42107dcdfc // indirect
	github.com/sagernet/netlink v0.0.0-20240612041022-b9a21c07ac6a // indirect
	github.com/sagernet/nftables v0.3.0-mod.2 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/tailscale/certstore v0.1.1-0.20260409135935-3638fb84b77d // indirect
	github.com/tailscale/go-winio v0.0.0-20231025203758-c4f33415bf55 // indirect
	github.com/tailscale/goupnp v1.0.1-0.20210804011211-c64d0f06ea05 // indirect
	github.com/tailscale/hujson v0.0.0-20221223112325-20486734a56a // indirect

	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/klauspost/compress v1.19.1 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/koron/go-ssdp v0.0.4 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/libp2p/go-nat v1.0.1-0.20250821073202-01afc089f138 // indirect
	github.com/libp2p/go-netroute v0.2.1 // indirect
	github.com/mdlayher/socket v0.6.0 // indirect
	github.com/mitchellh/go-ps v1.0.0 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.26 // indirect
	github.com/pion/dtls/v3 v3.1.5 // indirect
	github.com/pion/logging v0.2.4 // indirect
	github.com/pion/transport/v4 v4.0.2 // indirect
	github.com/pires/go-proxyproto v0.8.1 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/safchain/ethtool v0.3.0 // indirect
	github.com/sagernet/cronet-go/lib/android_386 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/android_amd64 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/android_arm v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/android_arm64 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/darwin_amd64 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/darwin_arm64 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/ios_amd64_simulator v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/ios_arm64 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/ios_arm64_simulator v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_386 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_386_musl v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_amd64 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_amd64_musl v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_arm v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_arm64 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_arm64_musl v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_arm_musl v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_loong64 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_loong64_musl v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_mips64le v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_mipsle v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_mipsle_musl v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_riscv64 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/linux_riscv64_musl v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/tvos_amd64_simulator v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/tvos_arm64 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/tvos_arm64_simulator v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/windows_amd64 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/sagernet/cronet-go/lib/windows_arm64 v0.0.0-20260831030607-f80ef37265e5 // indirect
	github.com/smallstep/pkcs7 v0.1.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/tailscale/certstore v0.1.1-0.20260409135935-3638fb84b77d // indirect
	github.com/tailscale/hujson v0.0.0-20260302212456-ecc657c15afd // indirect

	github.com/tailscale/netlink v1.1.1-0.20240822203006-4d49adab4de7 // indirect
	github.com/tailscale/peercred v0.0.0-20250107143737-35a0c7bd7edc // indirect
	github.com/tailscale/web-client-prebuilt v0.0.0-20250124233751-d4cd19a26976 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect

	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/u-root/uio v0.0.0-20240224005618-d2acac8f3701 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect

	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/u-root/uio v0.0.0-20240224005618-d2acac8f3701 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect

	github.com/zeebo/blake3 v0.2.4 // indirect
	gitlab.com/go-extension/aes-ccm v0.0.0-20230221065045-e58665ef23c7 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap/exp v0.3.0 // indirect
	go4.org/mem v0.0.0-20240501181205-ae6ca9944745 // indirect

	golang.org/x/oauth2 v0.34.0 // indirect
	golang.org/x/term v0.45.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	golang.org/x/time v0.11.0 // indirect

	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/time v0.15.0 // indirect

	golang.org/x/tools v0.47.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	golang.zx2c4.com/wireguard/windows v0.5.3 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect

	lukechampine.com/blake3 v1.3.0 // indirect
	zombiezen.com/go/capnproto2 v2.18.2+incompatible // indirect
)
