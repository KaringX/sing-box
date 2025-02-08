//go:build with_fastjson

package option

//karing
import (
	"net/netip"
	"time"

	C "github.com/sagernet/sing-box/constant"
	dns "github.com/sagernet/sing-dns"
	"github.com/sagernet/sing/common/auth"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json/badoption"
	"github.com/valyala/fastjson"
)

type FastjsonParseErrorCallbackFunc func(content string, err string)
var (
	FastjsonParseError FastjsonParseErrorCallbackFunc
)

var FastjsonErrorNil = E.New("nil")

func FastjsonHandleError(content string, err string) {
	if FastjsonParseError == nil {
		return
	}
	FastjsonParseError(content, err)
}

func FastjsonStringNotNil(v []byte) string {
	if v == nil {
		return ""
	}
	return string(v)
}

func FastjsonUnmarshalInt(fj *fastjson.Value, name string) (int, error) {
	if len(name) > 0 {
		return fj.GetInt(name), nil
	}
	return fj.GetInt(), nil
}

func FastjsonUnmarshalIntPtr(fj *fastjson.Value, name string) (*int, error) {
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return nil, FastjsonErrorNil
	}
	v, err := value.Int()
	if err != nil {
		return nil, err
	}
	nvalue := new(int)
    *nvalue = v
	return nvalue, nil
}

func FastjsonUnmarshalInt32(fj *fastjson.Value, name string) (int32, error) {
	if len(name) > 0 {
		return int32(fj.GetInt(name)), nil
	}
	return int32(fj.GetInt()), nil
}

func FastjsonUnmarshalUint32(fj *fastjson.Value, name string) (uint32, error) {
	if len(name) > 0 {
		return uint32(fj.GetUint(name)), nil
	}
	return uint32(fj.GetUint()), nil
}

func FastjsonUnmarshalUint32Ptr(fj *fastjson.Value, name string) (*uint32, error) {
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return nil, FastjsonErrorNil
	}
	v, err := value.Uint()
	if err != nil {
		return nil, err
	}
	nvalue := new(uint32)
    *nvalue = uint32(v)
	return nvalue, nil
}

func FastjsonUnmarshalUint16(fj *fastjson.Value, name string) (uint16, error) {
	if len(name) > 0 {
		return uint16(fj.GetInt(name)), nil
	}
	return uint16(fj.GetInt()), nil
}

func FastjsonUnmarshalUint8(fj *fastjson.Value, name string) (uint8, error) {
	if len(name) > 0 {
		return uint8(fj.GetInt(name)), nil
	}
	return uint8(fj.GetInt()), nil
}

func FastjsonUnmarshalBool(fj *fastjson.Value, name string) (bool, error) {
	if len(name) > 0 {
		return bool(fj.GetBool(name)), nil
	}
	return bool(fj.GetBool()), nil
}

func FastjsonUnmarshalBoolPtr(fj *fastjson.Value, name string) (*bool, error) {
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return nil, FastjsonErrorNil
	}
	v, err := value.Bool()
	if err != nil {
		return nil, err
	}
	nvalue := new(bool)
    *nvalue = bool(v)
	return nvalue, nil
}

func FastjsonUnmarshalString(fj *fastjson.Value, name string) (string, error) {
	var value []byte
	if len(name) > 0 {
		value = fj.GetStringBytes(name)
	} else {
		value = fj.GetStringBytes()
	}
	if value == nil {
		return "", FastjsonErrorNil
	}
	return string(value), nil
}

func FastjsonUnmarshalStringPtr(fj *fastjson.Value, name string) (*string, error) {
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return nil, FastjsonErrorNil
	}
	v, err := value.StringBytes()
	if err != nil {
		return nil, err
	}
	nvalue := new(string)
    *nvalue = string(v)
	return nvalue, nil
}

func FastjsonUnmarshal[T struct{}](fj *fastjson.Value, name string, fn func(fj *fastjson.Value) (T, error)) (T, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return T{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return  T{}, nil
	}

	return fn(value)
}

//[T int|int32|string|struct{}]
func FastjsonUnmarshalArrayT[T any](fj *fastjson.Value, name string, fallbackType fastjson.Type, fn func(fj *fastjson.Value, name string) (T, error)) []T { 
	if fj == nil || fj.Type() == fastjson.TypeNull {
		return nil
	}
	var arr []*fastjson.Value
	if len(name) > 0 {
		arr = fj.GetArray(name)
	} else {
		arr = fj.GetArray()
	}
	if arr != nil {
		list := make([]T, len(arr))
		for i, v := range arr {
			value, err := fn(v, "")
			if err == nil {
				list[i] = value
			}
		}
		return list
	}
	if fj.Type() != fallbackType {
		return nil
	}
	list := make([]T, 1)
	value, err := fn(fj.Get(), "")
	if err == nil {
		list[0] = value
	}
	return list
}

func FastjsonUnmarshalMapStringString(fj *fastjson.Object) map[string][]string {
	if fj == nil {
		return nil
	}

	protoMap := make(map[string][]string)
	fj.Visit(func(key []byte, value *fastjson.Value) {
		protoMap[string(key)] = FastjsonUnmarshalArrayString(value, "")
	})

	return protoMap
}

func FastjsonUnmarshalArrayString(fj *fastjson.Value, name string) []string {
	return FastjsonUnmarshalArrayT(fj, name, fastjson.TypeString, FastjsonUnmarshalString)
}

func FastjsonUnmarshalDuration(fj *fastjson.Value, name string) (time.Duration, error) {
	value := FastjsonStringNotNil(fj.GetStringBytes(name))
	if len(value) == 0 {
		return 0, nil
	}
	duration, err := time.ParseDuration(value)
	if err == nil {
		return 0, err
	}
	return duration, nil
}

//badoption
func FastjsonUnmarshalBadoptionDuration(fj *fastjson.Value, name string) (badoption.Duration, error) {
	value, ok := FastjsonUnmarshalDuration(fj, name)
	return badoption.Duration(value), ok
}

func FastjsonUnmarshalListableT[T any](fj *fastjson.Value, fallbackType fastjson.Type, fn func(fj *fastjson.Value, name string) (T, error)) badoption.Listable[T] {
	return badoption.Listable[T](FastjsonUnmarshalArrayT[T](fj, "", fallbackType, fn))
}

func FastjsonUnmarshalListableString(fj *fastjson.Value) badoption.Listable[string] {
	return FastjsonUnmarshalListableT(fj, fastjson.TypeString, FastjsonUnmarshalString)
}

func FastjsonUnmarshalListableInt(fj *fastjson.Value) badoption.Listable[int] {
	return FastjsonUnmarshalListableT(fj, fastjson.TypeNumber, FastjsonUnmarshalInt)
}

func FastjsonUnmarshalListableInt32(fj *fastjson.Value) badoption.Listable[int32] {
	return FastjsonUnmarshalListableT(fj, fastjson.TypeNumber, FastjsonUnmarshalInt32)
}

func FastjsonUnmarshalListableUInt32(fj *fastjson.Value) badoption.Listable[uint32] {
	return FastjsonUnmarshalListableT(fj, fastjson.TypeNumber, FastjsonUnmarshalUint32)
}

func FastjsonUnmarshalListableUint16(fj *fastjson.Value) badoption.Listable[uint16] {
	return FastjsonUnmarshalListableT(fj, fastjson.TypeNumber, FastjsonUnmarshalUint16)
}

func FastjsonUnmarshalListableUint8(fj *fastjson.Value) badoption.Listable[uint8] {
	return FastjsonUnmarshalListableT(fj, fastjson.TypeNumber, FastjsonUnmarshalUint8)
}

func FastjsonUnmarshalDNSQueryType(fj *fastjson.Value, name string) (DNSQueryType, error) {
	if len(name) > 0 {
		return DNSQueryType(fj.GetInt(name)), nil
	}
	return DNSQueryType(fj.GetInt()), nil
}

func FastjsonUnmarshalListableDNSQueryType(fj *fastjson.Value) badoption.Listable[DNSQueryType] {
	return FastjsonUnmarshalListableT(fj, fastjson.TypeNumber, FastjsonUnmarshalDNSQueryType)
}

func FastjsonUnmarshalHTTPHeader(fj *fastjson.Value, name string) badoption.HTTPHeader {
	value := fj.GetObject(name)
	if value == nil {
		return nil
	}
	list := make(badoption.HTTPHeader, value.Len())
	value.Visit(func(key []byte, value *fastjson.Value) {
		list[string(key)] = FastjsonUnmarshalArrayString(value, "")
	})

	return list
}

func FastjsonUnmarshalDomainStrategy(fj *fastjson.Value, name string) (DomainStrategy, bool) {
	value := FastjsonStringNotNil(fj.GetStringBytes(name))
	switch value {
	case "", "as_is":
		return DomainStrategy(dns.DomainStrategyAsIS), true
	case "prefer_ipv4":
		return DomainStrategy(dns.DomainStrategyPreferIPv4), true
	case "prefer_ipv6":
		return DomainStrategy(dns.DomainStrategyPreferIPv6), true
	case "ipv4_only":
		return DomainStrategy(dns.DomainStrategyUseIPv4), true
	case "ipv6_only":
		return DomainStrategy(dns.DomainStrategyUseIPv6), true
	default:
		return DomainStrategy(dns.DomainStrategyPreferIPv4), true
	}
}

// clash.go
func (o *ClashAPIOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.ExternalController, _ = FastjsonUnmarshalString(fj, "external_controller")
	o.ExternalUI, _ = FastjsonUnmarshalString(fj, "external_ui")
	o.ExternalUIDownloadURL, _ = FastjsonUnmarshalString(fj, "external_ui_download_url")
	o.ExternalUIDownloadDetour, _ = FastjsonUnmarshalString(fj, "external_ui_download_detour")
	o.Secret, _ = FastjsonUnmarshalString(fj, "secret")
	o.DefaultMode, _= FastjsonUnmarshalString(fj, "default_mode")
	return nil
}

func (o *SelectorOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Outbounds = FastjsonUnmarshalArrayString(fj, "outbounds")
	o.Default, _ = FastjsonUnmarshalString(fj, "default")
	o.InterruptExistConnections, _ = FastjsonUnmarshalBool(fj, "interrupt_exist_connections")
	return nil
}

