package option

//karing
import (
	"net/netip"
	"time"

	mDNS "github.com/miekg/dns"
	C "github.com/sagernet/sing-box/constant"
	dns "github.com/sagernet/sing-dns"
	"github.com/sagernet/sing/common/auth"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/valyala/fastjson"
	"github.com/xtls/xray-core/infra/conf"
)

func unmarshalFastJSONDuration(interval string) Duration {
	if len(interval) == 0 {
		return 0
	}
	duration, err := time.ParseDuration(interval)
	if err == nil {
		return Duration(duration)
	}
	return 0
}

func unmarshalFastJSONDomainStrategy(strategy string) DomainStrategy {
	switch strategy {
	case "", "as_is":
		return DomainStrategy(dns.DomainStrategyAsIS)
	case "prefer_ipv4":
		return DomainStrategy(dns.DomainStrategyPreferIPv4)
	case "prefer_ipv6":
		return DomainStrategy(dns.DomainStrategyPreferIPv6)
	case "ipv4_only":
		return DomainStrategy(dns.DomainStrategyUseIPv4)
	case "ipv6_only":
		return DomainStrategy(dns.DomainStrategyUseIPv6)
	default:
		return DomainStrategy(dns.DomainStrategyPreferIPv4)
	}
}

func unmarshalFastJSONMapHTTPHeader(fj *fastjson.Object) HTTPHeader {
	if fj == nil {
		return nil
	}

	list := make(HTTPHeader, fj.Len())
	fj.Visit(func(key []byte, value *fastjson.Value) {
		list[string(key)] = unmarshalFastJSONArrayString(value)
	})

	return list
}

// clash.go
func (o *ClashAPIOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.ExternalController = stringNotNil(fj.GetStringBytes("external_controller"))
	o.ExternalUI = stringNotNil(fj.GetStringBytes("external_ui"))
	o.ExternalUIDownloadURL = stringNotNil(fj.GetStringBytes("external_ui_download_url"))
	o.ExternalUIDownloadDetour = stringNotNil(fj.GetStringBytes("external_ui_download_detour"))
	o.Secret = stringNotNil(fj.GetStringBytes("secret"))
	o.DefaultMode = stringNotNil(fj.GetStringBytes("default_mode"))
}

func (o *SelectorOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Outbounds = unmarshalFastJSONArrayStringWithName(fj, "outbounds")
	o.Default = stringNotNil(fj.GetStringBytes("default"))
	o.InterruptExistConnections = fj.GetBool("interrupt_exist_connections")
}

func (o *URLTestOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Outbounds = unmarshalFastJSONArrayStringWithName(fj, "outbounds")
	o.URL = stringNotNil(fj.GetStringBytes("url"))
	o.Interval = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("interval")))
	o.Tolerance = uint16(fj.GetInt("tolerance"))
	o.IdleTimeout = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("idle_timeout")))
	o.InterruptExistConnections = fj.GetBool("interrupt_exist_connections")
	o.Default = stringNotNil(fj.GetStringBytes("default"))
	o.ReTestIfNetworkUpdate = fj.GetBool("retest_if_network_udpate")
}

// config.go
func (o *Options) UnmarshalFastJSON(content []byte) error {
	var parser fastjson.Parser
	jp, err := parser.ParseBytes(content)
	if err != nil {
		return err
	}

	o.unmarshalFastJSON(jp)
	return nil
}
func (o *Options) unmarshalFastJSON(fj *fastjson.Value) {
	o.Schema = stringNotNil(fj.GetStringBytes("schema"))
	log := fj.Get("log")
	if log != nil && log.Type() != fastjson.TypeNull {
		o.Log = &LogOptions{}
		o.Log.unmarshalFastJSON(log)
	}
	dns := fj.Get("dns")
	if dns != nil && dns.Type() != fastjson.TypeNull {
		o.DNS = &DNSOptions{}
		o.DNS.unmarshalFastJSON(dns)
	}
	ntp := fj.Get("ntp")
	if ntp != nil && ntp.Type() != fastjson.TypeNull {
		o.NTP = &NTPOptions{}
		o.NTP.unmarshalFastJSON(ntp)
	}
	o.Inbounds = unmarshalFastJSONArrayInbound(fj.Get("inbounds"))
	o.Outbounds = unmarshalFastJSONArrayOutbound(fj.Get("outbounds"))
	route := fj.Get("route")
	if route != nil && route.Type() != fastjson.TypeNull {
		o.Route = &RouteOptions{}
		o.Route.unmarshalFastJSON(route)
	}
	experimental := fj.Get("experimental")
	if experimental != nil && experimental.Type() != fastjson.TypeNull {
		o.Experimental = &ExperimentalOptions{}
		o.Experimental.unmarshalFastJSON(experimental)
	}
}

func (o *LogOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Disabled = fj.GetBool("disabled")
	o.Level = stringNotNil(fj.GetStringBytes("level"))
	o.Output = stringNotNil(fj.GetStringBytes("output"))
	o.Timestamp = fj.GetBool("timestamp")
	//o.DisableColor = fj.GetBool("-")
}

func stringNotNil(v []byte) string {
	if v == nil {
		return ""
	}
	return string(v)
}
func unmarshalFastJSONMapStringString(fj *fastjson.Object) map[string][]string {
	if fj == nil {
		return nil
	}

	protoMap := make(map[string][]string)
	fj.Visit(func(key []byte, value *fastjson.Value) {
		protoMap[string(key)] = unmarshalFastJSONArrayString(value)
	})

	return protoMap
}
func unmarshalFastJSONListableString(fj *fastjson.Value) Listable[string] {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make([]string, len(arr))
		for i, v := range arr {
			by, err := v.StringBytes()
			if err == nil {
				list[i] = stringNotNil(by)
			}
		}
		return list
	}
	if fj.Type() != fastjson.TypeString {
		return nil
	}
	list := make([]string, 1)
	list[0] = stringNotNil(fj.GetStringBytes())
	return list
}

func unmarshalFastJSONListableInt(fj *fastjson.Value) Listable[int] {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make(Listable[int], len(arr))
		for i, v := range arr {
			vv, err := v.Int()
			if err == nil {
				list[i] = int(vv)
			}
		}
		return list
	}
	if fj.Type() != fastjson.TypeNumber {
		return nil
	}
	list := make(Listable[int], 1)
	vv, _ := fj.Int()
	list[0] = vv
	return list
}

func unmarshalFastJSONListableInt32(fj *fastjson.Value) Listable[int32] {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}
	arr := fj.GetArray()
	if arr != nil {
		list := make(Listable[int32], len(arr))
		for i, v := range arr {
			vv, err := v.Int()
			if err == nil {
				list[i] = int32(vv)
			}
		}
		return list
	}
	if fj.Type() != fastjson.TypeNumber {
		return nil
	}
	list := make(Listable[int32], 1)
	vv, _ := fj.Int()
	list[0] = int32(vv)
	return list
}

func unmarshalFastJSONListableUInt32(fj *fastjson.Value) Listable[uint32] {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make(Listable[uint32], len(arr))
		for i, v := range arr {
			vv, err := v.Int()
			if err == nil {
				list[i] = uint32(vv)
			}
		}
		return list
	}
	if fj.Type() != fastjson.TypeNumber {
		return nil
	}
	list := make(Listable[uint32], 1)
	vv, _ := fj.Int()
	list[0] = uint32(vv)
	return list
}

func unmarshalFastJSONListableUint16(fj *fastjson.Value) Listable[uint16] {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}
	arr := fj.GetArray()
	if arr != nil {
		list := make(Listable[uint16], len(arr))
		for i, v := range arr {
			vv, err := v.Int()
			if err == nil {
				list[i] = uint16(vv)
			}
		}
		return list
	}
	if fj.Type() != fastjson.TypeNumber {
		return nil
	}
	list := make(Listable[uint16], 1)
	vv, _ := fj.Int()
	list[0] = uint16(vv)
	return list
}

func unmarshalFastJSONListableUint8(fj *fastjson.Value) Listable[uint8] {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make(Listable[uint8], len(arr))
		for i, v := range arr {
			vv, err := v.Int()
			if err == nil {
				list[i] = uint8(vv)
			}
		}
		return list
	}
	if fj.Type() != fastjson.TypeNumber {
		return nil
	}
	list := make(Listable[uint8], 1)
	vv, _ := fj.Int()
	list[0] = uint8(vv)
	return list
}

func unmarshalFastJSONListableDNSQueryType(fj *fastjson.Value) Listable[DNSQueryType] {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make(Listable[DNSQueryType], 0)
		for _, v := range arr {
			by, err := v.StringBytes()
			if err == nil {
				queryType, loaded := mDNS.StringToType[stringNotNil(by)]
				if loaded {
					list = append(list, DNSQueryType(queryType))
				}
			}
		}
		return list
	}
	if fj.Type() != fastjson.TypeString {
		return nil
	}
	list := make(Listable[DNSQueryType], 0)
	by, err := fj.StringBytes()
	if err == nil {
		queryType, loaded := mDNS.StringToType[stringNotNil(by)]
		if loaded {
			list = append(list, DNSQueryType(queryType))
		}
	}
	return list
}
func unmarshalFastJSONArrayString(fj *fastjson.Value) []string {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make([]string, len(arr))
		for i, v := range arr {
			by, err := v.StringBytes()
			if err == nil {
				list[i] = stringNotNil(by)
			}
		}
		return list
	}
	if fj.Type() != fastjson.TypeString {
		return nil
	}
	list := make([]string, 1)
	list[0] = stringNotNil(fj.GetStringBytes())
	return list
}
func unmarshalFastJSONArrayStringWithName(fj *fastjson.Value, name string) []string {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray(name)
	if arr != nil {
		list := make([]string, len(arr))
		for i, v := range arr {
			by, err := v.StringBytes()
			if err == nil {
				list[i] = stringNotNil(by)
			}
		}
		return list
	}
	if fj.Type() != fastjson.TypeString {
		return nil
	}
	list := make([]string, 1)
	list[0] = stringNotNil(fj.GetStringBytes(name))
	return list

}