func (o *URLTestOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Outbounds = FastjsonUnmarshalArrayString(fj, "outbounds")
	o.URL , _= FastjsonUnmarshalString(fj, "url")
	o.Interval, _ = FastjsonUnmarshalBadoptionDuration(fj, "interval")
	o.Tolerance, _ = FastjsonUnmarshalUint16(fj, "tolerance")
	o.IdleTimeout, _ = FastjsonUnmarshalBadoptionDuration(fj, "idle_timeout")
	o.InterruptExistConnections, _ = FastjsonUnmarshalBool(fj, "interrupt_exist_connections")
	o.Default, _ = FastjsonUnmarshalString(fj, "default")
	o.ReTestIfNetworkUpdate, _ = FastjsonUnmarshalBool(fj, "retest_if_network_udpate")
	return nil
}

// config.go
func (o *Options) FastjsonUnmarshal(content []byte) error {
	var parser fastjson.Parser
	jp, err := parser.ParseBytes(content)
	if err != nil {
		return err
	}

	o.FastjsonUnmarshal(jp)
	return nil
}

func (o *Options) FastjsonUnmarshal(fj *fastjson.Value) error{
	o.Schema, _ = FastjsonUnmarshalString(fj, "schema")
	log := fj.Get("log")
	if log != nil {
		o.Log = &LogOptions{}
		o.Log.FastjsonUnmarshal(log)
	}
	dns := fj.Get("dns")
	if dns != nil  {
		o.DNS = &DNSOptions{}
		o.DNS.FastjsonUnmarshal(dns)
	}
	ntp := fj.Get("ntp")
	if ntp != nil   {
		o.NTP = &NTPOptions{}
		o.NTP.FastjsonUnmarshal(ntp)
	}
	o.Inbounds = FastjsonUnmarshalArrayInbound(fj.Get("inbounds"))
	o.Outbounds = FastjsonUnmarshalArrayOutbound(fj.Get("outbounds"))
	route := fj.Get("route")
	if route != nil  {
		o.Route = &RouteOptions{}
		o.Route.FastjsonUnmarshal(route)
	}
	experimental := fj.Get("experimental")
	if experimental != nil   {
		o.Experimental = &ExperimentalOptions{}
		o.Experimental.FastjsonUnmarshal(experimental)
	}
	return nil
}

func (o *LogOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Disabled, _ = FastjsonUnmarshalBool(fj, "disabled")
	o.Level, _ = FastjsonUnmarshalString(fj, "level")
	o.Output, _ = FastjsonUnmarshalString(fj, "output")
	o.Timestamp, _ = FastjsonUnmarshalBool(fj, "timestamp")
	//o.DisableColor = FastjsonUnmarshalBool(fj, "-")
	return nil
}

// debug.go
func (o *DebugOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Listen, _ = FastjsonUnmarshalString(fj, "listen")
	o.GCPercent, _ = FastjsonUnmarshalIntPtr(fj, "gc_percent")
	o.MaxStack, _ = FastjsonUnmarshalIntPtr(fj, "max_stack")
	 
 
	o.MaxThreads, _ = FastjsonUnmarshalIntPtr(fj, "max_threads")
	 
	o.PanicOnFault, _ =FastjsonUnmarshalBoolPtr(fj, "panic_on_fault")

	o.TraceBack, _ = FastjsonUnmarshalString(fj, "trace_back")
	o.MemoryLimit = MemoryBytes(fj.GetInt64("memory_limit"))
	o.OOMKiller, _ = FastjsonUnmarshalBoolPtr(fj, "oom_killer")
	return nil
}

// direct.go
func (o *DirectInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.ListenOptions.FastjsonUnmarshal(fj)
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	o.OverrideAddress, _ = FastjsonUnmarshalString(fj, "override_address")
	o.OverridePort, _ = FastjsonUnmarshalUint16(fj, "override_port")
	return nil
}

func (o *DirectOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.DialerOptions.FastjsonUnmarshal(fj)
	o.OverrideAddress, _ = FastjsonUnmarshalString(fj, "override_address")
	o.OverridePort, _ = FastjsonUnmarshalUint16(fj, "override_port")
	o.ProxyProtocol = uint8(fj.GetUint("proxy_protocol"))
	return nil
}

// dns.go
func (o *DNSOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Servers = FastjsonUnmarshalArrayDNSServerOptions(fj.Get("servers"))
	o.Rules = FastjsonUnmarshalArrayDNSRule(fj.Get("rules"))
	fakeip := fj.Get("fakeip")
	o.Final, _ = FastjsonUnmarshalString(fj, "final")
	o.ReverseMapping, _ = FastjsonUnmarshalBool(fj, "reverse_mapping")
	if fakeip != nil && fakeip.Type() != fastjson.TypeNull {
		o.FakeIP = &DNSFakeIPOptions{}
		o.FakeIP.FastjsonUnmarshal(fakeip)
	}
	o.StaticIPs = FastjsonUnmarshalMapStringString(fj.GetObject("static_ips"))
	o.FastjsonUnmarshal(fj)
	o.DNSClientOptions.FastjsonUnmarshal(fj)
	return nil
}

func (o *DNSServerOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Tag , _= FastjsonUnmarshalString(fj, "tag")
	o.Address, _ = FastjsonUnmarshalString(fj, "address")
	o.Addresses = FastjsonUnmarshalArrayString(fj, "addresses")
	o.AddressResolver, _ = FastjsonUnmarshalString(fj, "address_resolver")
	o.AddressStrategy, _ = FastjsonUnmarshalDomainStrategy(fj, "address_strategy")
	o.AddressFallbackDelay, _ = FastjsonUnmarshalBadoptionDuration(fj, "address_fallback_delay")
	o.Strategy, _ = FastjsonUnmarshalDomainStrategy(fj, "strategy")
	o.Detour, _ = FastjsonUnmarshalString(fj, "detour")
	return nil
}

func (o *DNSClientOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Strategy, _ = FastjsonUnmarshalDomainStrategy(fj, "strategy")
	o.DisableCache, _ = FastjsonUnmarshalBool(fj, "disable_cache")
	o.DisableExpire, _ = FastjsonUnmarshalBool(fj, "disable_expire")
	o.IndependentCache, _ = FastjsonUnmarshalBool(fj, "independent_cache")
	return nil
}

func (o *DNSFakeIPOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	inet4Range, err := netip.ParsePrefix(FastjsonUnmarshalString(fj, "inet4_range"))
	if err == nil {
		o.Inet4Range = &inet4Range
	}

	inet6Range, err1 := netip.ParsePrefix(FastjsonUnmarshalString(fj, "inet6_range"))
	if err1 == nil {
		o.Inet6Range = &inet6Range
	}
	return nil
}

func FastjsonUnmarshaDNSServerOptions(fj *fastjson.Value, name string) (DNSServerOptions, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return DNSServerOptions{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return DNSServerOptions{}, nil
	}
	vv := DNSServerOptions{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayDNSServerOptions(fj *fastjson.Value) []DNSServerOptions {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshaDNSServerOptions)
}

func FastjsonUnmarshaDNSRule(fj *fastjson.Value, name string) (DNSRule, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return DNSRule{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return DNSRule{}, nil
	}
	vv := DNSRule{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayDNSRule(fj *fastjson.Value) []DNSRule {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshaDNSRule)
}

// experimental.go
func (o *ExperimentalOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	clash_api := fj.Get("clash_api")
	if clash_api != nil && clash_api.Type() != fastjson.TypeNull {
		o.ClashAPI = &ClashAPIOptions{}
		o.ClashAPI.FastjsonUnmarshal(clash_api)
	}
	v2ray_api := fj.Get("v2ray_api")
	if v2ray_api != nil && v2ray_api.Type() != fastjson.TypeNull {
		o.V2RayAPI = &V2RayAPIOptions{}
		o.V2RayAPI.FastjsonUnmarshal(v2ray_api)
	}
	cache_file := fj.Get("cache_file")
	if cache_file != nil && cache_file.Type() != fastjson.TypeNull {
		o.CacheFile = &CacheFileOptions{}
		o.CacheFile.FastjsonUnmarshal(cache_file)
	}
	debug := fj.Get("debug")
	if debug != nil && debug.Type() != fastjson.TypeNull {
		o.Debug = &DebugOptions{}
		o.Debug.FastjsonUnmarshal(debug)
	}
	return nil
}

// hysteria.go
func (o *HysteriaInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.ListenOptions.FastjsonUnmarshal(fj)
	o.Up, _ = FastjsonUnmarshalString(fj, "up")
	o.UpMbps, _ = FastjsonUnmarshalInt(fj, "up_mbps")
	o.Down, _ = FastjsonUnmarshalString(fj, "down")
	o.DownMbps, _ = FastjsonUnmarshalInt(fj, "down_mbps")
	o.Obfs, _ = FastjsonUnmarshalString(fj, "obfs")
	o.Users = FastjsonUnmarshalArrayHysteriaUser(fj.Get("users"))
	o.ReceiveWindowConn = fj.GetUint64("recv_window_conn")
	o.ReceiveWindowClient = fj.GetUint64("recv_window_client")
	o.MaxConnClient, _ = FastjsonUnmarshalInt(fj, "max_conn_client")
	o.DisableMTUDiscovery, _ = FastjsonUnmarshalBool(fj, "disable_mtu_discovery")
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	return nil
}

func (o *HysteriaUser) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	o.Auth = fj.GetStringBytes("auth")
	o.AuthString, _ = FastjsonUnmarshalString(fj, "auth_str")
	return nil
}

func (o *HysteriaOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.Up, _ = FastjsonUnmarshalString(fj, "up")
	o.UpMbps, _ = FastjsonUnmarshalInt(fj, "up_mbps")
	o.Down, _ = FastjsonUnmarshalString(fj, "down")
	o.DownMbps, _ = FastjsonUnmarshalInt(fj, "down_mbps")
	o.Obfs, _ = FastjsonUnmarshalString(fj, "obfs")
	o.Auth = fj.GetStringBytes("auth")
	o.AuthString, _ = FastjsonUnmarshalString(fj, "auth_str")
	o.ReceiveWindowConn = fj.GetUint64("recv_window_conn")
	o.ReceiveWindow = fj.GetUint64("recv_window")
	o.DisableMTUDiscovery, _ = FastjsonUnmarshalBool(fj, "disable_mtu_discovery")
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	turn_relay := fj.Get("turn_relay")
	if turn_relay != nil && turn_relay.Type() != fastjson.TypeNull {
		o.TurnRelay = &TurnRelayOptions{}
		o.TurnRelay.FastjsonUnmarshal(turn_relay)
	}
	o.HopPorts = FastjsonUnmarshalString(fj, "hop_ports")
	o.HopInterval = FastjsonUnmarshalInt(fj, "hop_interval")
	return nil
}

func FastjsonUnmarshaHysteriaUser(fj *fastjson.Value, name string) (HysteriaUser, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return HysteriaUser{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return HysteriaUser{}, nil
	}
	vv := HysteriaUser{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayHysteriaUser(fj *fastjson.Value) []HysteriaUser {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshaHysteriaUser)
}

// hysteria2.go
func (o *Hysteria2InboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.ListenOptions.FastjsonUnmarshal(fj)
	o.UpMbps, _= FastjsonUnmarshalInt(fj, "up_mbps")
	o.DownMbps, _ = FastjsonUnmarshalInt(fj, "down_mbps")
	obfs := fj.Get("obfs")
	if obfs != nil && obfs.Type() != fastjson.TypeNull {
		o.Obfs = &Hysteria2Obfs{}
		o.Obfs.FastjsonUnmarshal(obfs)
	}

	o.Users = FastjsonUnmarshalArrayHysteria2User(fj.Get("users"))
	o.IgnoreClientBandwidth, _ = FastjsonUnmarshalBool(fj, "ignore_client_bandwidth")
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	o.Masquerade = FastjsonUnmarshalString(fj, "masquerade")
	o.BrutalDebug, _ = FastjsonUnmarshalBool(fj, "brutal_debug")
	return nil
}

func (o *Hysteria2Obfs) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Type , _ = FastjsonUnmarshalString(fj, "type")
	o.Password , _= FastjsonUnmarshalString(fj, "password")
	return nil
}

func (o *Hysteria2User) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	o.Password , _= FastjsonUnmarshalString(fj, "password")
	return nil
}

func (o *Hysteria2OutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.UpMbps, _ = FastjsonUnmarshalInt(fj, "up_mbps")
	o.DownMbps, _ = FastjsonUnmarshalInt(fj, "down_mbps")
	obfs := fj.Get("obfs")
	if obfs != nil && obfs.Type() != fastjson.TypeNull {
		o.Obfs = &Hysteria2Obfs{}
		o.Obfs.FastjsonUnmarshal(obfs)
	}
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	o.BrutalDebug, _ = FastjsonUnmarshalBool(fj, "brutal_debug")
	turn_relay := fj.Get("turn_relay")
	if turn_relay != nil && turn_relay.Type() != fastjson.TypeNull {
		o.TurnRelay = &TurnRelayOptions{}
		o.TurnRelay.FastjsonUnmarshal(turn_relay)
	}
	o.HopPorts = HopPortsValue(FastjsonUnmarshalString(fj, "hop_ports"))
	o.HopInterval = FastjsonUnmarshalInt(fj, "hop_interval")
	return nil
}

func FastjsonUnmarshaHysteria2User(fj *fastjson.Value, name string) (Hysteria2User, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return Hysteria2User{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return Hysteria2User{}, nil
	}
	vv := Hysteria2User{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayHysteria2User(fj *fastjson.Value) []Hysteria2User {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshaHysteria2User)
}

// inbound.go
func (h *Inbound) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	h.Type, _ = FastjsonUnmarshalString(fj, "type")
	h.Tag, _ = FastjsonUnmarshalString(fj, "tag")
	switch h.Type {
	case C.TypeTun:
		h.TunOptions.FastjsonUnmarshal(fj)
	case C.TypeRedirect:
		h.RedirectOptions.FastjsonUnmarshal(fj)
	case C.TypeTProxy:
		h.TProxyOptions.FastjsonUnmarshal(fj)
	case C.TypeDirect:
		h.DirectOptions.FastjsonUnmarshal(fj)
	case C.TypeSOCKS:
		h.SocksOptions.FastjsonUnmarshal(fj)
	case C.TypeHTTP:
		h.HTTPOptions.FastjsonUnmarshal(fj)
	case C.TypeMixed:
		h.MixedOptions.FastjsonUnmarshal(fj)
	case C.TypeShadowsocks:
		h.ShadowsocksOptions.FastjsonUnmarshal(fj)
	case C.TypeVMess:
		h.VMessOptions.FastjsonUnmarshal(fj)
	case C.TypeTrojan:
		h.TrojanOptions.FastjsonUnmarshal(fj)
	case C.TypeNaive:
		h.NaiveOptions.FastjsonUnmarshal(fj)
	case C.TypeHysteria:
		h.HysteriaOptions.FastjsonUnmarshal(fj)
	case C.TypeShadowTLS:
		h.ShadowTLSOptions.FastjsonUnmarshal(fj)
	case C.TypeVLESS:
		h.VLESSOptions.FastjsonUnmarshal(fj)
	case C.TypeTUIC:
		h.TUICOptions.FastjsonUnmarshal(fj)
	case C.TypeHysteria2:
		h.Hysteria2Options.FastjsonUnmarshal(fj)
	default:
		E.New("unknown inbound type: ", h.Type, h.Tag)
		return nil
	}
	return nil
}

func (o *InboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.SniffEnabled, _ = FastjsonUnmarshalBool(fj, "sniff")
	o.SniffOverrideDestination, _ = FastjsonUnmarshalBool(fj, "sniff_override_destination")
	o.SniffTimeout, _ = FastjsonUnmarshalBadoptionDuration(fj, "sniff_timeout")
	o.DomainStrategy, _ = FastjsonUnmarshalDomainStrategy(fj, "domain_strategy")
	return nil
}

func (o *ListenOptions) FastjsonUnmarshal(fj *fastjson.Value) error {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	addr, err := netip.ParseAddr(FastjsonUnmarshalString(fj, "listen"))
	if err == nil {
		o.Listen = new(ListenAddress)
		*(o.Listen) = ListenAddress(addr)
	}

	o.ListenPort, _ = FastjsonUnmarshalUint16(fj, "listen_port")
	o.TCPFastOpen, _ = FastjsonUnmarshalBool(fj, "tcp_fast_open")
	o.TCPMultiPath, _ = FastjsonUnmarshalBool(fj, "tcp_multi_path")
	o.UDPFragment, _ =FastjsonUnmarshalBoolPtr(fj, "udp_fragment")
	//o.UDPFragmentDefault           = FastjsonUnmarshalBool(fj, "-")
	udpimeout, _ := FastjsonUnmarshalBadoptionDuration(fj, "udp_timeout")
	o.UDPTimeout = UDPTimeoutCompat(udpimeout)
	o.ProxyProtocol, _ = FastjsonUnmarshalBool(fj, "proxy_protocol")
	o.ProxyProtocolAcceptNoHeader, _ = FastjsonUnmarshalBool(fj, "proxy_protocol_accept_no_header")
	o.Detour, _ = FastjsonUnmarshalString(fj, "detour")
	o.InboundOptions.FastjsonUnmarshal(fj)
	return nil
}

func FastjsonUnmarshalInbound(fj *fastjson.Value, name string) (Inbound, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return Inbound{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return Inbound{}, nil
	}
	vv := Inbound{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}
func FastjsonUnmarshalArrayInbound(fj *fastjson.Value) []Inbound {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalInbound)
}

// naive.go
func (o *NaiveInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.ListenOptions.FastjsonUnmarshal(fj)
	o.Users = FastjsonUnmarshalArrayAuthUser(fj.Get("users"))
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	return nil
}

// ntp.go
func (o *NTPOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.Server, _ = FastjsonUnmarshalString(fj, "server")
	o.ServerPort, _ = FastjsonUnmarshalUint16(fj, "server_port")
	o.Interval, _ = FastjsonUnmarshalBadoptionDuration(fj, "interval")
	o.WriteToSystem, _ = FastjsonUnmarshalBool(fj, "write_to_system")
	o.DialerOptions.FastjsonUnmarshal(fj)
	return nil
}

// outbound.go
func (h *Outbound) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	h.Type, _ = FastjsonUnmarshalString(fj, "type")
	h.Tag, _ = FastjsonUnmarshalString(fj, "tag")
	switch h.Type {
	case C.TypeDirect:
		h.DirectOptions.FastjsonUnmarshal(fj)
	case C.TypeBlock, C.TypeDNS:
	case C.TypeSOCKS:
		h.SocksOptions.FastjsonUnmarshal(fj)
	case C.TypeHTTP:
		h.HTTPOptions.FastjsonUnmarshal(fj)
	case C.TypeShadowsocks:
		h.ShadowsocksOptions.FastjsonUnmarshal(fj)
	case C.TypeVMess:
		h.VMessOptions.FastjsonUnmarshal(fj)
	case C.TypeTrojan:
		h.TrojanOptions.FastjsonUnmarshal(fj)
	case C.TypeWireGuard:
		h.WireGuardOptions.FastjsonUnmarshal(fj)
	case C.TypeHysteria:
		h.HysteriaOptions.FastjsonUnmarshal(fj)
	case C.TypeTor:
		h.TorOptions.FastjsonUnmarshal(fj)
	case C.TypeSSH:
		h.SSHOptions.FastjsonUnmarshal(fj)
	case C.TypeShadowTLS:
		h.ShadowTLSOptions.FastjsonUnmarshal(fj)
	case C.TypeShadowsocksR:
		h.ShadowsocksROptions.FastjsonUnmarshal(fj)
	case C.TypeVLESS:
		h.VLESSOptions.FastjsonUnmarshal(fj)
	case C.TypeTUIC:
		h.TUICOptions.FastjsonUnmarshal(fj)
	case C.TypeHysteria2:
		h.Hysteria2Options.FastjsonUnmarshal(fj)
	case C.TypeSelector:
		h.SelectorOptions.FastjsonUnmarshal(fj)
	case C.TypeURLTest:
		h.URLTestOptions.FastjsonUnmarshal(fj)
	default:
		E.New("unknown outbound type: ", h.Type, h.Tag)
		return nil
	}
	return nil
}