// debug.go
func (o *DebugOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Listen = stringNotNil(fj.GetStringBytes("listen"))
	gc_percent := fj.Get("gc_percent")
	if gc_percent != nil && gc_percent.Type() != fastjson.TypeNull {
		o.GCPercent = new(int)
		*o.GCPercent = fj.GetInt("gc_percent")
	}
	max_stack := fj.Get("max_stack")
	if max_stack != nil && max_stack.Type() != fastjson.TypeNull {
		o.MaxStack = new(int)
		*o.MaxStack = fj.GetInt("max_stack")
	}
	max_threads := fj.Get("max_threads")
	if max_threads != nil && max_threads.Type() != fastjson.TypeNull {
		o.MaxThreads = new(int)
		*o.MaxThreads = fj.GetInt("max_threads")
	}
	panic_on_fault := fj.Get("panic_on_fault")
	if panic_on_fault != nil && panic_on_fault.Type() != fastjson.TypeNull {
		o.PanicOnFault = new(bool)
		*o.PanicOnFault = fj.GetBool("panic_on_fault")
	}
	o.TraceBack = stringNotNil(fj.GetStringBytes("trace_back"))
	o.MemoryLimit = MemoryBytes(fj.GetInt64("memory_limit"))
	oom_killer := fj.Get("oom_killer")
	if oom_killer != nil && oom_killer.Type() != fastjson.TypeNull {
		o.OOMKiller = new(bool)
		*o.OOMKiller = fj.GetBool("oom_killer")
	}
}

// direct.go
func (o *DirectInboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.ListenOptions.unmarshalFastJSON(fj)
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
	o.OverrideAddress = stringNotNil(fj.GetStringBytes("override_address"))
	o.OverridePort = uint16(fj.GetUint("override_port"))
}
func (o *DirectOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.DialerOptions.unmarshalFastJSON(fj)
	o.OverrideAddress = stringNotNil(fj.GetStringBytes("override_address"))
	o.OverridePort = uint16(fj.GetUint("override_port"))
	o.ProxyProtocol = uint8(fj.GetUint("proxy_protocol"))
}

// dns.go
func (o *DNSOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Servers = unmarshalFastJSONArrayDNSServerOptions(fj.Get("servers"))
	o.Rules = unmarshalFastJSONArrayDNSRule(fj.Get("rules"))
	fakeip := fj.Get("fakeip")
	o.Final = stringNotNil(fj.GetStringBytes("final"))
	o.ReverseMapping = fj.GetBool("reverse_mapping")
	if fakeip != nil && fakeip.Type() != fastjson.TypeNull {
		o.FakeIP = &DNSFakeIPOptions{}
		o.FakeIP.unmarshalFastJSON(fakeip)
	}
	o.StaticIPs = unmarshalFastJSONMapStringString(fj.GetObject("static_ips"))
	o.unmarshalJSONClient(fj)
	o.DNSClientOptions.unmarshalJSONClient(fj)
}

func (o *DNSServerOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Tag = stringNotNil(fj.GetStringBytes("tag"))
	o.Address = stringNotNil(fj.GetStringBytes("address"))
	o.Addresses = unmarshalFastJSONArrayStringWithName(fj, "addresses")
	o.AddressResolver = stringNotNil(fj.GetStringBytes("address_resolver"))
	o.AddressStrategy = unmarshalFastJSONDomainStrategy(stringNotNil(fj.GetStringBytes("address_strategy")))
	o.AddressFallbackDelay = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("address_fallback_delay")))
	o.Strategy = unmarshalFastJSONDomainStrategy(stringNotNil(fj.GetStringBytes("strategy")))
	o.Detour = stringNotNil(fj.GetStringBytes("detour"))
}

func (o *DNSClientOptions) unmarshalJSONClient(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Strategy = unmarshalFastJSONDomainStrategy(stringNotNil(fj.GetStringBytes("strategy")))
	o.DisableCache = fj.GetBool("disable_cache")
	o.DisableExpire = fj.GetBool("disable_expire")
	o.IndependentCache = fj.GetBool("independent_cache")
}

func (o *DNSFakeIPOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Enabled = fj.GetBool("enabled")
	inet4Range, err := netip.ParsePrefix(stringNotNil(fj.GetStringBytes("inet4_range")))
	if err == nil {
		o.Inet4Range = &inet4Range
	}

	inet6Range, err1 := netip.ParsePrefix(stringNotNil(fj.GetStringBytes("inet6_range")))
	if err1 == nil {
		o.Inet6Range = &inet6Range
	}
}