func (o *DialerOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Detour, _ = FastjsonUnmarshalString(fj, "detour")
	o.BindInterface, _ = FastjsonUnmarshalString(fj, "bind_interface")
	addr4, err4 := netip.ParseAddr(FastjsonUnmarshalString(fj, "inet4_bind_address"))
	if err4 == nil {
		o.Inet4BindAddress = new(ListenAddress)
		*(o.Inet4BindAddress) = ListenAddress(addr4)
	}
	addr6, err6 := netip.ParseAddr(FastjsonUnmarshalString(fj, "inet6_bind_address"))
	if err6 == nil {
		o.Inet6BindAddress = new(ListenAddress)
		*(o.Inet6BindAddress) = ListenAddress(addr6)
	}
	o.ProtectPath, _ = FastjsonUnmarshalString(fj, "protect_path")
	o.RoutingMark, _ = FwMark(FastjsonUnmarshalUint32(fj, "routing_mark"))
	o.ReuseAddr, _ = FastjsonUnmarshalBool(fj, "reuse_addr")
	o.ConnectTimeout, _ = FastjsonUnmarshalBadoptionDuration(fj,"connect_timeout")
	o.TCPFastOpen, _ = FastjsonUnmarshalBool(fj, "tcp_fast_open")
	o.TCPMultiPath, _ = FastjsonUnmarshalBool(fj, "tcp_multi_path")
	o.UDPFragment, _ =FastjsonUnmarshalBoolPtr(fj, "udp_fragment")

	tls_fragment := fj.Get("tls_fragment")
	 
	//o.UDPFragmentDefault = FastjsonUnmarshalBool(fj, "-")
	o.DomainStrategy, _ = FastjsonUnmarshalDomainStrategy(fj,"domain_strategy")
	o.FallbackDelay, _ = FastjsonUnmarshalBadoptionDuration(fj,"fallback_delay")
	if tls_fragment != nil && tls_fragment.Type() != fastjson.TypeNull {
		o.TLSFragment = &TLSFragmentOptions{}
		o.TLSFragment.FastjsonUnmarshal(tls_fragment)
	}
	return nil
}

func (o *ServerOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Server, _ = FastjsonUnmarshalString(fj, "server")
	o.ServerPort, _ = FastjsonUnmarshalUint16(fj, "server_port")
	return nil
}

func (o *OutboundMultiplexOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.Protocol, _ = FastjsonUnmarshalString(fj, "protocol")
	o.MaxConnections, _ = FastjsonUnmarshalInt(fj, "max_connections")
	o.MinStreams, _ = FastjsonUnmarshalInt(fj, "min_streams")
	o.MaxStreams, _ = FastjsonUnmarshalInt(fj, "max_streams")
	o.Padding, _ = FastjsonUnmarshalBool(fj, "padding")
	return nil
}

func FastjsonUnmarshalOutbound(fj *fastjson.Value, name string) (Outbound, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return Outbound{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return Outbound{}, nil
	}
	vv := Outbound{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayOutbound(fj *fastjson.Value) []Outbound {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalOutbound)
}

func FastjsonUnmarshalAuthUser2(o *auth.User, fj *fastjson.Value) error {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Username, _ = FastjsonUnmarshalString(fj, "username")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	return nil
}

func FastjsonUnmarshalAuthUser(fj *fastjson.Value, name string) (auth.User, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return auth.User{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return auth.User{}, nil
	}
	vv := auth.User{}
	err := FastjsonUnmarshalAuthUser2(&vv, value)
	return vv, err
}

func FastjsonUnmarshalArrayAuthUser(fj *fastjson.Value) []auth.User {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalAuthUser)
}

// platform.go
func (o *OnDemandOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.Rules = FastjsonUnmarshalArrayOnDemandRule(fj.Get("rules"))
	return nil
}

func (o *OnDemandRule) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	action := fj.Get("action")
	if action != nil && action.Type() != fastjson.TypeNull {
		o.Action = new(OnDemandRuleAction)
		*o.Action = OnDemandRuleAction(FastjsonUnmarshalInt(fj, "action"))
	}
	o.DNSSearchDomainMatch = FastjsonUnmarshalListableString(fj.Get("dns_search_domain_match"))
	o.DNSServerAddressMatch = FastjsonUnmarshalListableString(fj.Get("dns_server_address_match"))
	interface_type_match := fj.Get("interface_type_match")
	if interface_type_match != nil && interface_type_match.Type() != fastjson.TypeNull {
		o.InterfaceTypeMatch = new(OnDemandRuleInterfaceType)
		*o.InterfaceTypeMatch = OnDemandRuleInterfaceType(FastjsonUnmarshalInt(fj, "interface_type_match"))
	}

	o.SSIDMatch = FastjsonUnmarshalListableString(fj.Get("ssid_match"))
	o.ProbeURL, _ = FastjsonUnmarshalString(fj, "probe_url")
	return nil
}

func FastjsonUnmarshalDefaultOnDemandRule(fj *fastjson.Value, name string) (OnDemandRule, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return OnDemandRule{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return OnDemandRule{}, nil
	}
	vv := OnDemandRule{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayOnDemandRule(fj *fastjson.Value) []OnDemandRule {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalDefaultOnDemandRule)
}

// redir.go
func (o *RedirectInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.ListenOptions.FastjsonUnmarshal(fj)
	return nil
}

func (o *TProxyInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.ListenOptions.FastjsonUnmarshal(fj)
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	return nil
}

// route.go
func (o *RouteOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	geoip := fj.Get("geoip")
	if geoip != nil && geoip.Type() != fastjson.TypeNull {
		o.GeoIP = &GeoIPOptions{}
		o.GeoIP.FastjsonUnmarshal(geoip)
	}
	geosite := fj.Get("geosite")
	if geosite != nil && geosite.Type() != fastjson.TypeNull {
		o.Geosite = &GeositeOptions{}
		o.Geosite.FastjsonUnmarshal(geosite)
	}

	o.Rules = FastjsonUnmarshalArrayRule(fj.Get("rules"))
	o.RuleSet = FastjsonUnmarshalArrayRuleSet(fj.Get("rule_set"))
	o.Final, _ = FastjsonUnmarshalString(fj, "final")
	o.FindProcess, _ = FastjsonUnmarshalBool(fj, "find_process")
	o.AutoDetectInterface, _ = FastjsonUnmarshalBool(fj, "auto_detect_interface")
	o.OverrideAndroidVPN, _ = FastjsonUnmarshalBool(fj, "override_android_vpn")
	o.DefaultInterface, _ = FastjsonUnmarshalString(fj, "default_interface")
	o.DefaultMark, _ = FwMark(FastjsonUnmarshalUint32(fj, "default_mark"))
	return nil
}

func (o *GeoIPOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Path, _ = FastjsonUnmarshalString(fj, "path")
	o.DownloadURL, _ = FastjsonUnmarshalString(fj, "download_url")
	o.DownloadDetour, _ = FastjsonUnmarshalString(fj, "download_detour")
	return nil
}

func (o *GeositeOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Path, _ = FastjsonUnmarshalString(fj, "path")
	o.DownloadURL, _ = FastjsonUnmarshalString(fj, "download_url")
	o.DownloadDetour, _ = FastjsonUnmarshalString(fj, "download_detour")
	return nil
}

// rule_dns.go
func (o *DNSRule) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Type, _ = FastjsonUnmarshalString(fj, "type")
	switch o.Type {
	case "", C.RuleTypeDefault:
		o.Type = C.RuleTypeDefault
		o.DefaultOptions.FastjsonUnmarshal(fj)
	case C.RuleTypeLogical:
		o.LogicalOptions.FastjsonUnmarshal(fj)
	}
	return nil
}

func (o *DefaultDNSRule) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Inbound = FastjsonUnmarshalListableString(fj.Get("inbound"))
	o.IPVersion, _ = FastjsonUnmarshalInt(fj, "ip_version")
	o.QueryType = FastjsonUnmarshalListableDNSQueryType(fj.Get("query_type"))
	o.Network = FastjsonUnmarshalListableString(fj.Get("network"))
	o.AuthUser = FastjsonUnmarshalListableString(fj.Get("auth_user"))
	o.Protocol = FastjsonUnmarshalListableString(fj.Get("protocol"))
	o.Domain = FastjsonUnmarshalListableString(fj.Get("domain"))
	o.DomainSuffix = FastjsonUnmarshalListableString(fj.Get("domain_suffix"))
	o.DomainKeyword = FastjsonUnmarshalListableString(fj.Get("domain_keyword"))
	o.DomainRegex = FastjsonUnmarshalListableString(fj.Get("domain_regex"))
	o.Geosite = FastjsonUnmarshalListableString(fj.Get("geosite"))
	o.SourceGeoIP = FastjsonUnmarshalListableString(fj.Get("source_geoip"))
	//todo
	//o.GeoIP = FastjsonUnmarshalListableString(fj.Get("geoip"))
	//o.IPCIDR = FastjsonUnmarshalListableString(fj.Get("ip_cidr"))
	//o.IPIsPrivate = FastjsonUnmarshalBool(fj, "ip_is_private")
	o.SourceIPCIDR = FastjsonUnmarshalListableString(fj.Get("source_ip_cidr"))
	o.SourceIPIsPrivate, _ = FastjsonUnmarshalBool(fj, "source_ip_is_private")
	o.SourcePort = FastjsonUnmarshalListableUint16(fj.Get("source_port"))
	o.SourcePortRange = FastjsonUnmarshalListableString(fj.Get("source_port_range"))
	o.Port = FastjsonUnmarshalListableUint16(fj.Get("port"))
	o.PortRange = FastjsonUnmarshalListableString(fj.Get("port_range"))
	o.ProcessName = FastjsonUnmarshalListableString(fj.Get("process_name"))
	o.ProcessPath = FastjsonUnmarshalListableString(fj.Get("process_path"))
	o.PackageName = FastjsonUnmarshalListableString(fj.Get("package_name"))
	o.User = FastjsonUnmarshalListableString(fj.Get("user"))
	o.UserID = FastjsonUnmarshalListableInt32(fj.Get("user_id"))
	o.Outbound = FastjsonUnmarshalListableString(fj.Get("outbound"))
	o.ClashMode, _ = FastjsonUnmarshalString(fj, "clash_mode")
	o.WIFISSID = FastjsonUnmarshalListableString(fj.Get("wifi_ssid"))
	o.WIFIBSSID = FastjsonUnmarshalListableString(fj.Get("wifi_bssid"))
	o.RuleSet = FastjsonUnmarshalListableString(fj.Get("rule_set"))
	o.Invert, _ = FastjsonUnmarshalBool(fj, "invert")
	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	return nil
}

func (o *LogicalDNSRule) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Mode, _ = FastjsonUnmarshalString(fj, "mode")
	o.Rules = FastjsonUnmarshalArrayDNSRule(fj.Get("rules"))
	o.Invert, _ = FastjsonUnmarshalBool(fj, "invert")
	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	return nil
}

func (o *RuleSet) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	//LocalOptions  LocalRuleSet  `json:"-"`
	//RemoteOptions RemoteRuleSet `json:"-"`

	o.Type, _ = FastjsonUnmarshalString(fj, "type")
	o.Tag, _ = FastjsonUnmarshalString(fj, "tag")
	o.Format , _= FastjsonUnmarshalString(fj, "format")
	o.LocalOptions.IsAsset, _ = FastjsonUnmarshalBool(fj, "is_asset")
	o.LocalOptions.Path, _ = FastjsonUnmarshalString(fj, "path")
	o.RemoteOptions.URL, _ = FastjsonUnmarshalString(fj, "url")
	o.RemoteOptions.DownloadDetour, _ = FastjsonUnmarshalString(fj, "download_detour")
	o.RemoteOptions.UpdateInterval, _ = FastjsonUnmarshalBadoptionDuration(fj, "update_interval")
	return nil
}
func FastjsonUnmarshalDefaultRuleSet(fj *fastjson.Value, name string) (RuleSet, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return RuleSet{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return RuleSet{}, nil
	}
	vv := RuleSet{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayRuleSet(fj *fastjson.Value) []RuleSet {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalDefaultRuleSet)
}

func FastjsonUnmarshalDefaultRule(fj *fastjson.Value, name string) (Rule, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return Rule{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return Rule{}, nil
	}
	vv := Rule{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayRule(fj *fastjson.Value) []Rule {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalDefaultRule)
}

func FastjsonUnmarshalDefaultDNSRule(fj *fastjson.Value, name string) (DefaultDNSRule, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return DefaultDNSRule{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return DefaultDNSRule{}, nil
	}
	vv := DefaultDNSRule{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func unmarshalFastJSONArrayDefaultDNSRule(fj *fastjson.Value) []DefaultDNSRule {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalDefaultDNSRule)
}

// rule.go
func (o *Rule) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Type, _ = FastjsonUnmarshalString(fj, "type")
	switch o.Type {
	case "", C.RuleTypeDefault:
		o.Type = C.RuleTypeDefault
		o.DefaultOptions.FastjsonUnmarshal(fj)
	case C.RuleTypeLogical:
		o.LogicalOptions.FastjsonUnmarshal(fj)
	default:
		E.New("unknown rule type: " + o.Type)
	}
	return nil
}

func (o *DefaultRule) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Inbound = FastjsonUnmarshalListableString(fj.Get("inbound"))
	o.IPVersion, _ = FastjsonUnmarshalInt(fj, "ip_version")
	o.Network = FastjsonUnmarshalListableString(fj.Get("network"))
	o.AuthUser = FastjsonUnmarshalListableString(fj.Get("auth_user"))
	o.Protocol = FastjsonUnmarshalListableString(fj.Get("protocol"))
	o.Domain = FastjsonUnmarshalListableString(fj.Get("domain"))
	o.DomainSuffix = FastjsonUnmarshalListableString(fj.Get("domain_suffix"))
	o.DomainKeyword = FastjsonUnmarshalListableString(fj.Get("domain_keyword"))
	o.DomainRegex = FastjsonUnmarshalListableString(fj.Get("domain_regex"))
	o.Geosite = FastjsonUnmarshalListableString(fj.Get("geosite"))
	o.SourceGeoIP = FastjsonUnmarshalListableString(fj.Get("source_geoip"))
	o.GeoIP = FastjsonUnmarshalListableString(fj.Get("geoip"))
	o.SourceIPCIDR = FastjsonUnmarshalListableString(fj.Get("source_ip_cidr"))
	o.SourceIPIsPrivate, _ = FastjsonUnmarshalBool(fj, "source_ip_is_private")
	o.IPCIDR = FastjsonUnmarshalListableString(fj.Get("ip_cidr"))
	o.IPIsPrivate, _ = FastjsonUnmarshalBool(fj, "ip_is_private")
	o.SourcePort = FastjsonUnmarshalListableUint16(fj.Get("source_port"))
	o.SourcePortRange = FastjsonUnmarshalListableString(fj.Get("source_port_range"))
	o.Port = FastjsonUnmarshalListableUint16(fj.Get("port"))
	o.PortRange = FastjsonUnmarshalListableString(fj.Get("port_range"))
	o.ProcessName = FastjsonUnmarshalListableString(fj.Get("process_name"))
	o.ProcessPath = FastjsonUnmarshalListableString(fj.Get("process_path"))
	o.PackageName = FastjsonUnmarshalListableString(fj.Get("package_name"))
	o.User = FastjsonUnmarshalListableString(fj.Get("user"))
	o.UserID = FastjsonUnmarshalListableInt32(fj.Get("user_id"))
	o.ClashMode, _ = FastjsonUnmarshalString(fj, "clash_mode")
	o.WIFISSID = FastjsonUnmarshalListableString(fj.Get("wifi_ssid"))
	o.WIFIBSSID = FastjsonUnmarshalListableString(fj.Get("wifi_bssid"))
	o.RuleSet = FastjsonUnmarshalListableString(fj.Get("rule_set"))
	o.RuleSetIPCIDRMatchSource, _ = FastjsonUnmarshalBool(fj, "rule_set_ipcidr_match_source")
	o.Invert , _= FastjsonUnmarshalBool(fj, "invert")
	o.Outbound, _ = FastjsonUnmarshalString(fj, "outbound")
	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	return nil
}

func (o *LogicalRule) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Mode, _ = FastjsonUnmarshalString(fj, "mode")
	o.Rules = FastjsonUnmarshalArrayRule(fj.Get("rules"))
	o.Invert, _ = FastjsonUnmarshalBool(fj, "invert")
	o.Outbound = FastjsonUnmarshalString(fj, "outbound")
	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	return nil
}

// shadowsocks.go
func (o *ShadowsocksInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	
	o.ListenOptions.FastjsonUnmarshal(fj)
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	o.Method, _ = FastjsonUnmarshalString(fj, "method")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	o.Users = FastjsonUnmarshalArrayShadowsocksUser(fj.Get("users"))
	o.Destinations = FastjsonUnmarshalArrayShadowsocksDestination(fj.Get("destinations"))
	return nil
}

func (o *ShadowsocksUser) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	return nil
}

func (o *ShadowsocksDestination) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	o.ServerOptions.FastjsonUnmarshal(fj)
	return nil
}

func (o *ShadowsocksOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.Method, _ = FastjsonUnmarshalString(fj, "method")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	o.Plugin, _ = FastjsonUnmarshalString(fj, "plugin")
	o.PluginOptions, _ = FastjsonUnmarshalString(fj, "plugin_opts")
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	udp_over_tcp := fj.Get("udp_over_tcp")
	if udp_over_tcp != nil && udp_over_tcp.Type() != fastjson.TypeNull {
		o.UDPOverTCP = &UDPOverTCPOptions{}
		o.UDPOverTCP.FastjsonUnmarshal(udp_over_tcp)
	}
	multiplex := fj.Get("multiplex")
	if multiplex != nil && multiplex.Type() != fastjson.TypeNull {
		o.Multiplex = &OutboundMultiplexOptions{}
		o.Multiplex.FastjsonUnmarshal(multiplex)
	}
	return nil
}