func unmarshalFastJSONArrayDNSServerOptions(fj *fastjson.Value) []DNSServerOptions {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}
	arr := fj.GetArray()
	if arr != nil {
		list := make([]DNSServerOptions, len(arr))
		for i, v := range arr {
			vv := DNSServerOptions{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]DNSServerOptions, 1)
	vv := DNSServerOptions{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

func unmarshalFastJSONArrayDNSRule(fj *fastjson.Value) []DNSRule {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make([]DNSRule, len(arr))
		for i, v := range arr {
			vv := DNSRule{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]DNSRule, 1)
	vv := DNSRule{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

// experimental.go
func (o *ExperimentalOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	clash_api := fj.Get("clash_api")
	if clash_api != nil && clash_api.Type() != fastjson.TypeNull {
		o.ClashAPI = &ClashAPIOptions{}
		o.ClashAPI.unmarshalFastJSON(clash_api)
	}
	v2ray_api := fj.Get("v2ray_api")
	if v2ray_api != nil && v2ray_api.Type() != fastjson.TypeNull {
		o.V2RayAPI = &V2RayAPIOptions{}
		o.V2RayAPI.unmarshalFastJSON(v2ray_api)
	}
	cache_file := fj.Get("cache_file")
	if cache_file != nil && cache_file.Type() != fastjson.TypeNull {
		o.CacheFile = &CacheFileOptions{}
		o.CacheFile.unmarshalFastJSON(cache_file)
	}
	debug := fj.Get("debug")
	if debug != nil && debug.Type() != fastjson.TypeNull {
		o.Debug = &DebugOptions{}
		o.Debug.unmarshalFastJSON(debug)
	}
}

// hysteria.go
func (o *HysteriaInboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.ListenOptions.unmarshalFastJSON(fj)
	o.Up = stringNotNil(fj.GetStringBytes("up"))
	o.UpMbps = fj.GetInt("up_mbps")
	o.Down = stringNotNil(fj.GetStringBytes("down"))
	o.DownMbps = fj.GetInt("down_mbps")
	o.Obfs = stringNotNil(fj.GetStringBytes("obfs"))
	o.Users = unmarshalFastJSONArrayHysteriaUser(fj.Get("users"))
	o.ReceiveWindowConn = fj.GetUint64("recv_window_conn")
	o.ReceiveWindowClient = fj.GetUint64("recv_window_client")
	o.MaxConnClient = fj.GetInt("max_conn_client")
	o.DisableMTUDiscovery = fj.GetBool("disable_mtu_discovery")
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
}
func (o *HysteriaUser) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Name = stringNotNil(fj.GetStringBytes("name"))
	o.Auth = fj.GetStringBytes("auth")
	o.AuthString = stringNotNil(fj.GetStringBytes("auth_str"))
}
func (o *HysteriaOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.DialerOptions.unmarshalFastJSON(fj)
	o.ServerOptions.unmarshalFastJSON(fj)
	o.Up = stringNotNil(fj.GetStringBytes("up"))
	o.UpMbps = fj.GetInt("up_mbps")
	o.Down = stringNotNil(fj.GetStringBytes("down"))
	o.DownMbps = fj.GetInt("down_mbps")
	o.Obfs = stringNotNil(fj.GetStringBytes("obfs"))
	o.Auth = fj.GetStringBytes("auth")
	o.AuthString = stringNotNil(fj.GetStringBytes("auth_str"))
	o.ReceiveWindowConn = fj.GetUint64("recv_window_conn")
	o.ReceiveWindow = fj.GetUint64("recv_window")
	o.DisableMTUDiscovery = fj.GetBool("disable_mtu_discovery")
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
	turn_relay := fj.Get("turn_relay")
	if turn_relay != nil && turn_relay.Type() != fastjson.TypeNull {
		o.TurnRelay = &TurnRelayOptions{}
		o.TurnRelay.unmarshalFastJSON(turn_relay)
	}
	o.HopPorts = stringNotNil(fj.GetStringBytes("hop_ports"))
	o.HopInterval = fj.GetInt("hop_interval")
}
func unmarshalFastJSONArrayHysteriaUser(fj *fastjson.Value) []HysteriaUser {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make([]HysteriaUser, len(arr))
		for i, v := range arr {
			vv := HysteriaUser{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]HysteriaUser, 1)
	vv := HysteriaUser{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

// hysteria2.go
func (o *Hysteria2InboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.ListenOptions.unmarshalFastJSON(fj)
	o.UpMbps = fj.GetInt("up_mbps")
	o.DownMbps = fj.GetInt("down_mbps")
	obfs := fj.Get("obfs")
	if obfs != nil && obfs.Type() != fastjson.TypeNull {
		o.Obfs = &Hysteria2Obfs{}
		o.Obfs.unmarshalFastJSON(obfs)
	}

	o.Users = unmarshalFastJSONArrayHysteria2User(fj.Get("users"))
	o.IgnoreClientBandwidth = fj.GetBool("ignore_client_bandwidth")
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
	o.Masquerade = stringNotNil(fj.GetStringBytes("masquerade"))
	o.BrutalDebug = fj.GetBool("brutal_debug")
}
func (o *Hysteria2Obfs) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Type = stringNotNil(fj.GetStringBytes("type"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
}
func (o *Hysteria2User) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Name = stringNotNil(fj.GetStringBytes("name"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
}
func (o *Hysteria2OutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.DialerOptions.unmarshalFastJSON(fj)
	o.ServerOptions.unmarshalFastJSON(fj)
	o.UpMbps = fj.GetInt("up_mbps")
	o.DownMbps = fj.GetInt("down_mbps")
	obfs := fj.Get("obfs")
	if obfs != nil && obfs.Type() != fastjson.TypeNull {
		o.Obfs = &Hysteria2Obfs{}
		o.Obfs.unmarshalFastJSON(obfs)
	}
	o.Password = stringNotNil(fj.GetStringBytes("password"))
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
	o.BrutalDebug = fj.GetBool("brutal_debug")
	turn_relay := fj.Get("turn_relay")
	if turn_relay != nil && turn_relay.Type() != fastjson.TypeNull {
		o.TurnRelay = &TurnRelayOptions{}
		o.TurnRelay.unmarshalFastJSON(turn_relay)
	}
	o.HopPorts = stringNotNil(fj.GetStringBytes("hop_ports"))
	o.HopInterval = fj.GetInt("hop_interval")
}
func unmarshalFastJSONArrayHysteria2User(fj *fastjson.Value) []Hysteria2User {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make([]Hysteria2User, len(arr))
		for i, v := range arr {
			vv := Hysteria2User{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]Hysteria2User, 1)
	vv := Hysteria2User{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

// inbound.go
func (h *Inbound) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	h.Type = stringNotNil(fj.GetStringBytes("type"))
	h.Tag = stringNotNil(fj.GetStringBytes("tag"))
	switch h.Type {
	case C.TypeTun:
		h.TunOptions.unmarshalFastJSON(fj)
	case C.TypeRedirect:
		h.RedirectOptions.unmarshalFastJSON(fj)
	case C.TypeTProxy:
		h.TProxyOptions.unmarshalFastJSON(fj)
	case C.TypeDirect:
		h.DirectOptions.unmarshalFastJSON(fj)
	case C.TypeSOCKS:
		h.SocksOptions.unmarshalFastJSON(fj)
	case C.TypeHTTP:
		h.HTTPOptions.unmarshalFastJSON(fj)
	case C.TypeMixed:
		h.MixedOptions.unmarshalFastJSON(fj)
	case C.TypeShadowsocks:
		h.ShadowsocksOptions.unmarshalFastJSON(fj)
	case C.TypeVMess:
		h.VMessOptions.unmarshalFastJSON(fj)
	case C.TypeTrojan:
		h.TrojanOptions.unmarshalFastJSON(fj)
	case C.TypeNaive:
		h.NaiveOptions.unmarshalFastJSON(fj)
	case C.TypeHysteria:
		h.HysteriaOptions.unmarshalFastJSON(fj)
	case C.TypeShadowTLS:
		h.ShadowTLSOptions.unmarshalFastJSON(fj)
	case C.TypeVLESS:
		h.VLESSOptions.unmarshalFastJSON(fj)
	case C.TypeTUIC:
		h.TUICOptions.unmarshalFastJSON(fj)
	case C.TypeHysteria2:
		h.Hysteria2Options.unmarshalFastJSON(fj)
	default:
		E.New("unknown inbound type: ", h.Type, h.Tag)
	}
}

func (o *InboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.SniffEnabled = fj.GetBool("sniff")
	o.SniffOverrideDestination = fj.GetBool("sniff_override_destination")
	o.SniffTimeout = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes(("sniff_timeout"))))
	o.DomainStrategy = unmarshalFastJSONDomainStrategy(stringNotNil(fj.GetStringBytes("domain_strategy")))
}

func (o *ListenOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	addr, err := netip.ParseAddr(stringNotNil(fj.GetStringBytes("listen")))
	if err == nil {
		o.Listen = new(ListenAddress)
		*(o.Listen) = ListenAddress(addr)
	}

	o.ListenPort = uint16(fj.GetUint("listen_port"))
	o.TCPFastOpen = fj.GetBool("tcp_fast_open")
	o.TCPMultiPath = fj.GetBool("tcp_multi_path")
	udp_fragment := fj.Get("udp_fragment")
	if udp_fragment != nil && udp_fragment.Type() != fastjson.TypeNull {
		o.UDPFragment = new(bool)
		*o.UDPFragment = fj.GetBool("udp_fragment")
	}
	//o.UDPFragmentDefault           = fj.GetBool("-")
	o.UDPTimeout = UDPTimeoutCompat(unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("udp_timeout"))))
	o.ProxyProtocol = fj.GetBool("proxy_protocol")
	o.ProxyProtocolAcceptNoHeader = fj.GetBool("proxy_protocol_accept_no_header")
	o.Detour = stringNotNil(fj.GetStringBytes("detour"))
	o.InboundOptions.unmarshalFastJSON(fj)
}

func unmarshalFastJSONArrayInbound(fj *fastjson.Value) []Inbound {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make([]Inbound, len(arr))
		for i, v := range arr {
			vv := Inbound{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]Inbound, 1)
	vv := Inbound{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

// naive.go
func (o *NaiveInboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.ListenOptions.unmarshalFastJSON(fj)
	o.Users = unmarshalFastJSONArrayAuthUser(fj.Get("users"))
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
}

// ntp.go
func (o *NTPOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Enabled = fj.GetBool("enabled")
	o.Server = stringNotNil(fj.GetStringBytes("server"))
	o.ServerPort = uint16(fj.GetUint("server_port"))
	o.Interval = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes(("interval"))))
	o.WriteToSystem = fj.GetBool("write_to_system")
	o.DialerOptions.unmarshalFastJSON(fj)
}

// outbound.go
func (h *Outbound) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	h.Type = stringNotNil(fj.GetStringBytes("type"))
	h.Tag = stringNotNil(fj.GetStringBytes("tag"))
	switch h.Type {
	case C.TypeDirect:
		h.DirectOptions.unmarshalFastJSON(fj)
	case C.TypeBlock, C.TypeDNS:
	case C.TypeSOCKS:
		h.SocksOptions.unmarshalFastJSON(fj)
	case C.TypeHTTP:
		h.HTTPOptions.unmarshalFastJSON(fj)
	case C.TypeShadowsocks:
		h.ShadowsocksOptions.unmarshalFastJSON(fj)
	case C.TypeVMess:
		h.VMessOptions.unmarshalFastJSON(fj)
	case C.TypeTrojan:
		h.TrojanOptions.unmarshalFastJSON(fj)
	case C.TypeWireGuard:
		h.WireGuardOptions.unmarshalFastJSON(fj)
	case C.TypeHysteria:
		h.HysteriaOptions.unmarshalFastJSON(fj)
	case C.TypeTor:
		h.TorOptions.unmarshalFastJSON(fj)
	case C.TypeSSH:
		h.SSHOptions.unmarshalFastJSON(fj)
	case C.TypeShadowTLS:
		h.ShadowTLSOptions.unmarshalFastJSON(fj)
	case C.TypeShadowsocksR:
		h.ShadowsocksROptions.unmarshalFastJSON(fj)
	case C.TypeVLESS:
		h.VLESSOptions.unmarshalFastJSON(fj)
	case C.TypeTUIC:
		h.TUICOptions.unmarshalFastJSON(fj)
	case C.TypeHysteria2:
		h.Hysteria2Options.unmarshalFastJSON(fj)
	case C.TypeXray:
		h.XrayOptions.unmarshalFastJSON(fj)
	case C.TypeSelector:
		h.SelectorOptions.unmarshalFastJSON(fj)
	case C.TypeURLTest:
		h.URLTestOptions.unmarshalFastJSON(fj)
	default:
		E.New("unknown outbound type: ", h.Type, h.Tag)
	}
}

func (o *DialerOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Detour = stringNotNil(fj.GetStringBytes("detour"))
	o.BindInterface = stringNotNil(fj.GetStringBytes("bind_interface"))
	addr4, err4 := netip.ParseAddr(stringNotNil(fj.GetStringBytes("inet4_bind_address")))
	if err4 == nil {
		o.Inet4BindAddress = new(ListenAddress)
		*(o.Inet4BindAddress) = ListenAddress(addr4)
	}
	addr6, err6 := netip.ParseAddr(stringNotNil(fj.GetStringBytes("inet6_bind_address")))
	if err6 == nil {
		o.Inet6BindAddress = new(ListenAddress)
		*(o.Inet6BindAddress) = ListenAddress(addr6)
	}
	o.ProtectPath = stringNotNil(fj.GetStringBytes("protect_path"))
	o.RoutingMark = fj.GetInt("routing_mark")
	o.ReuseAddr = fj.GetBool("reuse_addr")
	o.ConnectTimeout = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("connect_timeout")))
	o.TCPFastOpen = fj.GetBool("tcp_fast_open")
	o.TCPMultiPath = fj.GetBool("tcp_multi_path")
	udp_fragment := fj.Get("udp_fragment")
	tls_fragment := fj.Get("tls_fragment")
	if udp_fragment != nil && udp_fragment.Type() != fastjson.TypeNull {
		o.UDPFragment = new(bool)
		*o.UDPFragment = fj.GetBool("udp_fragment")
	}
	//o.UDPFragmentDefault = fj.GetBool("-")
	o.DomainStrategy = unmarshalFastJSONDomainStrategy(stringNotNil(fj.GetStringBytes("domain_strategy")))
	o.FallbackDelay = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("fallback_delay")))
	if tls_fragment != nil && tls_fragment.Type() != fastjson.TypeNull {
		o.TLSFragment = &TLSFragmentOptions{}
		o.TLSFragment.unmarshalFastJSON(tls_fragment)
	}
}

func (o *ServerOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Server = stringNotNil(fj.GetStringBytes("server"))
	o.ServerPort = uint16(fj.GetInt("server_port"))
}

func (o *OutboundMultiplexOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Enabled = fj.GetBool("enabled")
	o.Protocol = stringNotNil(fj.GetStringBytes("protocol"))
	o.MaxConnections = fj.GetInt("max_connections")
	o.MinStreams = fj.GetInt("min_streams")
	o.MaxStreams = fj.GetInt("max_streams")
	o.Padding = fj.GetBool("padding")
}