func FastjsonUnmarshalShadowsocksUser(fj *fastjson.Value, name string) (ShadowsocksUser, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return ShadowsocksUser{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return ShadowsocksUser{}, nil
	}
	vv := ShadowsocksUser{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayShadowsocksUser(fj *fastjson.Value) []ShadowsocksUser {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalShadowsocksUser)
}

func FastjsonUnmarshalShadowsocksDestination(fj *fastjson.Value, name string) (ShadowsocksDestination, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return ShadowsocksDestination{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return ShadowsocksDestination{}, nil
	}
	vv := ShadowsocksDestination{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayShadowsocksDestination(fj *fastjson.Value) []ShadowsocksDestination {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalShadowsocksDestination)
}

// shadowsocksr.go
func (o *ShadowsocksROutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.Method, _ = FastjsonUnmarshalString(fj, "method")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	o.Obfs, _ = FastjsonUnmarshalString(fj, "obfs")
	o.ObfsParam, _ = FastjsonUnmarshalString(fj, "obfs_param")
	o.Protocol, _ = FastjsonUnmarshalString(fj, "protocol")
	o.ProtocolParam, _ = FastjsonUnmarshalString(fj, "protocol_param")
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	return nil
}

// shadowtls.go
func (o *ShadowTLSInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.ListenOptions.FastjsonUnmarshal(fj)
	o.Version, _ = FastjsonUnmarshalInt(fj, "version")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	o.Users = FastjsonUnmarshalArrayShadowTLSUser(fj.Get("users"))
	o.Handshake.FastjsonUnmarshal(fj.Get("handshake"))
	o.HandshakeForServerName = FastjsonUnmarshalMapShadowTLSHandshakeOptions(fj.GetArray("handshake_for_server_name"))
	o.StrictMode, _ = FastjsonUnmarshalBool(fj, "strict_mode")
	return nil
}

func (o *ShadowTLSUser) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	return nil
}

func (o *ShadowTLSHandshakeOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.ServerOptions.FastjsonUnmarshal(fj)
	o.DialerOptions.FastjsonUnmarshal(fj)
	return nil
}

func (o *ShadowTLSOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.Version, _ = FastjsonUnmarshalInt(fj, "version")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	return nil
}

func FastjsonUnmarshalShadowTLSUser(fj *fastjson.Value, name string) (ShadowTLSUser, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return ShadowTLSUser{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return ShadowTLSUser{}, nil
	}
	vv := ShadowTLSUser{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayShadowTLSUser(fj *fastjson.Value) []ShadowTLSUser {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalShadowTLSUser)
}

func FastjsonUnmarshalMapShadowTLSHandshakeOptions(fj []*fastjson.Value) map[string]ShadowTLSHandshakeOptions {
	if fj == nil {
		return nil
	}

	list := make(map[string]ShadowTLSHandshakeOptions, len(fj))
	/*for i, v := range fj {
		vv := ShadowTLSHandshakeOptions{}
		vv.FastjsonUnmarshal(v)
		list[i] = vv
	}*/
	return list
}

// simple.go
func (o *SocksInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.ListenOptions.FastjsonUnmarshal(fj)
	o.Users = FastjsonUnmarshalArrayAuthUser(fj.Get("users"))
	return nil
}

func (o *HTTPMixedInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.ListenOptions.FastjsonUnmarshal(fj)
	o.Users = FastjsonUnmarshalArrayAuthUser(fj.Get("users"))
	o.SetSystemProxy, _ = FastjsonUnmarshalBool(fj, "set_system_proxy")
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	return nil
}

func (o *SOCKSOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.Version, _ = FastjsonUnmarshalString(fj, "version")
	o.Username, _ = FastjsonUnmarshalString(fj, "username")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	return nil
}

func (o *HTTPOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ServerOptions.FastjsonUnmarshal(fj)

	o.Username, _ = FastjsonUnmarshalString(fj, "username")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}

	o.Path, _ = FastjsonUnmarshalString(fj, "path")
	o.Headers = FastjsonUnmarshalHTTPHeader(fj, "headers")
	return nil
}

// ssh.go
func (o *SSHOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.User, _ = FastjsonUnmarshalString(fj, "user")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	o.PrivateKey = FastjsonUnmarshalListableString(fj.Get("private_key"))
	o.PrivateKeyPath, _ = FastjsonUnmarshalString(fj, "private_key_path")
	o.PrivateKeyPassphrase, _ = FastjsonUnmarshalString(fj, "private_key_passphrase")
	o.HostKey = FastjsonUnmarshalListableString(fj.Get("host_key"))
	o.HostKeyAlgorithms = FastjsonUnmarshalListableString(fj.Get("host_key_algorithms"))
	o.ClientVersion, _ = FastjsonUnmarshalString(fj, "client_version")
	udp_over_tcp := fj.Get("udp_over_tcp")
	if udp_over_tcp != nil && udp_over_tcp.Type() != fastjson.TypeNull {
		o.UDPOverTCP = &UDPOverTCPOptions{}
		o.UDPOverTCP.FastjsonUnmarshal(udp_over_tcp)
	}
	return nil
}

// tls_acme.go
func (o *InboundACMEOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Domain = FastjsonUnmarshalListableString(fj.Get("domain"))
	o.DataDirectory, _ = FastjsonUnmarshalString(fj, "data_directory")
	o.DefaultServerName, _ = FastjsonUnmarshalString(fj, "default_server_name")
	o.Email, _ = FastjsonUnmarshalString(fj, "email")
	o.Provider, _ = FastjsonUnmarshalString(fj, "provider")
	o.DisableHTTPChallenge, _ = FastjsonUnmarshalBool(fj, "disable_http_challenge")
	o.DisableTLSALPNChallenge, _ = FastjsonUnmarshalBool(fj, "disable_tls_alpn_challenge")
	o.AlternativeHTTPPort, _ = FastjsonUnmarshalUint16(fj, "alternative_http_port")
	o.AlternativeTLSPort, _ = FastjsonUnmarshalUint16(fj, "alternative_tls_port")
	external_account := fj.Get("external_account")
	dns01_challenge := fj.Get("dns01_challenge")
	if external_account != nil && external_account.Type() != fastjson.TypeNull {
		o.ExternalAccount = &ACMEExternalAccountOptions{}
		o.ExternalAccount.FastjsonUnmarshal(external_account)
	}
	if dns01_challenge != nil && dns01_challenge.Type() != fastjson.TypeNull {
		o.DNS01Challenge = &ACMEDNS01ChallengeOptions{}
		o.DNS01Challenge.FastjsonUnmarshal(dns01_challenge)
	}
	return nil
}

func (o *ACMEExternalAccountOptions) FastjsonUnmarshal(fj *fastjson.Value)error {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.KeyID, _ = FastjsonUnmarshalString(fj, "key_id")
	o.MACKey, _ = FastjsonUnmarshalString(fj, "mac_key")
	return nil
}

func (o *ACMEDNS01ChallengeOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Provider, _ = FastjsonUnmarshalString(fj, "provider")
	switch o.Provider {
	case C.DNSProviderAliDNS:
		o.AliDNSOptions.FastjsonUnmarshal(fj)
	case C.DNSProviderCloudflare:
		o.CloudflareOptions.FastjsonUnmarshal(fj)
	default:
		E.New("unknown provider type: " + o.Provider)
	}
	return nil
}

func (o *ACMEDNS01AliDNSOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.AccessKeyID, _ = FastjsonUnmarshalString(fj, "access_key_id")
	o.AccessKeySecret, _ = FastjsonUnmarshalString(fj, "access_key_secret")
	o.RegionID, _ = FastjsonUnmarshalString(fj, "region_id")
	return nil
}

func (o *ACMEDNS01CloudflareOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.APIToken, _ = FastjsonUnmarshalString(fj, "api_token")
	return nil
}

// tls.go
func (o *InboundTLSOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.ServerName, _ = FastjsonUnmarshalString(fj, "server_name")
	o.Insecure, _ = FastjsonUnmarshalBool(fj, "insecure")
	o.ALPN = FastjsonUnmarshalListableString(fj.Get("alpn"))
	o.MinVersion, _ = FastjsonUnmarshalString(fj, "min_version")
	o.MaxVersion, _ = FastjsonUnmarshalString(fj, "max_version")
	o.CipherSuites = FastjsonUnmarshalListableString(fj.Get("cipher_suites"))
	o.Certificate = FastjsonUnmarshalListableString(fj.Get("certificate"))
	o.CertificatePath, _ = FastjsonUnmarshalString(fj, "certificate_path")
	o.Key = FastjsonUnmarshalListableString(fj.Get("key"))
	o.KeyPath, _ = FastjsonUnmarshalString(fj, "key_path")
	acme := fj.Get("acme")
	if acme != nil && acme.Type() != fastjson.TypeNull {
		o.ACME = &InboundACMEOptions{}
		o.ACME.FastjsonUnmarshal(acme)
	}
	ech := fj.Get("ech")
	if ech != nil && ech.Type() != fastjson.TypeNull {
		o.ECH = &InboundECHOptions{}
		o.ECH.FastjsonUnmarshal(ech)
	}
	reality := fj.Get("reality")
	if reality != nil && reality.Type() != fastjson.TypeNull {
		o.Reality = &InboundRealityOptions{}
		o.Reality.FastjsonUnmarshal(reality)
	}
	return nil
}

func (o *OutboundTLSOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.DisableSNI, _ = FastjsonUnmarshalBool(fj, "disable_sni")
	o.ServerName, _ = FastjsonUnmarshalString(fj, "server_name")
	o.Insecure, _ = FastjsonUnmarshalBool(fj, "insecure")
	o.ALPN = FastjsonUnmarshalListableString(fj.Get("alpn"))
	o.MinVersion, _ = FastjsonUnmarshalString(fj, "min_version")
	o.MaxVersion, _ = FastjsonUnmarshalString(fj, "max_version")
	o.CipherSuites = FastjsonUnmarshalListableString(fj.Get("cipher_suites"))
	o.Certificate = FastjsonUnmarshalListableString(fj.Get("certificate"))
	o.CertificatePath, _ = FastjsonUnmarshalString(fj, "certificate_path")
	ech := fj.Get("ech")
	if ech != nil && ech.Type() != fastjson.TypeNull {
		o.ECH = &OutboundECHOptions{}
		o.ECH.FastjsonUnmarshal(ech)
	}
	utls := fj.Get("utls")
	if utls != nil && utls.Type() != fastjson.TypeNull {
		o.UTLS = &OutboundUTLSOptions{}
		o.UTLS.FastjsonUnmarshal(utls)
	}
	reality := fj.Get("reality")
	if reality != nil && reality.Type() != fastjson.TypeNull {
		o.Reality = &OutboundRealityOptions{}
		o.Reality.FastjsonUnmarshal(reality)
	}
	tls_tricks := fj.Get("tls_tricks")
	if tls_tricks != nil && tls_tricks.Type() != fastjson.TypeNull {
		o.TLSTricks = &TLSTricksOptions{}
		o.TLSTricks.FastjsonUnmarshal(tls_tricks)
	}
	return nil
}

func (o *InboundRealityOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.Handshake.FastjsonUnmarshal(fj.Get("handshake"))
	o.PrivateKey, _ = FastjsonUnmarshalString(fj, "private_key")
	o.ShortID = FastjsonUnmarshalListableString(fj.Get("short_id"))
	o.MaxTimeDifference, _ = FastjsonUnmarshalBadoptionDuration(fj, "max_time_difference")
	return nil
}

func (o *InboundRealityHandshakeOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.DialerOptions.FastjsonUnmarshal(fj)
	return nil
}

func (o *InboundECHOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.PQSignatureSchemesEnabled, _ = FastjsonUnmarshalBool(fj, "pq_signature_schemes_enabled")
	o.DynamicRecordSizingDisabled, _ = FastjsonUnmarshalBool(fj, "dynamic_record_sizing_disabled")
	o.Key = FastjsonUnmarshalListableString(fj.Get("key"))
	o.KeyPath, _ = FastjsonUnmarshalString(fj, "key_path")
	return nil
}

func (o *OutboundECHOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.PQSignatureSchemesEnabled, _ = FastjsonUnmarshalBool(fj, "pq_signature_schemes_enabled")
	o.DynamicRecordSizingDisabled, _ = FastjsonUnmarshalBool(fj, "dynamic_record_sizing_disabled")
	o.Config = FastjsonUnmarshalListableString(fj.Get("config"))
	o.ConfigPath, _ = FastjsonUnmarshalString(fj, "config_path")
	return nil
}

func (o *OutboundUTLSOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.Fingerprint, _ = FastjsonUnmarshalString(fj, "fingerprint")
	return nil
}

func (o *OutboundRealityOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.PublicKey, _ = FastjsonUnmarshalString(fj, "public_key")
	o.ShortID, _ = FastjsonUnmarshalString(fj, "short_id")
	return nil
}

// tor.go
func (o *TorOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ExecutablePath, _ = FastjsonUnmarshalString(fj, "executable_path")
	o.ExtraArgs  = FastjsonUnmarshalArrayString(fj, "extra_args")
	o.DataDirectory, _ = FastjsonUnmarshalString(fj, "data_directory")
	o.Options = FastjsonUnmarshalMapString(fj.GetArray("torrc"))
	return nil
}

func FastjsonUnmarshalMapString(fj []*fastjson.Value) map[string]string {
	if fj == nil {
		return nil
	}

	list := make(map[string]string, len(fj))
	/*for i, v := range fj {
		by, err := v.StringBytes()
		if err == nil {
			list[i] = FastjsonStringNotNil(by)
		}
	}*/
	return list
}

// trojan.go
func (o *TrojanInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.ListenOptions.FastjsonUnmarshal(fj)
	o.Users = FastjsonUnmarshalArrayTrojanUser(fj.Get("users"))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	fallback := fj.Get("fallback")
	if fallback != nil && fallback.Type() != fastjson.TypeNull {
		o.Fallback = &ServerOptions{}
		o.Fallback.FastjsonUnmarshal(fallback)
	}

	o.FallbackForALPN = FastjsonUnmarshalMapServerOptions(fj.GetArray("fallback_for_alpn"))
	transport := fj.Get("transport")
	if transport != nil && transport.Type() != fastjson.TypeNull {
		o.Transport = &V2RayTransportOptions{}
		o.Transport.FastjsonUnmarshal(transport)
	}
	return nil
}

func (o *TrojanUser) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	return nil
}

func (o *TrojanOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	multiplex := fj.Get("multiplex")
	if multiplex != nil && multiplex.Type() != fastjson.TypeNull {
		o.Multiplex = &OutboundMultiplexOptions{}
		o.Multiplex.FastjsonUnmarshal(multiplex)
	}
	transport := fj.Get("transport")
	if transport != nil && transport.Type() != fastjson.TypeNull {
		o.Transport = &V2RayTransportOptions{}
		o.Transport.FastjsonUnmarshal(transport)
	}
	return nil
}

func FastjsonUnmarshalTrojanUser(fj *fastjson.Value, name string) (TrojanUser, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return TrojanUser{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return TrojanUser{}, nil
	}
	vv := TrojanUser{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayTrojanUser(fj *fastjson.Value) []TrojanUser {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalTrojanUser)
}

func FastjsonUnmarshalMapServerOptions(fj []*fastjson.Value) map[string]*ServerOptions {
	if fj == nil {
		return nil
	}

	list := make(map[string]*ServerOptions, len(fj))
	/*for i, v := range fj {
		by, err := v.StringBytes()
		if err == nil {
			list[i] = FastjsonStringNotNil(by)
		}
	}*/
	return list
}

// tuic.go
func (o *TUICInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.ListenOptions.FastjsonUnmarshal(fj)
	o.Users = FastjsonUnmarshalArrayTUICUser(fj.Get("users"))
	o.CongestionControl, _ = FastjsonUnmarshalString(fj, "congestion_control")
	o.AuthTimeout, _ = FastjsonUnmarshalBadoptionDuration(fj, "auth_timeout")
	o.ZeroRTTHandshake, _ = FastjsonUnmarshalBool(fj, "zero_rtt_handshake")
	o.Heartbeat, _ = FastjsonUnmarshalBadoptionDuration(fj, "heartbeat")
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	return nil
}

func (o *TUICUser) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	o.UUID, _ = FastjsonUnmarshalString(fj, "uuid")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	return nil
}

func (o *TUICOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.UUID, _ = FastjsonUnmarshalString(fj, "uuid")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	o.CongestionControl, _ = FastjsonUnmarshalString(fj, "congestion_control")
	o.UDPRelayMode, _ = FastjsonUnmarshalString(fj, "udp_relay_mode")
	o.UDPOverStream, _ = FastjsonUnmarshalBool(fj, "udp_over_stream")
	o.ZeroRTTHandshake, _ = FastjsonUnmarshalBool(fj, "zero_rtt_handshake")
	o.Heartbeat, _ = FastjsonUnmarshalBadoptionDuration(fj, "heartbeat")
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	turn_relay := fj.Get("turn_relay")
	if turn_relay != nil && turn_relay.Type() != fastjson.TypeNull {
		o.TurnRelay = &TurnRelayOptions{}
		o.TurnRelay.FastjsonUnmarshal(turn_relay)
	}
	return nil
}