func unmarshalFastJSONArrayOutbound(fj *fastjson.Value) []Outbound {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make([]Outbound, len(arr))
		for i, v := range arr {
			vv := Outbound{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]Outbound, 1)
	vv := Outbound{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

func unmarshalFastJSONUser(o *auth.User, fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Username = stringNotNil(fj.GetStringBytes("username"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
}

func unmarshalFastJSONArrayAuthUser(fj *fastjson.Value) []auth.User {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make([]auth.User, len(arr))
		for i, v := range arr {
			vv := auth.User{}
			unmarshalFastJSONUser(&vv, v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]auth.User, 1)
	vv := auth.User{}
	unmarshalFastJSONUser(&vv, fj.Get())
	list[0] = vv
	return list
}

// platform.go
func (o *OnDemandOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Enabled = fj.GetBool("enabled")
	o.Rules = unmarshalFastJSONArrayOnDemandRule(fj.Get("rules"))
}
func (o *OnDemandRule) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	action := fj.Get("action")
	if action != nil && action.Type() != fastjson.TypeNull {
		o.Action = new(OnDemandRuleAction)
		*o.Action = OnDemandRuleAction(fj.GetInt("action"))
	}
	o.DNSSearchDomainMatch = unmarshalFastJSONListableString(fj.Get("dns_search_domain_match"))
	o.DNSServerAddressMatch = unmarshalFastJSONListableString(fj.Get("dns_server_address_match"))
	interface_type_match := fj.Get("interface_type_match")
	if interface_type_match != nil && interface_type_match.Type() != fastjson.TypeNull {
		o.InterfaceTypeMatch = new(OnDemandRuleInterfaceType)
		*o.InterfaceTypeMatch = OnDemandRuleInterfaceType(fj.GetInt("interface_type_match"))
	}

	o.SSIDMatch = unmarshalFastJSONListableString(fj.Get("ssid_match"))
	o.ProbeURL = stringNotNil(fj.GetStringBytes("probe_url"))
}

func unmarshalFastJSONArrayOnDemandRule(fj *fastjson.Value) []OnDemandRule {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make([]OnDemandRule, len(arr))
		for i, v := range arr {
			vv := OnDemandRule{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]OnDemandRule, 1)
	vv := OnDemandRule{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

// redir.go
func (o *RedirectInboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.ListenOptions.unmarshalFastJSON(fj)
}

func (o *TProxyInboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.ListenOptions.unmarshalFastJSON(fj)
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
}

// route.go
func (o *RouteOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	geoip := fj.Get("geoip")
	if geoip != nil && geoip.Type() != fastjson.TypeNull {
		o.GeoIP = &GeoIPOptions{}
		o.GeoIP.unmarshalFastJSON(geoip)
	}
	geosite := fj.Get("geosite")
	if geosite != nil && geosite.Type() != fastjson.TypeNull {
		o.Geosite = &GeositeOptions{}
		o.Geosite.unmarshalFastJSON(geosite)
	}

	o.Rules = unmarshalFastJSONArrayRule(fj.Get("rules"))
	o.RuleSet = unmarshalFastJSONArrayRuleSet(fj.Get("rule_set"))
	o.Final = stringNotNil(fj.GetStringBytes("final"))
	o.FindProcess = fj.GetBool("find_process")
	o.AutoDetectInterface = fj.GetBool("auto_detect_interface")
	o.OverrideAndroidVPN = fj.GetBool("override_android_vpn")
	o.DefaultInterface = stringNotNil(fj.GetStringBytes("default_interface"))
	o.DefaultMark = fj.GetInt("default_mark")
}

func (o *GeoIPOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Path = stringNotNil(fj.GetStringBytes("path"))
	o.DownloadURL = stringNotNil(fj.GetStringBytes("download_url"))
	o.DownloadDetour = stringNotNil(fj.GetStringBytes("download_detour"))
}

func (o *GeositeOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Path = stringNotNil(fj.GetStringBytes("path"))
	o.DownloadURL = stringNotNil(fj.GetStringBytes("download_url"))
	o.DownloadDetour = stringNotNil(fj.GetStringBytes("download_detour"))
}

// rule_dns.go
func (o *DNSRule) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Type = stringNotNil(fj.GetStringBytes("type"))
	switch o.Type {
	case "", C.RuleTypeDefault:
		o.Type = C.RuleTypeDefault
		o.DefaultOptions.unmarshalFastJSON(fj)
	case C.RuleTypeLogical:
		o.LogicalOptions.unmarshalFastJSON(fj)
	}
}

func (o *DefaultDNSRule) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Inbound = unmarshalFastJSONListableString(fj.Get("inbound"))
	o.IPVersion = fj.GetInt("ip_version")
	o.QueryType = unmarshalFastJSONListableDNSQueryType(fj.Get("query_type"))
	o.Network = unmarshalFastJSONListableString(fj.Get("network"))
	o.AuthUser = unmarshalFastJSONListableString(fj.Get("auth_user"))
	o.Protocol = unmarshalFastJSONListableString(fj.Get("protocol"))
	o.Domain = unmarshalFastJSONListableString(fj.Get("domain"))
	o.DomainSuffix = unmarshalFastJSONListableString(fj.Get("domain_suffix"))
	o.DomainKeyword = unmarshalFastJSONListableString(fj.Get("domain_keyword"))
	o.DomainRegex = unmarshalFastJSONListableString(fj.Get("domain_regex"))
	o.Geosite = unmarshalFastJSONListableString(fj.Get("geosite"))
	o.SourceGeoIP = unmarshalFastJSONListableString(fj.Get("source_geoip"))
	//todo
	//o.GeoIP = unmarshalFastJSONListableString(fj.Get("geoip"))
	//o.IPCIDR = unmarshalFastJSONListableString(fj.Get("ip_cidr"))
	//o.IPIsPrivate = fj.GetBool("ip_is_private")
	o.SourceIPCIDR = unmarshalFastJSONListableString(fj.Get("source_ip_cidr"))
	o.SourceIPIsPrivate = fj.GetBool("source_ip_is_private")
	o.SourcePort = unmarshalFastJSONListableUint16(fj.Get("source_port"))
	o.SourcePortRange = unmarshalFastJSONListableString(fj.Get("source_port_range"))
	o.Port = unmarshalFastJSONListableUint16(fj.Get("port"))
	o.PortRange = unmarshalFastJSONListableString(fj.Get("port_range"))
	o.ProcessName = unmarshalFastJSONListableString(fj.Get("process_name"))
	o.ProcessPath = unmarshalFastJSONListableString(fj.Get("process_path"))
	o.PackageName = unmarshalFastJSONListableString(fj.Get("package_name"))
	o.User = unmarshalFastJSONListableString(fj.Get("user"))
	o.UserID = unmarshalFastJSONListableInt32(fj.Get("user_id"))
	o.Outbound = unmarshalFastJSONListableString(fj.Get("outbound"))
	o.ClashMode = stringNotNil(fj.GetStringBytes("clash_mode"))
	o.WIFISSID = unmarshalFastJSONListableString(fj.Get("wifi_ssid"))
	o.WIFIBSSID = unmarshalFastJSONListableString(fj.Get("wifi_bssid"))
	o.RuleSet = unmarshalFastJSONListableString(fj.Get("rule_set"))
	o.Invert = fj.GetBool("invert")
	o.Server = stringNotNil(fj.GetStringBytes("server"))
	o.DisableCache = fj.GetBool("disable_cache")
	rewrite_ttl := fj.Get("rewrite_ttl")
	if rewrite_ttl != nil && rewrite_ttl.Type() != fastjson.TypeNull {
		o.RewriteTTL = new(uint32)
		*o.RewriteTTL = uint32(fj.GetUint("rewrite_ttl"))
	}
	o.Name = stringNotNil(fj.GetStringBytes("name"))
}

func (o *LogicalDNSRule) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Mode = stringNotNil(fj.GetStringBytes("mode"))
	o.Rules = unmarshalFastJSONArrayDNSRule(fj.Get("rules"))
	o.Invert = fj.GetBool("invert")
	o.Server = stringNotNil(fj.GetStringBytes("server"))
	o.DisableCache = fj.GetBool("disable_cache")
	rewrite_ttl := fj.Get("rewrite_ttl")
	if rewrite_ttl != nil && rewrite_ttl.Type() != fastjson.TypeNull {
		o.RewriteTTL = new(uint32)
		*o.RewriteTTL = uint32(fj.GetUint("rewrite_ttl"))
	}
	o.Name = stringNotNil(fj.GetStringBytes("name"))
}

func (o *RuleSet) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	//LocalOptions  LocalRuleSet  `json:"-"`
	//RemoteOptions RemoteRuleSet `json:"-"`

	o.Type = stringNotNil(fj.GetStringBytes("type"))
	o.Tag = stringNotNil(fj.GetStringBytes("tag"))
	o.Format = stringNotNil(fj.GetStringBytes("format"))
	o.LocalOptions.IsAsset = fj.GetBool("is_asset")
	o.LocalOptions.Path = stringNotNil(fj.GetStringBytes("path"))
	o.RemoteOptions.URL = stringNotNil(fj.GetStringBytes("url"))
	o.RemoteOptions.DownloadDetour = stringNotNil(fj.GetStringBytes("download_detour"))
	o.RemoteOptions.UpdateInterval = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("update_interval")))
}
func unmarshalFastJSONArrayRuleSet(fj *fastjson.Value) []RuleSet {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make([]RuleSet, len(arr))
		for i, v := range arr {
			vv := RuleSet{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]RuleSet, 1)
	vv := RuleSet{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

func unmarshalFastJSONArrayRule(fj *fastjson.Value) []Rule {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}
	arr := fj.GetArray()
	if arr != nil {
		list := make([]Rule, len(arr))
		for i, v := range arr {
			vv := Rule{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]Rule, 1)
	vv := Rule{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

func unmarshalFastJSONArrayDefaultDNSRule(fj *fastjson.Value) []DefaultDNSRule {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make([]DefaultDNSRule, len(arr))
		for i, v := range arr {
			vv := DefaultDNSRule{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]DefaultDNSRule, 1)
	vv := DefaultDNSRule{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

// rule.go
func (o *Rule) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Type = stringNotNil(fj.GetStringBytes("type"))
	switch o.Type {
	case "", C.RuleTypeDefault:
		o.Type = C.RuleTypeDefault
		o.DefaultOptions.unmarshalFastJSON(fj)
	case C.RuleTypeLogical:
		o.LogicalOptions.unmarshalFastJSON(fj)
	default:
		E.New("unknown rule type: " + o.Type)
	}
}

func (o *DefaultRule) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Inbound = unmarshalFastJSONListableString(fj.Get("inbound"))
	o.IPVersion = fj.GetInt("ip_version")
	o.Network = unmarshalFastJSONListableString(fj.Get("network"))
	o.AuthUser = unmarshalFastJSONListableString(fj.Get("auth_user"))
	o.Protocol = unmarshalFastJSONListableString(fj.Get("protocol"))
	o.Domain = unmarshalFastJSONListableString(fj.Get("domain"))
	o.DomainSuffix = unmarshalFastJSONListableString(fj.Get("domain_suffix"))
	o.DomainKeyword = unmarshalFastJSONListableString(fj.Get("domain_keyword"))
	o.DomainRegex = unmarshalFastJSONListableString(fj.Get("domain_regex"))
	o.Geosite = unmarshalFastJSONListableString(fj.Get("geosite"))
	o.SourceGeoIP = unmarshalFastJSONListableString(fj.Get("source_geoip"))
	o.GeoIP = unmarshalFastJSONListableString(fj.Get("geoip"))
	o.SourceIPCIDR = unmarshalFastJSONListableString(fj.Get("source_ip_cidr"))
	o.SourceIPIsPrivate = fj.GetBool("source_ip_is_private")
	o.IPCIDR = unmarshalFastJSONListableString(fj.Get("ip_cidr"))
	o.IPIsPrivate = fj.GetBool("ip_is_private")
	o.SourcePort = unmarshalFastJSONListableUint16(fj.Get("source_port"))
	o.SourcePortRange = unmarshalFastJSONListableString(fj.Get("source_port_range"))
	o.Port = unmarshalFastJSONListableUint16(fj.Get("port"))
	o.PortRange = unmarshalFastJSONListableString(fj.Get("port_range"))
	o.ProcessName = unmarshalFastJSONListableString(fj.Get("process_name"))
	o.ProcessPath = unmarshalFastJSONListableString(fj.Get("process_path"))
	o.PackageName = unmarshalFastJSONListableString(fj.Get("package_name"))
	o.User = unmarshalFastJSONListableString(fj.Get("user"))
	o.UserID = unmarshalFastJSONListableInt32(fj.Get("user_id"))
	o.ClashMode = stringNotNil(fj.GetStringBytes("clash_mode"))
	o.WIFISSID = unmarshalFastJSONListableString(fj.Get("wifi_ssid"))
	o.WIFIBSSID = unmarshalFastJSONListableString(fj.Get("wifi_bssid"))
	o.RuleSet = unmarshalFastJSONListableString(fj.Get("rule_set"))
	o.RuleSetIPCIDRMatchSource = fj.GetBool("rule_set_ipcidr_match_source")
	o.Invert = fj.GetBool("invert")
	o.Outbound = stringNotNil(fj.GetStringBytes("outbound"))
	o.Name = stringNotNil(fj.GetStringBytes("name"))
}

func (o *LogicalRule) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Mode = stringNotNil(fj.GetStringBytes("mode"))
	o.Rules = unmarshalFastJSONArrayRule(fj.Get("rules"))
	o.Invert = fj.GetBool("invert")
	o.Outbound = stringNotNil(fj.GetStringBytes("outbound"))
	o.Name = stringNotNil(fj.GetStringBytes("name"))
}

// shadowsocks.go
func (o *ShadowsocksInboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.ListenOptions.unmarshalFastJSON(fj)
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
	o.Method = stringNotNil(fj.GetStringBytes("method"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
	o.Users = unmarshalFastJSONArrayShadowsocksUser(fj.Get("users"))
	o.Destinations = unmarshalFastJSONArrayShadowsocksDestination(fj.Get("destinations"))
}
func (o *ShadowsocksUser) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Name = stringNotNil(fj.GetStringBytes("name"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
}
func (o *ShadowsocksDestination) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Name = stringNotNil(fj.GetStringBytes("name"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
	o.ServerOptions.unmarshalFastJSON(fj)
}
func (o *ShadowsocksOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.DialerOptions.unmarshalFastJSON(fj)
	o.ServerOptions.unmarshalFastJSON(fj)
	o.Method = stringNotNil(fj.GetStringBytes("method"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
	o.Plugin = stringNotNil(fj.GetStringBytes("plugin"))
	o.PluginOptions = stringNotNil(fj.GetStringBytes("plugin_opts"))
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
	udp_over_tcp := fj.Get("udp_over_tcp")
	if udp_over_tcp != nil && udp_over_tcp.Type() != fastjson.TypeNull {
		o.UDPOverTCP = &UDPOverTCPOptions{}
		o.UDPOverTCP.unmarshalFastJSON(udp_over_tcp)
	}
	multiplex := fj.Get("multiplex")
	if multiplex != nil && multiplex.Type() != fastjson.TypeNull {
		o.Multiplex = &OutboundMultiplexOptions{}
		o.Multiplex.unmarshalFastJSON(multiplex)
	}

}
func unmarshalFastJSONArrayShadowsocksUser(fj *fastjson.Value) []ShadowsocksUser {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	arr := fj.GetArray()
	if arr != nil {
		list := make([]ShadowsocksUser, len(arr))
		for i, v := range arr {
			vv := ShadowsocksUser{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]ShadowsocksUser, 1)
	vv := ShadowsocksUser{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}
func unmarshalFastJSONArrayShadowsocksDestination(fj *fastjson.Value) []ShadowsocksDestination {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}
	arr := fj.GetArray()
	if arr != nil {
		list := make([]ShadowsocksDestination, len(arr))
		for i, v := range arr {
			vv := ShadowsocksDestination{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]ShadowsocksDestination, 1)
	vv := ShadowsocksDestination{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

// shadowsocksr.go
func (o *ShadowsocksROutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.DialerOptions.unmarshalFastJSON(fj)
	o.ServerOptions.unmarshalFastJSON(fj)
	o.Method = stringNotNil(fj.GetStringBytes("method"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
	o.Obfs = stringNotNil(fj.GetStringBytes("obfs"))
	o.ObfsParam = stringNotNil(fj.GetStringBytes("obfs_param"))
	o.Protocol = stringNotNil(fj.GetStringBytes("protocol"))
	o.ProtocolParam = stringNotNil(fj.GetStringBytes("protocol_param"))
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
}

// shadowtls.go
func (o *ShadowTLSInboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.ListenOptions.unmarshalFastJSON(fj)
	o.Version = fj.GetInt("version")
	o.Password = stringNotNil(fj.GetStringBytes("password"))
	o.Users = unmarshalFastJSONArrayShadowTLSUser(fj.Get("users"))
	o.Handshake.unmarshalFastJSON(fj.Get("handshake"))
	o.HandshakeForServerName = unmarshalFastJSONMapShadowTLSHandshakeOptions(fj.GetArray("handshake_for_server_name"))
	o.StrictMode = fj.GetBool("strict_mode")
}
func (o *ShadowTLSUser) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Name = stringNotNil(fj.GetStringBytes("name"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
}
func (o *ShadowTLSHandshakeOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.ServerOptions.unmarshalFastJSON(fj)
	o.DialerOptions.unmarshalFastJSON(fj)
}
func (o *ShadowTLSOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.DialerOptions.unmarshalFastJSON(fj)
	o.ServerOptions.unmarshalFastJSON(fj)
	o.Version = fj.GetInt("version")
	o.Password = stringNotNil(fj.GetStringBytes("password"))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
}
func unmarshalFastJSONArrayShadowTLSUser(fj *fastjson.Value) []ShadowTLSUser {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}
	arr := fj.GetArray()
	if arr != nil {
		list := make([]ShadowTLSUser, len(arr))
		for i, v := range arr {
			vv := ShadowTLSUser{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]ShadowTLSUser, 1)
	vv := ShadowTLSUser{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

func unmarshalFastJSONMapShadowTLSHandshakeOptions(fj []*fastjson.Value) map[string]ShadowTLSHandshakeOptions {
	if fj == nil {
		return nil
	}

	list := make(map[string]ShadowTLSHandshakeOptions, len(fj))
	/*for i, v := range fj {
		vv := ShadowTLSHandshakeOptions{}
		vv.unmarshalFastJSON(v)
		list[i] = vv
	}*/
	return list
}

// simple.go
func (o *SocksInboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.ListenOptions.unmarshalFastJSON(fj)
	o.Users = unmarshalFastJSONArrayAuthUser(fj.Get("users"))
}

func (o *HTTPMixedInboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.ListenOptions.unmarshalFastJSON(fj)
	o.Users = unmarshalFastJSONArrayAuthUser(fj.Get("users"))
	o.SetSystemProxy = fj.GetBool("set_system_proxy")
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
}

func (o *SocksOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.DialerOptions.unmarshalFastJSON(fj)
	o.ServerOptions.unmarshalFastJSON(fj)
	o.Version = stringNotNil(fj.GetStringBytes("version"))
	o.Username = stringNotNil(fj.GetStringBytes("username"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
}

func (o *HTTPOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.DialerOptions.unmarshalFastJSON(fj)
	o.ServerOptions.unmarshalFastJSON(fj)

	o.Username = stringNotNil(fj.GetStringBytes("username"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}

	o.Path = stringNotNil(fj.GetStringBytes("path"))
	o.Headers = unmarshalFastJSONMapHTTPHeader(fj.GetObject("headers"))
}

// ssh.go
func (o *SSHOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.DialerOptions.unmarshalFastJSON(fj)
	o.ServerOptions.unmarshalFastJSON(fj)
	o.User = stringNotNil(fj.GetStringBytes("user"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
	o.PrivateKey = unmarshalFastJSONListableString(fj.Get("private_key"))
	o.PrivateKeyPath = stringNotNil(fj.GetStringBytes("private_key_path"))
	o.PrivateKeyPassphrase = stringNotNil(fj.GetStringBytes("private_key_passphrase"))
	o.HostKey = unmarshalFastJSONListableString(fj.Get("host_key"))
	o.HostKeyAlgorithms = unmarshalFastJSONListableString(fj.Get("host_key_algorithms"))
	o.ClientVersion = stringNotNil(fj.GetStringBytes("client_version"))
	udp_over_tcp := fj.Get("udp_over_tcp")
	if udp_over_tcp != nil && udp_over_tcp.Type() != fastjson.TypeNull {
		o.UDPOverTCP = &UDPOverTCPOptions{}
		o.UDPOverTCP.unmarshalFastJSON(udp_over_tcp)
	}
}

// tls_acme.go
func (o *InboundACMEOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Domain = unmarshalFastJSONListableString(fj.Get("domain"))
	o.DataDirectory = stringNotNil(fj.GetStringBytes("data_directory"))
	o.DefaultServerName = stringNotNil(fj.GetStringBytes("default_server_name"))
	o.Email = stringNotNil(fj.GetStringBytes("email"))
	o.Provider = stringNotNil(fj.GetStringBytes("provider"))
	o.DisableHTTPChallenge = fj.GetBool("disable_http_challenge")
	o.DisableTLSALPNChallenge = fj.GetBool("disable_tls_alpn_challenge")
	o.AlternativeHTTPPort = uint16(fj.GetUint("alternative_http_port"))
	o.AlternativeTLSPort = uint16(fj.GetUint("alternative_tls_port"))
	external_account := fj.Get("external_account")
	dns01_challenge := fj.Get("dns01_challenge")
	if external_account != nil && external_account.Type() != fastjson.TypeNull {
		o.ExternalAccount = &ACMEExternalAccountOptions{}
		o.ExternalAccount.unmarshalFastJSON(external_account)
	}
	if dns01_challenge != nil && dns01_challenge.Type() != fastjson.TypeNull {
		o.DNS01Challenge = &ACMEDNS01ChallengeOptions{}
		o.DNS01Challenge.unmarshalFastJSON(dns01_challenge)
	}
}

func (o *ACMEExternalAccountOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.KeyID = stringNotNil(fj.GetStringBytes("key_id"))
	o.MACKey = stringNotNil(fj.GetStringBytes("mac_key"))
}

func (o *ACMEDNS01ChallengeOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Provider = stringNotNil(fj.GetStringBytes("provider"))
	switch o.Provider {
	case C.DNSProviderAliDNS:
		o.AliDNSOptions.unmarshalFastJSON(fj)
	case C.DNSProviderCloudflare:
		o.CloudflareOptions.unmarshalFastJSON(fj)
	default:
		E.New("unknown provider type: " + o.Provider)
	}
}

func (o *ACMEDNS01AliDNSOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.AccessKeyID = stringNotNil(fj.GetStringBytes("access_key_id"))
	o.AccessKeySecret = stringNotNil(fj.GetStringBytes("access_key_secret"))
	o.RegionID = stringNotNil(fj.GetStringBytes("region_id"))
}

func (o *ACMEDNS01CloudflareOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.APIToken = stringNotNil(fj.GetStringBytes("api_token"))
}

// tls.go
func (o *InboundTLSOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Enabled = fj.GetBool("enabled")
	o.ServerName = stringNotNil(fj.GetStringBytes("server_name"))
	o.Insecure = fj.GetBool("insecure")
	o.ALPN = unmarshalFastJSONListableString(fj.Get("alpn"))
	o.MinVersion = stringNotNil(fj.GetStringBytes("min_version"))
	o.MaxVersion = stringNotNil(fj.GetStringBytes("max_version"))
	o.CipherSuites = unmarshalFastJSONListableString(fj.Get("cipher_suites"))
	o.Certificate = unmarshalFastJSONListableString(fj.Get("certificate"))
	o.CertificatePath = stringNotNil(fj.GetStringBytes("certificate_path"))
	o.Key = unmarshalFastJSONListableString(fj.Get("key"))
	o.KeyPath = stringNotNil(fj.GetStringBytes("key_path"))
	acme := fj.Get("acme")
	if acme != nil && acme.Type() != fastjson.TypeNull {
		o.ACME = &InboundACMEOptions{}
		o.ACME.unmarshalFastJSON(acme)
	}
	ech := fj.Get("ech")
	if ech != nil && ech.Type() != fastjson.TypeNull {
		o.ECH = &InboundECHOptions{}
		o.ECH.unmarshalFastJSON(ech)
	}
	reality := fj.Get("reality")
	if reality != nil && reality.Type() != fastjson.TypeNull {
		o.Reality = &InboundRealityOptions{}
		o.Reality.unmarshalFastJSON(reality)
	}
}

func (o *OutboundTLSOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Enabled = fj.GetBool("enabled")
	o.DisableSNI = fj.GetBool("disable_sni")
	o.ServerName = stringNotNil(fj.GetStringBytes("server_name"))
	o.Insecure = fj.GetBool("insecure")
	o.ALPN = unmarshalFastJSONListableString(fj.Get("alpn"))
	o.MinVersion = stringNotNil(fj.GetStringBytes("min_version"))
	o.MaxVersion = stringNotNil(fj.GetStringBytes("max_version"))
	o.CipherSuites = unmarshalFastJSONListableString(fj.Get("cipher_suites"))
	o.Certificate = unmarshalFastJSONListableString(fj.Get("certificate"))
	o.CertificatePath = stringNotNil(fj.GetStringBytes("certificate_path"))
	ech := fj.Get("ech")
	if ech != nil && ech.Type() != fastjson.TypeNull {
		o.ECH = &OutboundECHOptions{}
		o.ECH.unmarshalFastJSON(ech)
	}
	utls := fj.Get("utls")
	if utls != nil && utls.Type() != fastjson.TypeNull {
		o.UTLS = &OutboundUTLSOptions{}
		o.UTLS.unmarshalFastJSON(utls)
	}
	reality := fj.Get("reality")
	if reality != nil && reality.Type() != fastjson.TypeNull {
		o.Reality = &OutboundRealityOptions{}
		o.Reality.unmarshalFastJSON(reality)
	}
	tls_tricks := fj.Get("tls_tricks")
	if tls_tricks != nil && tls_tricks.Type() != fastjson.TypeNull {
		o.TLSTricks = &TLSTricksOptions{}
		o.TLSTricks.unmarshalFastJSON(tls_tricks)
	}
}

func (o *InboundRealityOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Enabled = fj.GetBool("enabled")
	o.Handshake.unmarshalFastJSON(fj.Get("handshake"))
	o.PrivateKey = stringNotNil(fj.GetStringBytes("private_key"))
	o.ShortID = unmarshalFastJSONListableString(fj.Get("short_id"))
	o.MaxTimeDifference = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("max_time_difference")))
}

func (o *InboundRealityHandshakeOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.ServerOptions.unmarshalFastJSON(fj)
	o.DialerOptions.unmarshalFastJSON(fj)
}
func (o *InboundECHOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Enabled = fj.GetBool("enabled")
	o.PQSignatureSchemesEnabled = fj.GetBool("pq_signature_schemes_enabled")
	o.DynamicRecordSizingDisabled = fj.GetBool("dynamic_record_sizing_disabled")
	o.Key = unmarshalFastJSONListableString(fj.Get("key"))
	o.KeyPath = stringNotNil(fj.GetStringBytes("key_path"))
}
func (o *OutboundECHOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Enabled = fj.GetBool("enabled")
	o.PQSignatureSchemesEnabled = fj.GetBool("pq_signature_schemes_enabled")
	o.DynamicRecordSizingDisabled = fj.GetBool("dynamic_record_sizing_disabled")
	o.Config = unmarshalFastJSONListableString(fj.Get("config"))
	o.ConfigPath = stringNotNil(fj.GetStringBytes("config_path"))

}
func (o *OutboundUTLSOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Enabled = fj.GetBool("enabled")
	o.Fingerprint = stringNotNil(fj.GetStringBytes("fingerprint"))
}
func (o *OutboundRealityOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Enabled = fj.GetBool("enabled")
	o.PublicKey = stringNotNil(fj.GetStringBytes("public_key"))
	o.ShortID = stringNotNil(fj.GetStringBytes("short_id"))
}

// tor.go
func (o *TorOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.DialerOptions.unmarshalFastJSON(fj)
	o.ExecutablePath = stringNotNil(fj.GetStringBytes("executable_path"))
	o.ExtraArgs = unmarshalFastJSONArrayStringWithName(fj, "extra_args")
	o.DataDirectory = stringNotNil(fj.GetStringBytes("data_directory"))
	o.Options = unmarshalFastJSONMapString(fj.GetArray("torrc"))
}
func unmarshalFastJSONMapString(fj []*fastjson.Value) map[string]string {
	if fj == nil {
		return nil
	}

	list := make(map[string]string, len(fj))
	/*for i, v := range fj {
		by, err := v.StringBytes()
		if err == nil {
			list[i] = stringNotNil(by)
		}
	}*/
	return list
}

// trojan.go
func (o *TrojanInboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.ListenOptions.unmarshalFastJSON(fj)
	o.Users = unmarshalFastJSONArrayTrojanUser(fj.Get("users"))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
	fallback := fj.Get("fallback")
	if fallback != nil && fallback.Type() != fastjson.TypeNull {
		o.Fallback = &ServerOptions{}
		o.Fallback.unmarshalFastJSON(fallback)
	}

	o.FallbackForALPN = unmarshalFastJSONMapServerOptions(fj.GetArray("fallback_for_alpn"))
	transport := fj.Get("transport")
	if transport != nil && transport.Type() != fastjson.TypeNull {
		o.Transport = &V2RayTransportOptions{}
		o.Transport.unmarshalFastJSON(transport)
	}
}
func (o *TrojanUser) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Name = stringNotNil(fj.GetStringBytes("name"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
}
func (o *TrojanOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.DialerOptions.unmarshalFastJSON(fj)
	o.ServerOptions.unmarshalFastJSON(fj)
	o.Password = stringNotNil(fj.GetStringBytes("password"))
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
	multiplex := fj.Get("multiplex")
	if multiplex != nil && multiplex.Type() != fastjson.TypeNull {
		o.Multiplex = &OutboundMultiplexOptions{}
		o.Multiplex.unmarshalFastJSON(multiplex)
	}
	transport := fj.Get("transport")
	if transport != nil && transport.Type() != fastjson.TypeNull {
		o.Transport = &V2RayTransportOptions{}
		o.Transport.unmarshalFastJSON(transport)
	}
}
func unmarshalFastJSONArrayTrojanUser(fj *fastjson.Value) []TrojanUser {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	arr := fj.GetArray()
	if arr != nil {
		list := make([]TrojanUser, len(arr))
		for i, v := range arr {
			vv := TrojanUser{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]TrojanUser, 1)
	vv := TrojanUser{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}
func unmarshalFastJSONMapServerOptions(fj []*fastjson.Value) map[string]*ServerOptions {
	if fj == nil {
		return nil
	}

	list := make(map[string]*ServerOptions, len(fj))
	/*for i, v := range fj {
		by, err := v.StringBytes()
		if err == nil {
			list[i] = stringNotNil(by)
		}
	}*/
	return list
}

// tuic.go
func (o *TUICInboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.ListenOptions.unmarshalFastJSON(fj)
	o.Users = unmarshalFastJSONArrayTUICUser(fj.Get("users"))
	o.CongestionControl = stringNotNil(fj.GetStringBytes("congestion_control"))
	o.AuthTimeout = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("auth_timeout")))
	o.ZeroRTTHandshake = fj.GetBool("zero_rtt_handshake")
	o.Heartbeat = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("heartbeat")))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
}
func (o *TUICUser) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Name = stringNotNil(fj.GetStringBytes("name"))
	o.UUID = stringNotNil(fj.GetStringBytes("uuid"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
}
func (o *TUICOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.DialerOptions.unmarshalFastJSON(fj)
	o.ServerOptions.unmarshalFastJSON(fj)
	o.UUID = stringNotNil(fj.GetStringBytes("uuid"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
	o.CongestionControl = stringNotNil(fj.GetStringBytes("congestion_control"))
	o.UDPRelayMode = stringNotNil(fj.GetStringBytes("udp_relay_mode"))
	o.UDPOverStream = fj.GetBool("udp_over_stream")
	o.ZeroRTTHandshake = fj.GetBool("zero_rtt_handshake")
	o.Heartbeat = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("heartbeat")))
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
	turn_relay := fj.Get("turn_relay")
	if turn_relay != nil && turn_relay.Type() != fastjson.TypeNull {
		o.TurnRelay = &TurnRelayOptions{}
		o.TurnRelay.unmarshalFastJSON(turn_relay)
	}
}
func unmarshalFastJSONArrayTUICUser(fj *fastjson.Value) []TUICUser {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}
	arr := fj.GetArray()
	if arr != nil {
		list := make([]TUICUser, len(arr))
		for i, v := range arr {
			vv := TUICUser{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]TUICUser, 1)
	vv := TUICUser{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

// tun_platform.go
func (o *TunPlatformOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	http_proxy := fj.Get("http_proxy")
	if http_proxy != nil && http_proxy.Type() != fastjson.TypeNull {
		o.HTTPProxy = &HTTPProxyOptions{}
		o.HTTPProxy.unmarshalFastJSON(http_proxy)
	}
	o.AllowBypass = fj.GetBool("allow_bypass")
}
func (o *HTTPProxyOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Enabled = fj.GetBool("enabled")
	o.ServerOptions.unmarshalFastJSON(fj)
}

// tun.go
func (o *TunInboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.InterfaceName = stringNotNil(fj.GetStringBytes("interface_name"))
	o.MTU = uint32(fj.GetUint("mtu"))
	o.GSO = fj.GetBool("gso")
	o.Inet4Address = unmarshalFastJSONListableNetipPrefix(fj.Get("inet4_address"))
	o.Inet6Address = unmarshalFastJSONListableNetipPrefix(fj.Get("inet6_address"))
	o.AutoRoute = fj.GetBool("auto_route")
	o.StrictRoute = fj.GetBool("strict_route")
	o.Inet4RouteAddress = unmarshalFastJSONListableNetipPrefix(fj.Get("inet4_route_address"))
	o.Inet6RouteAddress = unmarshalFastJSONListableNetipPrefix(fj.Get("inet6_route_address"))
	o.Inet4RouteExcludeAddress = unmarshalFastJSONListableNetipPrefix(fj.Get("inet4_route_exclude_address"))
	o.Inet6RouteExcludeAddress = unmarshalFastJSONListableNetipPrefix(fj.Get("inet6_route_exclude_address"))
	o.IncludeInterface = unmarshalFastJSONListableString(fj.Get("include_interface"))
	o.ExcludeInterface = unmarshalFastJSONListableString(fj.Get("exclude_interface"))
	o.IncludeUID = unmarshalFastJSONListableUInt32(fj.Get("include_uid"))
	o.IncludeUIDRange = unmarshalFastJSONListableString(fj.Get("include_uid_range"))
	o.ExcludeUID = unmarshalFastJSONListableUInt32(fj.Get("exclude_uid"))
	o.ExcludeUIDRange = unmarshalFastJSONListableString(fj.Get("exclude_uid_range"))
	o.IncludeAndroidUser = unmarshalFastJSONListableInt(fj.Get("include_android_user"))
	o.IncludePackage = unmarshalFastJSONListableString(fj.Get("include_package"))
	o.ExcludePackage = unmarshalFastJSONListableString(fj.Get("exclude_package"))
	o.EndpointIndependentNat = fj.GetBool("endpoint_independent_nat")
	o.UDPTimeout = UDPTimeoutCompat(unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("udp_timeout"))))
	o.Stack = stringNotNil(fj.GetStringBytes("stack"))
	platform := fj.Get("platform")
	if platform != nil && platform.Type() != fastjson.TypeNull {
		o.Platform = &TunPlatformOptions{}
		o.Platform.unmarshalFastJSON(platform)
	}

	o.InboundOptions.unmarshalFastJSON(fj)
}

// udp_over_tcp.go
func (o *UDPOverTCPOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Enabled = fj.GetBool("enabled")
	o.Version = uint8(fj.GetUint("version"))
}

// v2ray_transport.go
func (o *V2RayTransportOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Type = stringNotNil(fj.GetStringBytes("type"))
	switch o.Type {
	case C.V2RayTransportTypeHTTP:
		o.HTTPOptions.unmarshalFastJSON(fj)
	case C.V2RayTransportTypeWebsocket:
		o.WebsocketOptions.unmarshalFastJSON(fj)
	case C.V2RayTransportTypeQUIC:
		o.QUICOptions.unmarshalFastJSON(fj)
	case C.V2RayTransportTypeGRPC:
		o.GRPCOptions.unmarshalFastJSON(fj)
	case C.V2RayTransportTypeHTTPUpgrade:
		o.HTTPUpgradeOptions.unmarshalFastJSON(fj)
	default:
		E.New("unknown transport type: " + o.Type)
	}
}
func (o *V2RayHTTPOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Host = unmarshalFastJSONListableString(fj.Get("host"))
	o.Path = stringNotNil(fj.GetStringBytes("path"))
	o.Method = stringNotNil(fj.GetStringBytes("method"))
	o.Headers = unmarshalFastJSONMapHTTPHeader(fj.GetObject("headers"))
	o.IdleTimeout = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("idle_timeout")))
	o.PingTimeout = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("ping_timeout")))
}
func (o *V2RayWebsocketOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Path = stringNotNil(fj.GetStringBytes("path"))
	o.Headers = unmarshalFastJSONMapHTTPHeader(fj.GetObject("headers"))
	o.MaxEarlyData = uint32(fj.GetUint("max_early_data"))
	o.EarlyDataHeaderName = stringNotNil(fj.GetStringBytes("early_data_header_name"))
}
func (o *V2RayQUICOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

}
func (o *V2RayGRPCOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.ServiceName = stringNotNil(fj.GetStringBytes("service_name"))
	o.IdleTimeout = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("idle_timeout")))
	o.PingTimeout = unmarshalFastJSONDuration(stringNotNil(fj.GetStringBytes("ping_timeout")))
	o.PermitWithoutStream = fj.GetBool("permit_without_stream")
	//o.ForceLite        = fj.GetBool(""-")
}
func (o *V2RayHTTPUpgradeOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Host = stringNotNil(fj.GetStringBytes("host"))
	o.Path = stringNotNil(fj.GetStringBytes("path"))
	o.Headers = unmarshalFastJSONMapHTTPHeader(fj.GetObject("headers"))
}

// v2ray.go
func unmarshalFastJSONArrayVLESSUser(fj *fastjson.Value) []VLESSUser {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}
	arr := fj.GetArray()
	if arr != nil {
		list := make([]VLESSUser, len(arr))
		for i, v := range arr {
			vv := VLESSUser{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]VLESSUser, 1)
	vv := VLESSUser{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

func (o *V2RayAPIOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Listen = stringNotNil(fj.GetStringBytes("listen"))
	stats := fj.Get("stats")
	if stats != nil && stats.Type() != fastjson.TypeNull {
		o.Stats = &V2RayStatsServiceOptions{}
		o.Stats.unmarshalFastJSON(stats)
	}
}

func (o *CacheFileOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Enabled = fj.GetBool("enabled")
	o.Path = stringNotNil(fj.GetStringBytes("path"))
	o.CacheID = stringNotNil(fj.GetStringBytes("cache_id"))
	o.StoreFakeIP = fj.GetBool("store_fakeip")
}

func (o *V2RayStatsServiceOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.Enabled = fj.GetBool("enabled")
	o.Inbounds = unmarshalFastJSONArrayStringWithName(fj, "inbounds")
	o.Outbounds = unmarshalFastJSONArrayStringWithName(fj, "outbounds")
	o.Users = unmarshalFastJSONArrayStringWithName(fj, "users")
}

// vless.go
func (o *VLESSInboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.ListenOptions.unmarshalFastJSON(fj)
	o.Users = unmarshalFastJSONArrayVLESSUser(fj.Get("users"))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
	transport := fj.Get("transport")
	if transport != nil && transport.Type() != fastjson.TypeNull {
		o.Transport = &V2RayTransportOptions{}
		o.Transport.unmarshalFastJSON(transport)
	}

}
func (o *VLESSUser) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.Name = stringNotNil(fj.GetStringBytes("name"))
	o.UUID = stringNotNil(fj.GetStringBytes("uuid"))
	o.Flow = stringNotNil(fj.GetStringBytes("flow"))
}
func (o *VLESSOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.DialerOptions.unmarshalFastJSON(fj)
	o.ServerOptions.unmarshalFastJSON(fj)
	o.UUID = stringNotNil(fj.GetStringBytes("uuid"))
	o.Flow = stringNotNil(fj.GetStringBytes("flow"))
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
	multiplex := fj.Get("multiplex")
	if multiplex != nil && multiplex.Type() != fastjson.TypeNull {
		o.Multiplex = &OutboundMultiplexOptions{}
		o.Multiplex.unmarshalFastJSON(multiplex)
	}
	transport := fj.Get("transport")
	if transport != nil && transport.Type() != fastjson.TypeNull {
		o.Transport = &V2RayTransportOptions{}
		o.Transport.unmarshalFastJSON(transport)
	}
	packet_encoding := fj.Get("packet_encoding")
	if packet_encoding != nil && packet_encoding.Type() != fastjson.TypeNull {
		o.PacketEncoding = new(string)
		*o.PacketEncoding = stringNotNil(fj.GetStringBytes("packet_encoding"))
	}

}

// vmess.go
func (o *VMessOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}
	o.DialerOptions.unmarshalFastJSON(fj)
	o.ServerOptions.unmarshalFastJSON(fj)
	o.UUID = stringNotNil(fj.GetStringBytes("uuid"))
	o.Security = stringNotNil(fj.GetStringBytes("security"))
	o.AlterId = fj.GetInt("alter_id")
	o.GlobalPadding = fj.GetBool("global_padding")
	o.AuthenticatedLength = fj.GetBool("authenticated_length")
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.unmarshalFastJSON(tls)
	}
	o.PacketEncoding = stringNotNil(fj.GetStringBytes("packet_encoding"))

	multiplex := fj.Get("multiplex")
	if multiplex != nil && multiplex.Type() != fastjson.TypeNull {
		o.Multiplex = &OutboundMultiplexOptions{}
		o.Multiplex.unmarshalFastJSON(multiplex)
	}
	transport := fj.Get("transport")
	if transport != nil && transport.Type() != fastjson.TypeNull {
		o.Transport = &V2RayTransportOptions{}
		o.Transport.unmarshalFastJSON(transport)
	}
}

// wireguard.go
func (o *WireGuardOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.DialerOptions.unmarshalFastJSON(fj)
	o.SystemInterface = fj.GetBool("system_interface")
	o.GSO = fj.GetBool("gso")
	o.InterfaceName = stringNotNil(fj.GetStringBytes("interface_name"))
	o.LocalAddress = unmarshalFastJSONListableNetipPrefix(fj.Get("local_address"))
	o.PrivateKey = stringNotNil(fj.GetStringBytes("private_key"))
	o.Peers = unmarshalFastJSONArrayWireGuardPeer(fj.Get("peers"))
	o.ServerOptions.unmarshalFastJSON(fj)
	o.PeerPublicKey = stringNotNil(fj.GetStringBytes("peer_public_key"))
	o.PreSharedKey = stringNotNil(fj.GetStringBytes("pre_shared_key"))
	o.Reserved = unmarshalFastJSONListableUint8(fj.Get("reserved"))
	o.Workers = fj.GetInt("workers")
	o.MTU = uint32(fj.GetUint("mtu"))
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
	turn_relay := fj.Get("turn_relay")
	if turn_relay != nil && turn_relay.Type() != fastjson.TypeNull {
		o.TurnRelay = &TurnRelayOptions{}
		o.TurnRelay.unmarshalFastJSON(turn_relay)
	}
	o.FakePackets = stringNotNil(fj.GetStringBytes("fake_packets"))
	o.FakePacketsSize = stringNotNil(fj.GetStringBytes("fake_packets_size"))
	o.FakePacketsDelay = stringNotNil(fj.GetStringBytes("fake_packets_delay"))
	o.FakePacketsMode = stringNotNil(fj.GetStringBytes("fake_packets_mode"))
}
func (o *WireGuardPeer) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.ServerOptions.unmarshalFastJSON(fj)
	o.PublicKey = stringNotNil(fj.GetStringBytes("public_key"))
	o.PreSharedKey = stringNotNil(fj.GetStringBytes("pre_shared_key"))
	o.AllowedIPs = unmarshalFastJSONListableString(fj.Get("allowed_ips"))
	o.Reserved = unmarshalFastJSONListableUint8(fj.Get("reserved"))
}

func unmarshalFastJSONArrayWireGuardPeer(fj *fastjson.Value) []WireGuardPeer {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return make([]WireGuardPeer, 0)
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make([]WireGuardPeer, len(arr))
		for i, v := range arr {
			vv := WireGuardPeer{}
			vv.unmarshalFastJSON(v)
			list[i] = vv
		}
		return list
	}
	if fj.Type() != fastjson.TypeObject {
		return nil
	}
	list := make([]WireGuardPeer, 1)
	vv := WireGuardPeer{}
	vv.unmarshalFastJSON(fj.Get())
	list[0] = vv
	return list
}

// xray.go
func unmarshalFastJSON(fj *fastjson.Value)*conf.Fragment {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	fragment :=  &conf.Fragment{}
	fragment.Packets = stringNotNil(fj.GetStringBytes("packets"))
	fragment.Length = stringNotNil(fj.GetStringBytes("length"))
	fragment.Interval = stringNotNil(fj.GetStringBytes("interval"))
	fragment.Host1_header = stringNotNil(fj.GetStringBytes("host1_header"))
	fragment.Host1_domain = stringNotNil(fj.GetStringBytes("host1_domain"))
	fragment.Host2_header = stringNotNil(fj.GetStringBytes("host2_header"))
	fragment.Host2_domain = stringNotNil(fj.GetStringBytes("host2_domain"))
	return fragment
}

func (o *XrayOutboundOptions) unmarshalFastJSON(fj *fastjson.Value) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return
	}

	o.DialerOptions.unmarshalFastJSON(fj)
	o.Network = NetworkList(stringNotNil(fj.GetStringBytes("network")))
	udp_over_tcp := fj.Get("udp_over_tcp")
	xray_outbound_raw := fj.GetObject("xray_outbound_raw")
	xray_fragment := fj.Get("xray_fragment")
	if udp_over_tcp != nil && udp_over_tcp.Type() == fastjson.TypeObject {
		o.UDPOverTCP = &UDPOverTCPOptions{}
		o.UDPOverTCP.unmarshalFastJSON(udp_over_tcp)
	}
	if xray_outbound_raw != nil {
		map_data := make(map[string]any)
		xray_outbound_raw.Visit(func(key []byte, value *fastjson.Value) {
			map_data[string(key)] = value
		})
		o.XrayOutboundJson = &map_data
	}
	if xray_fragment != nil && xray_fragment.Type() != fastjson.TypeNull {
		o.Fragment = unmarshalFastJSON(xray_fragment)
	}
	o.LogLevel = stringNotNil(fj.GetStringBytes("xray_loglevel"))
}


 
func unmarshalFastJSONListableNetipPrefix(fj *fastjson.Value) Listable[netip.Prefix] {
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return make(Listable[netip.Prefix], 0)
	}

	arr := fj.GetArray()
	if arr != nil {
		list := make(Listable[netip.Prefix], len(arr))
		for i, v := range arr {
			by, err := v.StringBytes()
			if err == nil {
				vv, err := netip.ParsePrefix(stringNotNil(by))
				if err == nil {
					list[i] = vv
				}
			}
		}
		return list
	}
	if fj.Type() != fastjson.TypeString {
		return nil
	}
	by, err := fj.StringBytes()
	if err == nil {
		vv, err := netip.ParsePrefix(stringNotNil(by))
		if err == nil {
			list := make(Listable[netip.Prefix], 1)
			list[0] = vv
			return list
		}
	}

	return make(Listable[netip.Prefix], 0)
}

func (o *TLSFragmentOptions) unmarshalFastJSON(fj *fastjson.Value) {
	o.Enabled = fj.GetBool("enabled")
	o.Size = stringNotNil(fj.GetStringBytes("size"))
	o.Sleep = stringNotNil(fj.GetStringBytes("sleep"))
}

func (o *TurnRelayOptions) unmarshalFastJSON(fj *fastjson.Value) {
	o.ServerOptions.unmarshalFastJSON(fj)
	o.Username = stringNotNil(fj.GetStringBytes("username"))
	o.Password = stringNotNil(fj.GetStringBytes("password"))
	o.Realm = stringNotNil(fj.GetStringBytes("realm"))
}

func (o *TLSTricksOptions) unmarshalFastJSON(fj *fastjson.Value) {
	o.MixedCaseSNI = fj.GetBool("mixedcase_sni")
	o.PaddingMode = stringNotNil(fj.GetStringBytes("padding_mode"))
	o.PaddingSize = stringNotNil(fj.GetStringBytes("padding_size"))
	o.PaddingSNI = stringNotNil(fj.GetStringBytes("padding_sni"))
}