func FastjsonUnmarshalTUICUser(fj *fastjson.Value, name string) (TUICUser, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return TUICUser{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return TUICUser{}, nil
	}
	vv := TUICUser{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayTUICUser(fj *fastjson.Value) []TUICUser {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalTUICUser)
}

// tun_platform.go
func (o *TunPlatformOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	http_proxy := fj.Get("http_proxy")
	if http_proxy != nil && http_proxy.Type() != fastjson.TypeNull {
		o.HTTPProxy = &HTTPProxyOptions{}
		o.HTTPProxy.FastjsonUnmarshal(http_proxy)
	}
	o.AllowBypass, _ = FastjsonUnmarshalBool(fj, "allow_bypass")
	return nil
}

func (o *HTTPProxyOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.ServerOptions.FastjsonUnmarshal(fj)
	return nil
}

// tun.go
func (o *TunInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.InterfaceName, _ = FastjsonUnmarshalString(fj, "interface_name")
	o.MTU, _ = FastjsonUnmarshalUint32(fj, "mtu") 
	o.GSO, _ = FastjsonUnmarshalBool(fj, "gso")
	o.Inet4Address = FastjsonUnmarshalListableNetipPrefix(fj.Get("inet4_address"))
	o.Inet6Address = FastjsonUnmarshalListableNetipPrefix(fj.Get("inet6_address"))
	o.AutoRoute, _ = FastjsonUnmarshalBool(fj, "auto_route")
	o.StrictRoute, _ = FastjsonUnmarshalBool(fj, "strict_route")
	o.Inet4RouteAddress = FastjsonUnmarshalListableNetipPrefix(fj.Get("inet4_route_address"))
	o.Inet6RouteAddress = FastjsonUnmarshalListableNetipPrefix(fj.Get("inet6_route_address"))
	o.Inet4RouteExcludeAddress = FastjsonUnmarshalListableNetipPrefix(fj.Get("inet4_route_exclude_address"))
	o.Inet6RouteExcludeAddress = FastjsonUnmarshalListableNetipPrefix(fj.Get("inet6_route_exclude_address"))
	o.IncludeInterface = FastjsonUnmarshalListableString(fj.Get("include_interface"))
	o.ExcludeInterface = FastjsonUnmarshalListableString(fj.Get("exclude_interface"))
	o.IncludeUID = FastjsonUnmarshalListableUInt32(fj.Get("include_uid"))
	o.IncludeUIDRange = FastjsonUnmarshalListableString(fj.Get("include_uid_range"))
	o.ExcludeUID = FastjsonUnmarshalListableUInt32(fj.Get("exclude_uid"))
	o.ExcludeUIDRange = FastjsonUnmarshalListableString(fj.Get("exclude_uid_range"))
	o.IncludeAndroidUser = FastjsonUnmarshalListableInt(fj.Get("include_android_user"))
	o.IncludePackage = FastjsonUnmarshalListableString(fj.Get("include_package"))
	o.ExcludePackage = FastjsonUnmarshalListableString(fj.Get("exclude_package"))
	o.EndpointIndependentNat, _ = FastjsonUnmarshalBool(fj, "endpoint_independent_nat")
	udpimeout, _ := FastjsonUnmarshalBadoptionDuration(fj, "udp_timeout")
	o.UDPTimeout = UDPTimeoutCompat(udpimeout)
	o.Stack, _ = FastjsonUnmarshalString(fj, "stack")
	platform := fj.Get("platform")
	if platform != nil && platform.Type() != fastjson.TypeNull {
		o.Platform = &TunPlatformOptions{}
		o.Platform.FastjsonUnmarshal(platform)
	}

	o.InboundOptions.FastjsonUnmarshal(fj)
	return nil
}

// udp_over_tcp.go
func (o *UDPOverTCPOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.Version = uint8(fj.GetUint("version"))
	return nil
}

// v2ray_transport.go
func (o *V2RayTransportOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Type, _ = FastjsonUnmarshalString(fj, "type")
	switch o.Type {
	case C.V2RayTransportTypeHTTP:
		o.HTTPOptions.FastjsonUnmarshal(fj)
	case C.V2RayTransportTypeWebsocket:
		o.WebsocketOptions.FastjsonUnmarshal(fj)
	case C.V2RayTransportTypeQUIC:
		o.QUICOptions.FastjsonUnmarshal(fj)
	case C.V2RayTransportTypeGRPC:
		o.GRPCOptions.FastjsonUnmarshal(fj)
	case C.V2RayTransportTypeHTTPUpgrade:
		o.HTTPUpgradeOptions.FastjsonUnmarshal(fj)
	default:
		E.New("unknown transport type: " + o.Type)
	}
	return nil
}

func (o *V2RayHTTPOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Host = FastjsonUnmarshalListableString(fj.Get("host"))
	o.Path, _ = FastjsonUnmarshalString(fj, "path")
	o.Method, _ = FastjsonUnmarshalString(fj, "method")
	o.Headers = FastjsonUnmarshalHTTPHeader(fj, "headers")
	o.IdleTimeout, _ = FastjsonUnmarshalBadoptionDuration(fj, "idle_timeout")
	o.PingTimeout, _ = FastjsonUnmarshalBadoptionDuration(fj, "ping_timeout")
	return nil
}

func (o *V2RayWebsocketOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Path, _ = FastjsonUnmarshalString(fj, "path")
	o.Headers = FastjsonUnmarshalHTTPHeader(fj, "headers")
	o.MaxEarlyData = uint32(fj.GetUint("max_early_data"))
	o.EarlyDataHeaderName, _ = FastjsonUnmarshalString(fj, "early_data_header_name")
	return nil
}

func (o *V2RayQUICOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	return nil
}

func (o *V2RayGRPCOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.ServiceName, _ = FastjsonUnmarshalString(fj, "service_name")
	o.IdleTimeout, _ = FastjsonUnmarshalBadoptionDuration(fj, "idle_timeout")
	o.PingTimeout, _ = FastjsonUnmarshalBadoptionDuration(fj, "ping_timeout")
	o.PermitWithoutStream, _= FastjsonUnmarshalBool(fj, "permit_without_stream")
	//o.ForceLite        = FastjsonUnmarshalBool(fj, ""-")
	return nil
}

func (o *V2RayHTTPUpgradeOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Host, _ = FastjsonUnmarshalString(fj, "host")
	o.Path , _= FastjsonUnmarshalString(fj, "path")
	o.Headers = FastjsonUnmarshalHTTPHeader(fj, "headers")
	return nil
}

// v2ray.go
func FastjsonUnmarshalVLESSUser(fj *fastjson.Value, name string) (VLESSUser, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return VLESSUser{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return VLESSUser{}, nil
	}
	vv := VLESSUser{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayVLESSUser(fj *fastjson.Value) []VLESSUser {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalVLESSUser)
}

func (o *V2RayAPIOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Listen, _ = FastjsonUnmarshalString(fj, "listen")
	stats := fj.Get("stats")
	if stats != nil && stats.Type() != fastjson.TypeNull {
		o.Stats = &V2RayStatsServiceOptions{}
		o.Stats.FastjsonUnmarshal(stats)
	}
	return nil
}

func (o *CacheFileOptions) FastjsonUnmarshal(fj *fastjson.Value) error {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.Path, _ = FastjsonUnmarshalString(fj, "path")
	o.CacheID, _ = FastjsonUnmarshalString(fj, "cache_id")
	o.StoreFakeIP, _ = FastjsonUnmarshalBool(fj, "store_fakeip")
	return nil
}

func (o *V2RayStatsServiceOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.Inbounds = FastjsonUnmarshalArrayString(fj, "inbounds")
	o.Outbounds = FastjsonUnmarshalArrayString(fj, "outbounds")
	o.Users = FastjsonUnmarshalArrayString(fj, "users")
	return nil
}

// vless.go
func (o *VLESSInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.ListenOptions.FastjsonUnmarshal(fj)
	o.Users = FastjsonUnmarshalArrayVLESSUser(fj.Get("users"))
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &InboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	transport := fj.Get("transport")
	if transport != nil && transport.Type() != fastjson.TypeNull {
		o.Transport = &V2RayTransportOptions{}
		o.Transport.FastjsonUnmarshal(transport)
	}
	return nil
}

func (o *VLESSUser) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	o.UUID, _ = FastjsonUnmarshalString(fj, "uuid")
	o.Flow, _ = FastjsonUnmarshalString(fj, "flow")
	return nil
}

func (o *VLESSOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.UUID, _ = FastjsonUnmarshalString(fj, "uuid")
	o.Flow, _ = FastjsonUnmarshalString(fj, "flow")
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	multiplex := fj.Get("multiplex")
	if multiplex != nil && multiplex.Type() != fastjson.TypeNull {
		o.Multiplex = &OutboundMultiplexOptions{}
		o.Multiplex.FastjsonUnmarshal(multiplex)
	}
	transport := fj.Get("transport")
	if transport != nil && transport.Type() != fastjson.TypeNull {
		o.Transport = &V2RayTransportOptions{}
		o.Transport.FastjsonUnmarshal(transport)
	}
	o.PacketEncoding, _ =FastjsonUnmarshalStringPtr(fj, "packet_encoding")
	return nil
}

// vmess.go
func (o *VMessOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.UUID, _ = FastjsonUnmarshalString(fj, "uuid")
	o.Security, _ = FastjsonUnmarshalString(fj, "security")
	o.AlterId, _ = FastjsonUnmarshalInt(fj, "alter_id")
	o.GlobalPadding, _ = FastjsonUnmarshalBool(fj, "global_padding")
	o.AuthenticatedLength, _ = FastjsonUnmarshalBool(fj, "authenticated_length")
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	tls := fj.Get("tls")
	if tls != nil && tls.Type() != fastjson.TypeNull {
		o.TLS = &OutboundTLSOptions{}
		o.TLS.FastjsonUnmarshal(tls)
	}
	o.PacketEncoding, _ = FastjsonUnmarshalString(fj, "packet_encoding")

	multiplex := fj.Get("multiplex")
	if multiplex != nil && multiplex.Type() != fastjson.TypeNull {
		o.Multiplex = &OutboundMultiplexOptions{}
		o.Multiplex.FastjsonUnmarshal(multiplex)
	}
	transport := fj.Get("transport")
	if transport != nil && transport.Type() != fastjson.TypeNull {
		o.Transport = &V2RayTransportOptions{}
		o.Transport.FastjsonUnmarshal(transport)
	}
	return nil
}

// wireguard.go
func (o *LegacyWireGuardOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.DialerOptions.FastjsonUnmarshal(fj)
	o.SystemInterface, _ = FastjsonUnmarshalBool(fj, "system_interface")
	o.GSO, _ = FastjsonUnmarshalBool(fj, "gso")
	o.InterfaceName, _ = FastjsonUnmarshalString(fj, "interface_name")
	o.LocalAddress = FastjsonUnmarshalListableNetipPrefix(fj.Get("local_address"))
	o.PrivateKey, _ = FastjsonUnmarshalString(fj, "private_key")
	o.Peers = FastjsonUnmarshalArrayWireGuardPeer(fj.Get("peers"))
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.PeerPublicKey, _ = FastjsonUnmarshalString(fj, "peer_public_key")
	o.PreSharedKey, _ = FastjsonUnmarshalString(fj, "pre_shared_key")
	o.Reserved = FastjsonUnmarshalListableUint8(fj.Get("reserved"))
	o.Workers, _ = FastjsonUnmarshalInt(fj, "workers")
	o.MTU = uint32(fj.GetUint("mtu"))
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	turn_relay := fj.Get("turn_relay")
	if turn_relay != nil && turn_relay.Type() != fastjson.TypeNull {
		o.TurnRelay = &TurnRelayOptions{}
		o.TurnRelay.FastjsonUnmarshal(turn_relay)
	}
	o.FakePackets, _ = FastjsonUnmarshalString(fj, "fake_packets")
	o.FakePacketsSize, _ = FastjsonUnmarshalString(fj, "fake_packets_size")
	o.FakePacketsDelay, _ = FastjsonUnmarshalString(fj, "fake_packets_delay")
	o.FakePacketsMode, _ = FastjsonUnmarshalString(fj, "fake_packets_mode")
	return nil
}

func (o *LegacyWireGuardPeer) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.ServerOptions.FastjsonUnmarshal(fj)
	o.PublicKey, _ = FastjsonUnmarshalString(fj, "public_key")
	o.PreSharedKey, _ = FastjsonUnmarshalString(fj, "pre_shared_key")
	o.AllowedIPs = FastjsonUnmarshalListableNetipPrefix(fj.Get("allowed_ips"))
	o.Reserved = FastjsonUnmarshalListableUint8(fj.Get("reserved"))
	return nil
}

func FastjsonUnmarshalWireGuardPeer(fj *fastjson.Value, name string) (LegacyWireGuardPeer, error) {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return LegacyWireGuardPeer{}, nil
	}
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return LegacyWireGuardPeer{}, nil
	}
	vv := LegacyWireGuardPeer{}
	err := vv.FastjsonUnmarshal(value)
	return vv, err
}

func FastjsonUnmarshalArrayWireGuardPeer(fj *fastjson.Value) []LegacyWireGuardPeer {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalWireGuardPeer)
}

func FastjsonUnmarshalNetipPrefix(fj *fastjson.Value, name string) (netip.Prefix, error) {
	var value []byte
	if len(name) > 0 {
		value = fj.GetStringBytes(name)
	} else {
		value = fj.GetStringBytes()
	}
	if value == nil {
		return netip.Prefix{}, nil
	}
	vv, err := netip.ParsePrefix(string(value))
	if err != nil{
		return netip.Prefix{}, err
	}
	return vv, nil
}

func FastjsonUnmarshalListableNetipPrefix(fj *fastjson.Value) badoption.Listable[netip.Prefix] {
	return FastjsonUnmarshalArrayT(fj, "", fastjson.TypeObject, FastjsonUnmarshalNetipPrefix)
}

func (o *TLSFragmentOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.Size, _ = FastjsonUnmarshalString(fj, "size")
	o.Sleep, _ = FastjsonUnmarshalString(fj, "sleep")
	return nil
}

func (o *TurnRelayOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.Username, _ = FastjsonUnmarshalString(fj, "username")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	o.Realm, _ = FastjsonUnmarshalString(fj, "realm")
	return nil
}

func (o *TLSTricksOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	o.MixedCaseSNI, _ = FastjsonUnmarshalBool(fj, "mixedcase_sni")
	o.PaddingMode, _ = FastjsonUnmarshalString(fj, "padding_mode")
	o.PaddingSize, _ = FastjsonUnmarshalString(fj, "padding_size")
	o.PaddingSNI, _ = FastjsonUnmarshalString(fj, "padding_sni")
	return nil
}