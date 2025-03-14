//go:build !with_fastjson

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

type FastjsonUnmarshalInterface interface{
	FastjsonUnmarshal(fj *fastjson.Value) error 
}

type FastjsonUnmarshalStruct struct{
	FastjsonUnmarshalInterface 
}

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

func FastjsonUnmarshalFwMark(fj *fastjson.Value, name string) (FwMark, error) {
	value, err := FastjsonUnmarshalUint32(fj, name)
	return FwMark(value), err
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

func FastjsonUnmarshalDefaultInt()int {
	return 0
}

func FastjsonUnmarshalConvertInt(fj *fastjson.Value)(int, error){
	return fj.Int()
}

func FastjsonUnmarshalDefaultUint()uint{
	return 0
}

func FastjsonUnmarshalConvertUint(fj *fastjson.Value)(uint, error){
	return fj.Uint()
}

func FastjsonUnmarshalDefaultBool()bool{
	return false
}

func FastjsonUnmarshalConvertBool(fj *fastjson.Value)(bool, error){
	return fj.Bool()
}

func FastjsonUnmarshalDefaulString()string{
	return ""
}

func FastjsonUnmarshalConvertString(fj *fastjson.Value)(string, error){
	value, err := fj.StringBytes()
	if err != nil {
		return "", err
	}
	return string(value), nil
}

func FastjsonUnmarshalAny[T any](fnGet func(fj *fastjson.Value)(T, error), fnDefault func()T, fj *fastjson.Value, name string) (T, error) {
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return fnDefault(), FastjsonErrorNil
	}
	v, err := fnGet(value)
	if err != nil {
		return fnDefault(), err
	}
	return T(v), nil
}

func FastjsonUnmarshalAnyPtr[T any](fnGet func(fj *fastjson.Value)(T, error), fj *fastjson.Value, name string) (*T, error) {
	var value *fastjson.Value
	if len(name) > 0 {
		value = fj.Get(name)
	} else {
		value = fj.Get()
	}
	if value == nil {
		return nil, FastjsonErrorNil
	}
	v, err := fnGet(value)
	if err != nil {
		return nil, err
	}
	nvalue := new(T)
    *nvalue = T(v)
	return nvalue, nil
}

func FastjsonUnmarshal[T struct{}]( fnGet func(fj *fastjson.Value) (T, error), fj *fastjson.Value, name string) (T, error) {
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

	return fnGet(value)
}

//[T int|int32|string|struct{}]
func FastjsonUnmarshalArrayT[T any]( fnGet func(fj *fastjson.Value, name string) (T, error), fj *fastjson.Value, name string, fallbackType fastjson.Type) []T { 
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
			value, err := fnGet(v, "")
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
	value, err := fnGet(fj.Get(), "")
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
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalString, fj, name, fastjson.TypeString)
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

func FastjsonUnmarshalListableT[T any](fn func(fj *fastjson.Value, name string) (T, error), fj *fastjson.Value, name string, fallbackType fastjson.Type) badoption.Listable[T] {
	return badoption.Listable[T](FastjsonUnmarshalArrayT[T](fn, fj, name, fallbackType ))
}

func FastjsonUnmarshalListableString(fj *fastjson.Value, name string) badoption.Listable[string] {
	return FastjsonUnmarshalListableT(FastjsonUnmarshalString, fj, name, fastjson.TypeString)
}

func FastjsonUnmarshalListableInt(fj *fastjson.Value, name string) badoption.Listable[int] {
	return FastjsonUnmarshalListableT(FastjsonUnmarshalInt, fj, name, fastjson.TypeNumber)
}

func FastjsonUnmarshalListableInt32(fj *fastjson.Value, name string) badoption.Listable[int32] {
	return FastjsonUnmarshalListableT(FastjsonUnmarshalInt32, fj, name, fastjson.TypeNumber)
}

func FastjsonUnmarshalListableUInt32(fj *fastjson.Value, name string) badoption.Listable[uint32] {
	return FastjsonUnmarshalListableT(FastjsonUnmarshalUint32, fj, name, fastjson.TypeNumber)
}

func FastjsonUnmarshalListableUint16(fj *fastjson.Value, name string) badoption.Listable[uint16] {
	return FastjsonUnmarshalListableT(FastjsonUnmarshalUint16, fj, name, fastjson.TypeNumber)
}

func FastjsonUnmarshalListableUint8(fj *fastjson.Value, name string) badoption.Listable[uint8] {
	return FastjsonUnmarshalListableT(FastjsonUnmarshalUint8, fj, name, fastjson.TypeNumber)
}

func FastjsonUnmarshalDNSQueryType(fj *fastjson.Value, name string) (DNSQueryType, error) {
	if len(name) > 0 {
		return DNSQueryType(fj.GetInt(name)), nil
	}
	return DNSQueryType(fj.GetInt()), nil
}

func FastjsonUnmarshalListableDNSQueryType(fj *fastjson.Value, name string) badoption.Listable[DNSQueryType] {
	return FastjsonUnmarshalListableT(FastjsonUnmarshalDNSQueryType, fj,name, fastjson.TypeNumber )
}

func FastjsonUnmarshalHTTPHeader(fj *fastjson.Value, name string) (badoption.HTTPHeader, error) {
	value := fj.GetObject(name)
	if value == nil {
		return nil, nil
	}
	list := make(badoption.HTTPHeader, value.Len())
	value.Visit(func(key []byte, value *fastjson.Value) {
		list[string(key)] = FastjsonUnmarshalArrayString(value, "")
	})

	return list, nil
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
	//ModeList                         []string                   `json:"-"`
	o.AccessControlAllowOrigin = FastjsonUnmarshalListableString(fj, "access_control_allow_origin")
	o.AccessControlAllowPrivateNetwork, _  = FastjsonUnmarshalBool(fj, "access_control_allow_private_network") 
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
func (o *Options) UnmarshalFastjson(content []byte) error {
	var parser fastjson.Parser
	value, err := parser.ParseBytes(content)
	if err != nil {
		return err
	}

	o.FastjsonUnmarshal(value)
	return nil
}

func FastjsonUnmarshalDefaultLogOptions() (LogOptions) {
	return LogOptions{}
}

func FastjsonUnmarshalConvertLogOptions(fj *fastjson.Value) (LogOptions, error) {
	vv := LogOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalLogOptions(fj *fastjson.Value, name string) (*LogOptions, error) {
	return FastjsonUnmarshalAnyPtr[LogOptions](FastjsonUnmarshalConvertLogOptions, fj,name)
}

func FastjsonUnmarshalDefaultDNSOptions() (DNSOptions) {
	return DNSOptions{}
}

func FastjsonUnmarshalConvertDNSOptions(fj *fastjson.Value) (DNSOptions, error) {
	vv := DNSOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalDNSOptions(fj *fastjson.Value, name string) (*DNSOptions, error) {
	return FastjsonUnmarshalAnyPtr[DNSOptions](FastjsonUnmarshalConvertDNSOptions, fj,name)
}

func FastjsonUnmarshalDefaultNTPOptions() (NTPOptions) {
	return NTPOptions{}
}

func FastjsonUnmarshalConvertNTPOptions(fj *fastjson.Value) (NTPOptions, error) {
	vv := NTPOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalNTPOptions(fj *fastjson.Value, name string) (*NTPOptions, error) {
	return FastjsonUnmarshalAnyPtr[NTPOptions](FastjsonUnmarshalConvertNTPOptions, fj,name)
}

func FastjsonUnmarshalDefaultRouteOptions() (RouteOptions) {
	return RouteOptions{}
}

func FastjsonUnmarshalConvertRouteOptions(fj *fastjson.Value) (RouteOptions, error) {
	vv := RouteOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalRouteOptions(fj *fastjson.Value, name string) (*RouteOptions, error) {
	return FastjsonUnmarshalAnyPtr[RouteOptions](FastjsonUnmarshalConvertRouteOptions, fj,name)
}

func FastjsonUnmarshalDefaultExperimentalOptions() (ExperimentalOptions) {
	return ExperimentalOptions{}
}

func FastjsonUnmarshalConvertExperimentalOptions(fj *fastjson.Value) (ExperimentalOptions, error) {
	vv := ExperimentalOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalExperimentalOptions(fj *fastjson.Value, name string) (*ExperimentalOptions, error) {
	return FastjsonUnmarshalAnyPtr[ExperimentalOptions](FastjsonUnmarshalConvertExperimentalOptions, fj,name)
}

func (o *Options) FastjsonUnmarshal(fj *fastjson.Value) error{
	o.Schema, _ = FastjsonUnmarshalString(fj, "schema")
	o.Log , _ = FastjsonUnmarshalLogOptions(fj, "log")
	o.DNS , _ = FastjsonUnmarshalDNSOptions(fj, "dns")
	o.NTP , _ = FastjsonUnmarshalNTPOptions(fj, "ntp")
	o.Inbounds = FastjsonUnmarshalArrayInbound(fj,"inbounds")
	o.Outbounds = FastjsonUnmarshalArrayOutbound(fj, "outbounds")
	o.Route , _ = FastjsonUnmarshalRouteOptions(fj, "route")
	o.Experimental , _ = FastjsonUnmarshalExperimentalOptions(fj, "experimental")
 
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
	o.GCPercent, _ = FastjsonUnmarshalAnyPtr[int](FastjsonUnmarshalConvertInt, fj, "gc_percent")
	o.MaxStack, _ = FastjsonUnmarshalAnyPtr[int](FastjsonUnmarshalConvertInt,fj, "max_stack")
	o.MaxThreads, _ = FastjsonUnmarshalAnyPtr[int](FastjsonUnmarshalConvertInt,fj, "max_threads")
	o.PanicOnFault, _ = FastjsonUnmarshalAnyPtr[bool](FastjsonUnmarshalConvertBool, fj, "panic_on_fault")

	o.TraceBack, _ = FastjsonUnmarshalString(fj, "trace_back")
	o.MemoryLimit = MemoryBytes(fj.GetInt64("memory_limit"))
	o.OOMKiller, _ = FastjsonUnmarshalAnyPtr[bool](FastjsonUnmarshalConvertBool, fj, "oom_killer")
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
func FastjsonUnmarshalDefaultDNSFakeIPOptions() (DNSFakeIPOptions) {
	return DNSFakeIPOptions{}
}

func FastjsonUnmarshalConvertDNSFakeIPOptions(fj *fastjson.Value) (DNSFakeIPOptions, error) {
	vv := DNSFakeIPOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalDNSFakeIPOptions(fj *fastjson.Value, name string) (*DNSFakeIPOptions, error) {
	return FastjsonUnmarshalAnyPtr[DNSFakeIPOptions](FastjsonUnmarshalConvertDNSFakeIPOptions, fj,name)
}

func (o *DNSOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Servers = FastjsonUnmarshalArrayDNSServerOptions(fj,"servers")
	o.Rules = FastjsonUnmarshalArrayDNSRule(fj,"rules")
	o.FakeIP , _ = FastjsonUnmarshalDNSFakeIPOptions(fj, "fakeip")
	o.Final, _ = FastjsonUnmarshalString(fj, "final")
	o.ReverseMapping, _ = FastjsonUnmarshalBool(fj, "reverse_mapping")
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
	o.Inet4Range, _ = FastjsonUnmarshalAnyPtr[netip.Prefix](FastjsonUnmarshalConvertNetipPrefix, fj, "inet4_range")
	o.Inet6Range, _ = FastjsonUnmarshalAnyPtr[netip.Prefix](FastjsonUnmarshalConvertNetipPrefix, fj, "inet6_range")

	return nil
}

func FastjsonUnmarshalDefaultDNSServerOptions() DNSServerOptions {
	return DNSServerOptions{}
}

func FastjsonUnmarshalConvertDNSServerOptions(fj *fastjson.Value) (DNSServerOptions, error) {
	vv := DNSServerOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalDNSServerOptions(fj *fastjson.Value, name string) (DNSServerOptions, error) {
	return FastjsonUnmarshalAny[DNSServerOptions](FastjsonUnmarshalConvertDNSServerOptions, FastjsonUnmarshalDefaultDNSServerOptions, fj, name)
}
 
func FastjsonUnmarshalArrayDNSServerOptions(fj *fastjson.Value, name string) []DNSServerOptions {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalDNSServerOptions,fj, name, fastjson.TypeObject )
}

// experimental.go
func FastjsonUnmarshalDefaultClashAPIOptions() ClashAPIOptions {
	return ClashAPIOptions{}
}

func FastjsonUnmarshalConvertClashAPIOptions(fj *fastjson.Value) (ClashAPIOptions, error) {
	vv := ClashAPIOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshaClashAPIOptions(fj *fastjson.Value, name string) (*ClashAPIOptions, error) {
	return FastjsonUnmarshalAnyPtr[ClashAPIOptions](FastjsonUnmarshalConvertClashAPIOptions,  fj, name)
}

func FastjsonUnmarshalDefaultV2RayAPIOptions() V2RayAPIOptions {
	return V2RayAPIOptions{}
}

func FastjsonUnmarshalConvertV2RayAPIOptions(fj *fastjson.Value) (V2RayAPIOptions, error) {
	vv := V2RayAPIOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshaV2RayAPIOptions(fj *fastjson.Value, name string) (*V2RayAPIOptions, error) {
	return FastjsonUnmarshalAnyPtr[V2RayAPIOptions](FastjsonUnmarshalConvertV2RayAPIOptions,  fj, name)
}

func FastjsonUnmarshalDefaultCacheFileOptions() CacheFileOptions {
	return CacheFileOptions{}
}

func FastjsonUnmarshalConvertCacheFileOptions(fj *fastjson.Value) (CacheFileOptions, error) {
	vv := CacheFileOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshaCacheFileOptions(fj *fastjson.Value, name string) (*CacheFileOptions, error) {
	return FastjsonUnmarshalAnyPtr[CacheFileOptions](FastjsonUnmarshalConvertCacheFileOptions,  fj, name)
}

func FastjsonUnmarshalDefaultDebugOptions() DebugOptions {
	return DebugOptions{}
}

func FastjsonUnmarshalConvertDebugOptions(fj *fastjson.Value) (DebugOptions, error) {
	vv := DebugOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshaDebugOptions(fj *fastjson.Value, name string) (*DebugOptions, error) {
	return FastjsonUnmarshalAnyPtr[DebugOptions](FastjsonUnmarshalConvertDebugOptions,  fj, name)
}


func (o *ExperimentalOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.ClashAPI ,_=FastjsonUnmarshaClashAPIOptions(fj, "clash_api")
	o.V2RayAPI ,_=FastjsonUnmarshaV2RayAPIOptions(fj, "v2ray_api")
	o.CacheFile ,_=FastjsonUnmarshaCacheFileOptions(fj, "cache_file")
	o.Debug ,_=FastjsonUnmarshaDebugOptions(fj, "debug")
 
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
	o.Users = FastjsonUnmarshalArrayHysteriaUser(fj,"users")
	o.ReceiveWindowConn = fj.GetUint64("recv_window_conn")
	o.ReceiveWindowClient = fj.GetUint64("recv_window_client")
	o.MaxConnClient, _ = FastjsonUnmarshalInt(fj, "max_conn_client")
	o.DisableMTUDiscovery, _ = FastjsonUnmarshalBool(fj, "disable_mtu_discovery")
	o.TLS , _ =FastjsonUnmarshalInboundTLSOptions(fj, "tls")
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
	o.TLS, _ =FastjsonUnmarshalOutboundTLSOptions(fj, "tls")
	o.TurnRelay, _= FastjsonUnmarshalTurnRelayOptions(fj, "turn_relay")
	o.HopPorts, _ = FastjsonUnmarshalString(fj, "hop_ports")
	o.HopInterval, _ = FastjsonUnmarshalInt(fj, "hop_interval")
	return nil
}

func FastjsonUnmarshalDefaultHysteriaUser() HysteriaUser {
	return HysteriaUser{}
}

func FastjsonUnmarshalConvertHysteriaUser(fj *fastjson.Value) (HysteriaUser, error) {
	vv := HysteriaUser{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshaHysteriaUser(fj *fastjson.Value, name string) (HysteriaUser, error) {
	return FastjsonUnmarshalAny[HysteriaUser](FastjsonUnmarshalConvertHysteriaUser,FastjsonUnmarshalDefaultHysteriaUser, fj, name)
}

func FastjsonUnmarshalArrayHysteriaUser(fj *fastjson.Value, name string) []HysteriaUser {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshaHysteriaUser, fj, name, fastjson.TypeObject )
}

// hysteria2.go
func (o *Hysteria2MasqueradeFile) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Directory,_ = FastjsonUnmarshalString(fj,"directory")
	return nil
}

func (o *Hysteria2MasqueradeProxy) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.URL,_ = FastjsonUnmarshalString(fj,"url")
	o.RewriteHost,_ = FastjsonUnmarshalBool(fj,"rewrite_host")
	return nil
}

func (o *Hysteria2MasqueradeString) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.StatusCode,_ = FastjsonUnmarshalInt(fj,"status_code")
	o.Headers,_ = FastjsonUnmarshalHTTPHeader(fj, "headers")
	o.Content,_ = FastjsonUnmarshalString(fj,"content")
	return nil
}
 
func (o *Hysteria2Masquerade) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Type,_ = FastjsonUnmarshalString(fj,"type")
	switch o.Type {
	case C.Hysterai2MasqueradeTypeFile:
		o.FileOptions.FastjsonUnmarshal(fj)
	case C.Hysterai2MasqueradeTypeProxy:
		o.ProxyOptions.FastjsonUnmarshal(fj)
	case C.Hysterai2MasqueradeTypeString:
		o.StringOptions.FastjsonUnmarshal(fj)
	default:
		return  E.New("unknown masquerade type: ", o.Type)
	}
	return nil
}
func FastjsonUnmarshalDefaultHysteria2Masquerade() Hysteria2Masquerade {
	return Hysteria2Masquerade{}
}

func FastjsonUnmarshalConvertHysteria2Masquerade(fj *fastjson.Value) (Hysteria2Masquerade, error) {
	vv := Hysteria2Masquerade{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshaHysteria2Masquerade(fj *fastjson.Value, name string) (*Hysteria2Masquerade, error) {
	return FastjsonUnmarshalAnyPtr[Hysteria2Masquerade](FastjsonUnmarshalConvertHysteria2Masquerade, fj, name)
}

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

	o.Users = FastjsonUnmarshalArrayHysteria2User(fj,"users")
	o.IgnoreClientBandwidth, _ = FastjsonUnmarshalBool(fj, "ignore_client_bandwidth")
	o.TLS , _ =FastjsonUnmarshalInboundTLSOptions(fj, "tls")
	o.Masquerade, _ = FastjsonUnmarshaHysteria2Masquerade(fj, "masquerade")
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
	o.TLS, _ =FastjsonUnmarshalOutboundTLSOptions(fj, "tls")
	o.BrutalDebug, _ = FastjsonUnmarshalBool(fj, "brutal_debug")
	o.TurnRelay, _= FastjsonUnmarshalTurnRelayOptions(fj, "turn_relay")
	hopPorts, _ := FastjsonUnmarshalString(fj, "hop_ports")
	o.HopPorts  = HopPortsValue(hopPorts)
	hopInterval, _ := FastjsonUnmarshalInt(fj, "hop_interval")
	o.HopInterval = HopIntervalValue(hopInterval)
	return nil
}

func FastjsonUnmarshalDefaultHysteria2User() Hysteria2User {
	return Hysteria2User{}
}

func FastjsonUnmarshalConvertHysteria2User(fj *fastjson.Value) (Hysteria2User, error) {
	vv := Hysteria2User{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshaHysteria2User(fj *fastjson.Value, name string) (Hysteria2User, error) {
	return FastjsonUnmarshalAny[Hysteria2User](FastjsonUnmarshalConvertHysteria2User,FastjsonUnmarshalDefaultHysteria2User, fj, name)
}

func FastjsonUnmarshalArrayHysteria2User(fj *fastjson.Value, name string) []Hysteria2User {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshaHysteria2User, fj, name, fastjson.TypeObject )
}

// inbound.go
func (h *Inbound) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	h.Type, _ = FastjsonUnmarshalString(fj, "type")
	h.Tag, _ = FastjsonUnmarshalString(fj, "tag")
	/*switch h.Type {
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
	}*/
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
	o.Listen ,_= FastjsonUnmarshalBadoptionAddr(fj, "listen")
	o.ListenPort, _ = FastjsonUnmarshalUint16(fj, "listen_port")
	o.TCPFastOpen, _ = FastjsonUnmarshalBool(fj, "tcp_fast_open")
	o.TCPMultiPath, _ = FastjsonUnmarshalBool(fj, "tcp_multi_path")
	o.UDPFragment, _ =FastjsonUnmarshalAnyPtr[bool](FastjsonUnmarshalConvertBool, fj, "udp_fragment")
	//o.UDPFragmentDefault           = FastjsonUnmarshalBool(fj, "-")
	udpimeout, _ := FastjsonUnmarshalBadoptionDuration(fj, "udp_timeout")
	o.UDPTimeout = UDPTimeoutCompat(udpimeout)
	o.ProxyProtocol, _ = FastjsonUnmarshalBool(fj, "proxy_protocol")
	o.ProxyProtocolAcceptNoHeader, _ = FastjsonUnmarshalBool(fj, "proxy_protocol_accept_no_header")
	o.Detour, _ = FastjsonUnmarshalString(fj, "detour")
	o.InboundOptions.FastjsonUnmarshal(fj)
	return nil
}

func FastjsonUnmarshalDefaultInbound() Inbound {
	return Inbound{}
}

func FastjsonUnmarshalConvertInbound(fj *fastjson.Value) (Inbound, error) {
	vv := Inbound{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalInbound(fj *fastjson.Value, name string) (Inbound, error) {
	return FastjsonUnmarshalAny[Inbound](FastjsonUnmarshalConvertInbound,FastjsonUnmarshalDefaultInbound, fj, name)
}

func FastjsonUnmarshalArrayInbound(fj *fastjson.Value, name string) []Inbound {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalInbound, fj, name, fastjson.TypeObject )
}

// naive.go
func (o *NaiveInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.ListenOptions.FastjsonUnmarshal(fj)
	o.Users = FastjsonUnmarshalArrayAuthUser(fj,"users")
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	o.TLS , _ =FastjsonUnmarshalInboundTLSOptions(fj, "tls")
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
	/*switch h.Type {
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
	}*/
	return nil
}

func FastjsonUnmarshalDefaultBadoptionAddr() (badoption.Addr) {
	return badoption.Addr{}
}

func FastjsonUnmarshalConvertBadoptionAddr(fj *fastjson.Value) (badoption.Addr, error) {
	value, err := FastjsonUnmarshalString(fj, "inet4_bind_address")
	if err != nil {
		return badoption.Addr{}, err
	}
	addr, err4 := netip.ParseAddr(value)
	if err4 != nil {
		return badoption.Addr{}, err4
	}
	 
	return badoption.Addr(addr), nil
}

func FastjsonUnmarshalBadoptionAddr(fj *fastjson.Value, name string) (*badoption.Addr, error) {
	return FastjsonUnmarshalAnyPtr[badoption.Addr](FastjsonUnmarshalConvertBadoptionAddr, fj,name)
}

func (o *DialerOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Detour, _ = FastjsonUnmarshalString(fj, "detour")
	o.BindInterface, _ = FastjsonUnmarshalString(fj, "bind_interface")
	o.Inet4BindAddress ,_= FastjsonUnmarshalBadoptionAddr(fj, "inet4_bind_address")
	o.Inet6BindAddress ,_= FastjsonUnmarshalBadoptionAddr(fj, "inet6_bind_address")
	o.ProtectPath, _ = FastjsonUnmarshalString(fj, "protect_path")
	o.RoutingMark, _ = FastjsonUnmarshalFwMark(fj, "routing_mark")
	o.ReuseAddr, _ = FastjsonUnmarshalBool(fj, "reuse_addr")
	o.ConnectTimeout, _ = FastjsonUnmarshalBadoptionDuration(fj,"connect_timeout")
	o.TCPFastOpen, _ = FastjsonUnmarshalBool(fj, "tcp_fast_open")
	o.TCPMultiPath, _ = FastjsonUnmarshalBool(fj, "tcp_multi_path")
	o.UDPFragment, _ = FastjsonUnmarshalAnyPtr[bool](FastjsonUnmarshalConvertBool, fj, "udp_fragment")
	//o.UDPFragmentDefault = FastjsonUnmarshalBool(fj, "-")
	o.DomainStrategy, _ = FastjsonUnmarshalDomainStrategy(fj,"domain_strategy")
	o.FallbackDelay, _ = FastjsonUnmarshalBadoptionDuration(fj,"fallback_delay")
	o.TLSFragment,_ = FastjsonUnmarshalTLSFragmentOptions(fj, "tls_fragment")
	 
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

func FastjsonUnmarshalDefaultOutbound() Outbound {
	return Outbound{}
}

func FastjsonUnmarshalConvertOutbound(fj *fastjson.Value) (Outbound, error) {
	vv := Outbound{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalOutbound(fj *fastjson.Value, name string) (Outbound, error) {
	return FastjsonUnmarshalAny[Outbound](FastjsonUnmarshalConvertOutbound,FastjsonUnmarshalDefaultOutbound, fj, name)
}

func FastjsonUnmarshalArrayOutbound(fj *fastjson.Value, name string) []Outbound {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalOutbound, fj, name, fastjson.TypeObject )
}

func FastjsonUnmarshalAuthUser2(o *auth.User, fj *fastjson.Value) error {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Username, _ = FastjsonUnmarshalString(fj, "username")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	return nil
}

func FastjsonUnmarshalDefaultAuthUser() auth.User {
	return auth.User{}
}

func FastjsonUnmarshalConvertAuthUser(fj *fastjson.Value) (auth.User, error) {
	vv := auth.User{}
	err := FastjsonUnmarshalAuthUser2(&vv, fj)
	return vv, err
}

func FastjsonUnmarshalAuthUser(fj *fastjson.Value, name string) (auth.User, error) {
	return FastjsonUnmarshalAny[auth.User](FastjsonUnmarshalConvertAuthUser,FastjsonUnmarshalDefaultAuthUser, fj, name)
}

func FastjsonUnmarshalArrayAuthUser(fj *fastjson.Value, name string) []auth.User {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalAuthUser, fj, name, fastjson.TypeObject)
}

// platform.go
func (o *OnDemandOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.Rules = FastjsonUnmarshalArrayOnDemandRule(fj,"rules")
	return nil
}

func FastjsonUnmarshalConvertOnDemandRuleAction(fj *fastjson.Value)(OnDemandRuleAction, error){
	value, err := fj.Int()
	return OnDemandRuleAction(value), err
}

func FastjsonUnmarshalConvertOnDemandRuleInterfaceType(fj *fastjson.Value)(OnDemandRuleInterfaceType, error){
	return OnDemandRuleInterfaceType(fj.GetInt()), nil
}

func (o *OnDemandRule) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Action,_ = FastjsonUnmarshalAnyPtr[OnDemandRuleAction](FastjsonUnmarshalConvertOnDemandRuleAction, fj, "action") 
	o.DNSSearchDomainMatch = FastjsonUnmarshalListableString(fj, "dns_search_domain_match")
	o.DNSServerAddressMatch = FastjsonUnmarshalListableString(fj, "dns_server_address_match")
	o.InterfaceTypeMatch,_ =FastjsonUnmarshalAnyPtr[OnDemandRuleInterfaceType](FastjsonUnmarshalConvertOnDemandRuleInterfaceType, fj, "interface_type_match")
	o.SSIDMatch = FastjsonUnmarshalListableString(fj, "ssid_match")
	o.ProbeURL, _ = FastjsonUnmarshalString(fj, "probe_url")
	return nil
}

func FastjsonUnmarshalDefaultOnDemandRule() OnDemandRule {
	return OnDemandRule{}
}

func FastjsonUnmarshalConvertOnDemandRule(fj *fastjson.Value) (OnDemandRule, error) {
	vv := OnDemandRule{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalOnDemandRule(fj *fastjson.Value, name string) (OnDemandRule, error) {
	return FastjsonUnmarshalAny[OnDemandRule](FastjsonUnmarshalConvertOnDemandRule,FastjsonUnmarshalDefaultOnDemandRule, fj, name)
}

func FastjsonUnmarshalArrayOnDemandRule(fj *fastjson.Value, name string) []OnDemandRule {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalOnDemandRule, fj, name, fastjson.TypeObject)
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
func FastjsonUnmarshalDefaultGeoIPOptions() (GeoIPOptions) {
	return GeoIPOptions{}
}

func FastjsonUnmarshalConvertGeoIPOptions(fj *fastjson.Value) (GeoIPOptions, error) {
	vv := GeoIPOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalGeoIPOptions(fj *fastjson.Value, name string) (*GeoIPOptions, error) {
	return FastjsonUnmarshalAnyPtr[GeoIPOptions](FastjsonUnmarshalConvertGeoIPOptions, fj,name)
}

func FastjsonUnmarshalDefaultGeositeOptions() (GeositeOptions) {
	return GeositeOptions{}
}

func FastjsonUnmarshalConvertGeositeOptions(fj *fastjson.Value) (GeositeOptions, error) {
	vv := GeositeOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalGeositeOptions(fj *fastjson.Value, name string) (*GeositeOptions, error) {
	return FastjsonUnmarshalAnyPtr[GeositeOptions](FastjsonUnmarshalConvertGeositeOptions, fj,name)
}


func (o *RouteOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.GeoIP, _ = FastjsonUnmarshalGeoIPOptions(fj,"geoip")
	o.Geosite, _ = FastjsonUnmarshalGeositeOptions(fj,"geosite")
	o.Rules = FastjsonUnmarshalArrayRule(fj,"rules")
	o.RuleSet = FastjsonUnmarshalArrayRuleSet(fj,"rule_set")
	o.Final, _ = FastjsonUnmarshalString(fj, "final")
	o.FindProcess, _ = FastjsonUnmarshalBool(fj, "find_process")
	o.AutoDetectInterface, _ = FastjsonUnmarshalBool(fj, "auto_detect_interface")
	o.OverrideAndroidVPN, _ = FastjsonUnmarshalBool(fj, "override_android_vpn")
	o.DefaultInterface, _ = FastjsonUnmarshalString(fj, "default_interface")
	o.DefaultMark, _ = FastjsonUnmarshalFwMark(fj, "default_mark")
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
	o.Inbound = FastjsonUnmarshalListableString(fj, "inbound")
	o.IPVersion, _ = FastjsonUnmarshalInt(fj, "ip_version")
	o.QueryType = FastjsonUnmarshalListableDNSQueryType(fj, "query_type")
	o.Network = FastjsonUnmarshalListableString(fj, "network")
	o.AuthUser = FastjsonUnmarshalListableString(fj, "auth_user")
	o.Protocol = FastjsonUnmarshalListableString(fj, "protocol")
	o.Domain = FastjsonUnmarshalListableString(fj, "domain")
	o.DomainSuffix = FastjsonUnmarshalListableString(fj, "domain_suffix")
	o.DomainKeyword = FastjsonUnmarshalListableString(fj, "domain_keyword")
	o.DomainRegex = FastjsonUnmarshalListableString(fj, "domain_regex")
	o.Geosite = FastjsonUnmarshalListableString(fj, "geosite")
	o.SourceGeoIP = FastjsonUnmarshalListableString(fj, "source_geoip")
	//todo
	//o.GeoIP = FastjsonUnmarshalListableString(fj, "geoip"))
	//o.IPCIDR = FastjsonUnmarshalListableString(fj, "ip_cidr"))
	//o.IPIsPrivate = FastjsonUnmarshalBool(fj, "ip_is_private")
	o.SourceIPCIDR = FastjsonUnmarshalListableString(fj, "source_ip_cidr")
	o.SourceIPIsPrivate, _ = FastjsonUnmarshalBool(fj, "source_ip_is_private")
	o.SourcePort = FastjsonUnmarshalListableUint16(fj, "source_port")
	o.SourcePortRange = FastjsonUnmarshalListableString(fj, "source_port_range")
	o.Port = FastjsonUnmarshalListableUint16(fj, "port")
	o.PortRange = FastjsonUnmarshalListableString(fj, "port_range")
	o.ProcessName = FastjsonUnmarshalListableString(fj, "process_name")
	o.ProcessPath = FastjsonUnmarshalListableString(fj, "process_path")
	o.PackageName = FastjsonUnmarshalListableString(fj, "package_name")
	o.User = FastjsonUnmarshalListableString(fj, "user")
	o.UserID = FastjsonUnmarshalListableInt32(fj, "user_id")
	o.Outbound = FastjsonUnmarshalListableString(fj, "outbound")
	o.ClashMode, _ = FastjsonUnmarshalString(fj, "clash_mode")
	o.WIFISSID = FastjsonUnmarshalListableString(fj, "wifi_ssid")
	o.WIFIBSSID = FastjsonUnmarshalListableString(fj, "wifi_bssid")
	o.RuleSet = FastjsonUnmarshalListableString(fj, "rule_set")
	o.Invert, _ = FastjsonUnmarshalBool(fj, "invert")
	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	return nil
}

func (o *LogicalDNSRule) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Mode, _ = FastjsonUnmarshalString(fj, "mode")
	o.Rules = FastjsonUnmarshalArrayDNSRule(fj,"rules")
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
	o.LocalOptions.AutoReload, _ = FastjsonUnmarshalBool(fj, "auto_load")
	o.LocalOptions.Path, _ = FastjsonUnmarshalString(fj, "path")
	o.RemoteOptions.IsAsset, _ = FastjsonUnmarshalBool(fj, "is_asset")
	o.RemoteOptions.AutoReload, _ = FastjsonUnmarshalBool(fj, "auto_load")
	o.RemoteOptions.Path, _ = FastjsonUnmarshalString(fj, "path")
	o.RemoteOptions.URL, _ = FastjsonUnmarshalString(fj, "url")
	o.RemoteOptions.DownloadDetour, _ = FastjsonUnmarshalString(fj, "download_detour")
	o.RemoteOptions.UpdateInterval, _ = FastjsonUnmarshalBadoptionDuration(fj, "update_interval")
	return nil
}

func FastjsonUnmarshalDefaultRuleSet() RuleSet {
	return RuleSet{}
}

func FastjsonUnmarshalConvertRuleSet(fj *fastjson.Value) (RuleSet, error) {
	vv := RuleSet{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalOnRuleSet(fj *fastjson.Value, name string) (RuleSet, error) {
	return FastjsonUnmarshalAny[RuleSet](FastjsonUnmarshalConvertRuleSet,FastjsonUnmarshalDefaultRuleSet, fj, name)
}

func FastjsonUnmarshalArrayRuleSet(fj *fastjson.Value, name string) []RuleSet {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalOnRuleSet, fj, name, fastjson.TypeObject)
}

func FastjsonUnmarshalDefaultRule() Rule {
	return Rule{}
}

func FastjsonUnmarshalConvertRule(fj *fastjson.Value) (Rule, error) {
	vv := Rule{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalRule(fj *fastjson.Value, name string) (Rule, error) {
	return FastjsonUnmarshalAny[Rule](FastjsonUnmarshalConvertRule,FastjsonUnmarshalDefaultRule, fj, name)
}

func FastjsonUnmarshalArrayRule(fj *fastjson.Value, name string) []Rule {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalRule, fj, name, fastjson.TypeObject )
}

func FastjsonUnmarshalDefaultDefaultDNSRule() DefaultDNSRule {
	return DefaultDNSRule{}
}

func FastjsonUnmarshalConvertDefaultDNSRule(fj *fastjson.Value) (DefaultDNSRule, error) {
	vv := DefaultDNSRule{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalDefaultDNSRule() DNSRule {
	return DNSRule{}
}

func FastjsonUnmarshalConvertDNSRule(fj *fastjson.Value) (DNSRule, error) {
	vv := DNSRule{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalDNSRule(fj *fastjson.Value, name string) (DNSRule, error) {
	return FastjsonUnmarshalAny[DNSRule](FastjsonUnmarshalConvertDNSRule, FastjsonUnmarshalDefaultDNSRule, fj, name)
}

func FastjsonUnmarshalArrayDNSRule(fj *fastjson.Value, name string) []DNSRule {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalDNSRule, fj, name, fastjson.TypeObject)
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

	o.Inbound = FastjsonUnmarshalListableString(fj, "inbound")
	o.IPVersion, _ = FastjsonUnmarshalInt(fj, "ip_version")
	o.Network = FastjsonUnmarshalListableString(fj, "network")
	o.AuthUser = FastjsonUnmarshalListableString(fj, "auth_user")
	o.Protocol = FastjsonUnmarshalListableString(fj, "protocol")
	o.Domain = FastjsonUnmarshalListableString(fj, "domain")
	o.DomainSuffix = FastjsonUnmarshalListableString(fj, "domain_suffix")
	o.DomainKeyword = FastjsonUnmarshalListableString(fj, "domain_keyword")
	o.DomainRegex = FastjsonUnmarshalListableString(fj, "domain_regex")
	o.Geosite = FastjsonUnmarshalListableString(fj, "geosite")
	o.SourceGeoIP = FastjsonUnmarshalListableString(fj, "source_geoip")
	o.GeoIP = FastjsonUnmarshalListableString(fj, "geoip")
	o.SourceIPCIDR = FastjsonUnmarshalListableString(fj, "source_ip_cidr")
	o.SourceIPIsPrivate, _ = FastjsonUnmarshalBool(fj, "source_ip_is_private")
	o.IPCIDR = FastjsonUnmarshalListableString(fj, "ip_cidr")
	o.IPIsPrivate, _ = FastjsonUnmarshalBool(fj, "ip_is_private")
	o.SourcePort = FastjsonUnmarshalListableUint16(fj,"source_port")
	o.SourcePortRange = FastjsonUnmarshalListableString(fj, "source_port_range")
	o.Port = FastjsonUnmarshalListableUint16(fj, "port")
	o.PortRange = FastjsonUnmarshalListableString(fj, "port_range")
	o.ProcessName = FastjsonUnmarshalListableString(fj, "process_name")
	o.ProcessPath = FastjsonUnmarshalListableString(fj, "process_path")
	o.PackageName = FastjsonUnmarshalListableString(fj, "package_name")
	o.User = FastjsonUnmarshalListableString(fj, "user")
	o.UserID = FastjsonUnmarshalListableInt32(fj, "user_id")
	o.ClashMode, _ = FastjsonUnmarshalString(fj, "clash_mode")
	o.WIFISSID = FastjsonUnmarshalListableString(fj, "wifi_ssid")
	o.WIFIBSSID = FastjsonUnmarshalListableString(fj, "wifi_bssid")
	o.RuleSet = FastjsonUnmarshalListableString(fj, "rule_set")
	o.RuleSetIPCIDRMatchSource, _ = FastjsonUnmarshalBool(fj, "rule_set_ipcidr_match_source")
	o.Invert , _= FastjsonUnmarshalBool(fj, "invert")
	//o.Outbound, _ = FastjsonUnmarshalString(fj, "outbound")
	o.Name, _ = FastjsonUnmarshalString(fj, "name")
	return nil
}

func (o *LogicalRule) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Mode, _ = FastjsonUnmarshalString(fj, "mode")
	o.Rules = FastjsonUnmarshalArrayRule(fj, "rules")
	o.Invert, _ = FastjsonUnmarshalBool(fj, "invert")
	//o.Outbound = FastjsonUnmarshalString(fj, "outbound")
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
	o.Users = FastjsonUnmarshalArrayShadowsocksUser(fj, "users")
	o.Destinations = FastjsonUnmarshalArrayShadowsocksDestination(fj, "destinations")
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
	o.UDPOverTCP,_ =FastjsonUnmarshalUDPOverTCPOptions(fj, "udp_over_tcp")
	o.Multiplex ,_= FastjsonUnmarshalOutboundMultiplexOptions(fj, "multiplex")
	return nil
}

func FastjsonUnmarshalDefaultShadowsocksUser() ShadowsocksUser {
	return ShadowsocksUser{}
}

func FastjsonUnmarshalConvertShadowsocksUser(fj *fastjson.Value) (ShadowsocksUser, error) {
	vv := ShadowsocksUser{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalShadowsocksUser(fj *fastjson.Value, name string) (ShadowsocksUser, error) {
	return FastjsonUnmarshalAny[ShadowsocksUser](FastjsonUnmarshalConvertShadowsocksUser, FastjsonUnmarshalDefaultShadowsocksUser, fj, name)
}
 
func FastjsonUnmarshalArrayShadowsocksUser(fj *fastjson.Value, name string) []ShadowsocksUser {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalShadowsocksUser, fj, name, fastjson.TypeObject)
}

func FastjsonUnmarshalDefaultShadowsocksDestination() ShadowsocksDestination {
	return ShadowsocksDestination{}
}

func FastjsonUnmarshalConvertShadowsocksDestination(fj *fastjson.Value) (ShadowsocksDestination, error) {
	vv := ShadowsocksDestination{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalShadowsocksDestination(fj *fastjson.Value, name string) (ShadowsocksDestination, error) {
	return FastjsonUnmarshalAny[ShadowsocksDestination](FastjsonUnmarshalConvertShadowsocksDestination, FastjsonUnmarshalDefaultShadowsocksDestination, fj, name)
}

func FastjsonUnmarshalArrayShadowsocksDestination(fj *fastjson.Value, name string) []ShadowsocksDestination {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalShadowsocksDestination, fj, name, fastjson.TypeObject)
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
	o.Users = FastjsonUnmarshalArrayShadowTLSUser(fj, "users")
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
	o.TLS, _ =FastjsonUnmarshalOutboundTLSOptions(fj, "tls")
	return nil
}

func FastjsonUnmarshalDefaultShadowTLSUser() ShadowTLSUser {
	return ShadowTLSUser{}
}

func FastjsonUnmarshalConvertShadowTLSUser(fj *fastjson.Value) (ShadowTLSUser, error) {
	vv := ShadowTLSUser{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalShadowTLSUser(fj *fastjson.Value, name string) (ShadowTLSUser, error) {
	return FastjsonUnmarshalAny[ShadowTLSUser](FastjsonUnmarshalConvertShadowTLSUser, FastjsonUnmarshalDefaultShadowTLSUser, fj, name)
}

func FastjsonUnmarshalArrayShadowTLSUser(fj *fastjson.Value, name string) []ShadowTLSUser {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalShadowTLSUser, fj, name, fastjson.TypeObject)
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
	o.Users = FastjsonUnmarshalArrayAuthUser(fj,"users")
	return nil
}

func (o *HTTPMixedInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error {
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.ListenOptions.FastjsonUnmarshal(fj)
	o.Users = FastjsonUnmarshalArrayAuthUser(fj, "users")
	o.SetSystemProxy, _ = FastjsonUnmarshalBool(fj, "set_system_proxy")
	o.TLS , _ =FastjsonUnmarshalInboundTLSOptions(fj, "tls")
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
	o.TLS, _ =FastjsonUnmarshalOutboundTLSOptions(fj, "tls")
	o.Path, _ = FastjsonUnmarshalString(fj, "path")
	o.Headers, _ = FastjsonUnmarshalHTTPHeader(fj, "headers")
	return nil
}

// ssh.go
func FastjsonUnmarshalDefaultUDPOverTCPOptions() (UDPOverTCPOptions) {
	return UDPOverTCPOptions{}
}

func FastjsonUnmarshalConvertUDPOverTCPOptions(fj *fastjson.Value) (UDPOverTCPOptions, error) {
	vv := UDPOverTCPOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalUDPOverTCPOptions(fj *fastjson.Value, name string) (*UDPOverTCPOptions, error) {
	return FastjsonUnmarshalAnyPtr[UDPOverTCPOptions](FastjsonUnmarshalConvertUDPOverTCPOptions, fj,name)
}

func (o *SSHOutboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.DialerOptions.FastjsonUnmarshal(fj)
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.User, _ = FastjsonUnmarshalString(fj, "user")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	o.PrivateKey = FastjsonUnmarshalListableString(fj, "private_key")
	o.PrivateKeyPath, _ = FastjsonUnmarshalString(fj, "private_key_path")
	o.PrivateKeyPassphrase, _ = FastjsonUnmarshalString(fj, "private_key_passphrase")
	o.HostKey = FastjsonUnmarshalListableString(fj, "host_key")
	o.HostKeyAlgorithms = FastjsonUnmarshalListableString(fj, "host_key_algorithms")
	o.ClientVersion, _ = FastjsonUnmarshalString(fj, "client_version")
	o.UDPOverTCP,_ =FastjsonUnmarshalUDPOverTCPOptions(fj, "udp_over_tcp")
	
	return nil
}

// tls_acme.go

func FastjsonUnmarshalDefaultACMEExternalAccountOptions() (ACMEExternalAccountOptions) {
	return ACMEExternalAccountOptions{}
}

func FastjsonUnmarshalConvertACMEExternalAccountOptions(fj *fastjson.Value) (ACMEExternalAccountOptions, error) {
	vv := ACMEExternalAccountOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalACMEExternalAccountOptions(fj *fastjson.Value, name string) (*ACMEExternalAccountOptions, error) {
	return FastjsonUnmarshalAnyPtr[ACMEExternalAccountOptions](FastjsonUnmarshalConvertACMEExternalAccountOptions, fj,name)
}

func FastjsonUnmarshalDefaultACMEDNS01ChallengeOptions() (ACMEDNS01ChallengeOptions) {
	return ACMEDNS01ChallengeOptions{}
}

func FastjsonUnmarshalConvertACMEDNS01ChallengeOptions(fj *fastjson.Value) (ACMEDNS01ChallengeOptions, error) {
	vv := ACMEDNS01ChallengeOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalACMEDNS01ChallengeOptions(fj *fastjson.Value, name string) (*ACMEDNS01ChallengeOptions, error) {
	return FastjsonUnmarshalAnyPtr[ACMEDNS01ChallengeOptions](FastjsonUnmarshalConvertACMEDNS01ChallengeOptions, fj,name)
}

func (o *InboundACMEOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Domain = FastjsonUnmarshalListableString(fj, "domain")
	o.DataDirectory, _ = FastjsonUnmarshalString(fj, "data_directory")
	o.DefaultServerName, _ = FastjsonUnmarshalString(fj, "default_server_name")
	o.Email, _ = FastjsonUnmarshalString(fj, "email")
	o.Provider, _ = FastjsonUnmarshalString(fj, "provider")
	o.DisableHTTPChallenge, _ = FastjsonUnmarshalBool(fj, "disable_http_challenge")
	o.DisableTLSALPNChallenge, _ = FastjsonUnmarshalBool(fj, "disable_tls_alpn_challenge")
	o.AlternativeHTTPPort, _ = FastjsonUnmarshalUint16(fj, "alternative_http_port")
	o.AlternativeTLSPort, _ = FastjsonUnmarshalUint16(fj, "alternative_tls_port")
	o.ExternalAccount ,_= FastjsonUnmarshalACMEExternalAccountOptions(fj, "external_account")
	o.DNS01Challenge ,_= FastjsonUnmarshalACMEDNS01ChallengeOptions(fj, "dns01_challenge")
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
func FastjsonUnmarshalDefaultInboundACMEOptions() (InboundACMEOptions) {
	return InboundACMEOptions{}
}

func FastjsonUnmarshalConvertInboundACMEOptions(fj *fastjson.Value) (InboundACMEOptions, error) {
	vv := InboundACMEOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalInboundACMEOptions(fj *fastjson.Value, name string) (*InboundACMEOptions, error) {
	return FastjsonUnmarshalAnyPtr[InboundACMEOptions](FastjsonUnmarshalConvertInboundACMEOptions, fj,name)
}

func FastjsonUnmarshalDefaultInboundECHOptions() (InboundECHOptions) {
	return InboundECHOptions{}
}

func FastjsonUnmarshalConvertInboundECHOptions(fj *fastjson.Value) (InboundECHOptions, error) {
	vv := InboundECHOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalInboundECHOptions(fj *fastjson.Value, name string) (*InboundECHOptions, error) {
	return FastjsonUnmarshalAnyPtr[InboundECHOptions](FastjsonUnmarshalConvertInboundECHOptions, fj,name)
}

func FastjsonUnmarshalDefaultInboundRealityOptions() (InboundRealityOptions) {
	return InboundRealityOptions{}
}

func FastjsonUnmarshalConvertInboundRealityOptions(fj *fastjson.Value) (InboundRealityOptions, error) {
	vv := InboundRealityOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalInboundRealityOptions(fj *fastjson.Value, name string) (*InboundRealityOptions, error) {
	return FastjsonUnmarshalAnyPtr[InboundRealityOptions](FastjsonUnmarshalConvertInboundRealityOptions, fj,name)
}

func (o *InboundTLSOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.ServerName, _ = FastjsonUnmarshalString(fj, "server_name")
	o.Insecure, _ = FastjsonUnmarshalBool(fj, "insecure")
	o.ALPN = FastjsonUnmarshalListableString(fj, "alpn")
	o.MinVersion, _ = FastjsonUnmarshalString(fj, "min_version")
	o.MaxVersion, _ = FastjsonUnmarshalString(fj, "max_version")
	o.CipherSuites = FastjsonUnmarshalListableString(fj, "cipher_suites")
	o.Certificate = FastjsonUnmarshalListableString(fj, "certificate")
	o.CertificatePath, _ = FastjsonUnmarshalString(fj, "certificate_path")
	o.Key = FastjsonUnmarshalListableString(fj, "key")
	o.KeyPath, _ = FastjsonUnmarshalString(fj, "key_path")
	o.ACME ,_= FastjsonUnmarshalInboundACMEOptions(fj, "acme")
	o.ECH ,_= FastjsonUnmarshalInboundECHOptions(fj, "ech")
	o.Reality ,_= FastjsonUnmarshalInboundRealityOptions(fj, "reality")
	 
	return nil
}

func FastjsonUnmarshalDefaultInboundTLSOptions(fj *fastjson.Value) InboundTLSOptions {
	return InboundTLSOptions{}
}

func FastjsonUnmarshalConvertInboundTLSOptions(fj *fastjson.Value) (InboundTLSOptions, error) {
	vv := InboundTLSOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalInboundTLSOptions(fj *fastjson.Value, name string) (*InboundTLSOptions, error) {
	return FastjsonUnmarshalAnyPtr[InboundTLSOptions](FastjsonUnmarshalConvertInboundTLSOptions, fj,name)
}

func FastjsonUnmarshalDefaultOutboundECHOptions() (OutboundECHOptions) {
	return OutboundECHOptions{}
}

func FastjsonUnmarshalConvertOutboundECHOptions(fj *fastjson.Value) (OutboundECHOptions, error) {
	vv := OutboundECHOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalOutboundECHOptions(fj *fastjson.Value, name string) (*OutboundECHOptions, error) {
	return FastjsonUnmarshalAnyPtr[OutboundECHOptions](FastjsonUnmarshalConvertOutboundECHOptions, fj,name)
}

func FastjsonUnmarshalDefaultOutboundUTLSOptions() (OutboundUTLSOptions) {
	return OutboundUTLSOptions{}
}

func FastjsonUnmarshalConvertOutboundUTLSOptions(fj *fastjson.Value) (OutboundUTLSOptions, error) {
	vv := OutboundUTLSOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalOutboundUTLSOptions(fj *fastjson.Value, name string) (*OutboundUTLSOptions, error) {
	return FastjsonUnmarshalAnyPtr[OutboundUTLSOptions](FastjsonUnmarshalConvertOutboundUTLSOptions, fj,name)
}

func FastjsonUnmarshalDefaultOutboundRealityOptions() (OutboundRealityOptions) {
	return OutboundRealityOptions{}
}

func FastjsonUnmarshalConvertOutboundRealityOptions(fj *fastjson.Value) (OutboundRealityOptions, error) {
	vv := OutboundRealityOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalOutboundRealityOptions(fj *fastjson.Value, name string) (*OutboundRealityOptions, error) {
	return FastjsonUnmarshalAnyPtr[OutboundRealityOptions](FastjsonUnmarshalConvertOutboundRealityOptions, fj,name)
}

func (o *OutboundTLSOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.DisableSNI, _ = FastjsonUnmarshalBool(fj, "disable_sni")
	o.ServerName, _ = FastjsonUnmarshalString(fj, "server_name")
	o.Insecure, _ = FastjsonUnmarshalBool(fj, "insecure")
	o.ALPN = FastjsonUnmarshalListableString(fj, "alpn")
	o.MinVersion, _ = FastjsonUnmarshalString(fj, "min_version")
	o.MaxVersion, _ = FastjsonUnmarshalString(fj, "max_version")
	o.CipherSuites = FastjsonUnmarshalListableString(fj, "cipher_suites")
	o.Certificate = FastjsonUnmarshalListableString(fj, "certificate")
	o.CertificatePath, _ = FastjsonUnmarshalString(fj, "certificate_path")
	o.ECH,_ = FastjsonUnmarshalOutboundECHOptions(fj, "ech")
	o.UTLS,_ = FastjsonUnmarshalOutboundUTLSOptions(fj, "utls")
	o.Reality,_ =FastjsonUnmarshalOutboundRealityOptions(fj, "reality")
	o.TLSTricks,_ = FastjsonUnmarshalTLSTricksOptions(fj, "tls_tricks")

	return nil
}

func FastjsonUnmarshalConvertOutboundTLSOptions(fj *fastjson.Value) (OutboundTLSOptions, error) {
	vv := OutboundTLSOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalOutboundTLSOptions(fj *fastjson.Value, name string) (*OutboundTLSOptions, error) {
    return FastjsonUnmarshalAnyPtr[OutboundTLSOptions](FastjsonUnmarshalConvertOutboundTLSOptions, fj,name)
}

func (o *InboundRealityOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.Handshake.FastjsonUnmarshal(fj.Get("handshake"))
	o.PrivateKey, _ = FastjsonUnmarshalString(fj, "private_key")
	o.ShortID = FastjsonUnmarshalListableString(fj, "short_id")
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
	o.Key = FastjsonUnmarshalListableString(fj, "key")
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
	o.Config = FastjsonUnmarshalListableString(fj, "config")
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
func FastjsonUnmarshalDefaultServerOptions() (ServerOptions) {
	return ServerOptions{}
}

func FastjsonUnmarshalConvertServerOptions(fj *fastjson.Value) (ServerOptions, error) {
	vv := ServerOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalServerOptions(fj *fastjson.Value, name string) (*ServerOptions, error) {
	return FastjsonUnmarshalAnyPtr[ServerOptions](FastjsonUnmarshalConvertServerOptions, fj,name)
}

func (o *TrojanInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.ListenOptions.FastjsonUnmarshal(fj)
	o.Users = FastjsonUnmarshalArrayTrojanUser(fj,"users")
	o.TLS , _ =FastjsonUnmarshalInboundTLSOptions(fj, "tls")
	o.Fallback,_ =FastjsonUnmarshalServerOptions(fj, "fallback")
	o.FallbackForALPN = FastjsonUnmarshalMapServerOptions(fj.GetArray("fallback_for_alpn"))
	o.Transport ,_= FastjsonUnmarshalV2RayTransportOptions(fj, "transport")
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

func FastjsonUnmarshalDefaultOutboundMultiplexOptions() (OutboundMultiplexOptions) {
	return OutboundMultiplexOptions{}
}

func FastjsonUnmarshalConvertOutboundMultiplexOptions(fj *fastjson.Value) (OutboundMultiplexOptions, error) {
	vv := OutboundMultiplexOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalOutboundMultiplexOptions(fj *fastjson.Value, name string) (*OutboundMultiplexOptions, error) {
	return FastjsonUnmarshalAnyPtr[OutboundMultiplexOptions](FastjsonUnmarshalConvertOutboundMultiplexOptions, fj,name)
}
//
func FastjsonUnmarshalDefaultV2RayTransportOptions() (V2RayTransportOptions) {
	return V2RayTransportOptions{}
}

func FastjsonUnmarshalConvertV2RayTransportOptions(fj *fastjson.Value) (V2RayTransportOptions, error) {
	vv := V2RayTransportOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalV2RayTransportOptions(fj *fastjson.Value, name string) (*V2RayTransportOptions, error) {
	return FastjsonUnmarshalAnyPtr[V2RayTransportOptions](FastjsonUnmarshalConvertV2RayTransportOptions, fj,name)
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
	o.TLS, _ =FastjsonUnmarshalOutboundTLSOptions(fj, "tls")
	o.Multiplex ,_= FastjsonUnmarshalOutboundMultiplexOptions(fj, "multiplex")
	o.Transport ,_= FastjsonUnmarshalV2RayTransportOptions(fj, "transport")
	return nil
}

func FastjsonUnmarshalDefaultTrojanUser() TrojanUser {
	return TrojanUser{}
}

func FastjsonUnmarshalConvertTrojanUser(fj *fastjson.Value) (TrojanUser, error) {
	vv := TrojanUser{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalTrojanUser(fj *fastjson.Value, name string) (TrojanUser, error) {
	return FastjsonUnmarshalAny[TrojanUser](FastjsonUnmarshalConvertTrojanUser, FastjsonUnmarshalDefaultTrojanUser, fj, name)
}

func FastjsonUnmarshalArrayTrojanUser(fj *fastjson.Value, name string) []TrojanUser {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalTrojanUser, fj, name, fastjson.TypeObject)
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
	o.Users = FastjsonUnmarshalArrayTUICUser(fj, "users")
	o.CongestionControl, _ = FastjsonUnmarshalString(fj, "congestion_control")
	o.AuthTimeout, _ = FastjsonUnmarshalBadoptionDuration(fj, "auth_timeout")
	o.ZeroRTTHandshake, _ = FastjsonUnmarshalBool(fj, "zero_rtt_handshake")
	o.Heartbeat, _ = FastjsonUnmarshalBadoptionDuration(fj, "heartbeat")
	o.TLS , _ =FastjsonUnmarshalInboundTLSOptions(fj, "tls")
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
	o.TLS, _ =FastjsonUnmarshalOutboundTLSOptions(fj, "tls")
	o.TurnRelay, _= FastjsonUnmarshalTurnRelayOptions(fj, "turn_relay")
	return nil
}

func FastjsonUnmarshalDefaultTUICUser() TUICUser {
	return TUICUser{}
}

func FastjsonUnmarshalConvertTUICUser(fj *fastjson.Value) (TUICUser, error) {
	vv := TUICUser{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalTUICUser(fj *fastjson.Value, name string) (TUICUser, error) {
	return FastjsonUnmarshalAny[TUICUser](FastjsonUnmarshalConvertTUICUser, FastjsonUnmarshalDefaultTUICUser, fj, name)
}

func FastjsonUnmarshalArrayTUICUser(fj *fastjson.Value, name string) []TUICUser {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalTUICUser, fj, name, fastjson.TypeObject)
}

// tun_platform.go
func FastjsonUnmarshalDefaultHTTPProxyOptions() (HTTPProxyOptions) {
	return HTTPProxyOptions{}
}

func FastjsonUnmarshalConvertHTTPProxyOptions(fj *fastjson.Value) (HTTPProxyOptions, error) {
	vv := HTTPProxyOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalHTTPProxyOptions(fj *fastjson.Value, name string) (*HTTPProxyOptions, error) {
	return FastjsonUnmarshalAnyPtr[HTTPProxyOptions](FastjsonUnmarshalConvertHTTPProxyOptions, fj,name)
}

func (o *TunPlatformOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.HTTPProxy,_ = FastjsonUnmarshalHTTPProxyOptions(fj, "http_proxy")
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
func FastjsonUnmarshalDefaultTunPlatformOptions() (TunPlatformOptions) {
	return TunPlatformOptions{}
}

func FastjsonUnmarshalConvertTunPlatformOptions(fj *fastjson.Value) (TunPlatformOptions, error) {
	vv := TunPlatformOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalTunPlatformOptions(fj *fastjson.Value, name string) (*TunPlatformOptions, error) {
	return FastjsonUnmarshalAnyPtr[TunPlatformOptions](FastjsonUnmarshalConvertTunPlatformOptions, fj,name)
}

func (o *TunInboundOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.InterfaceName, _ = FastjsonUnmarshalString(fj, "interface_name")
	o.MTU, _ = FastjsonUnmarshalUint32(fj, "mtu") 
	o.GSO, _ = FastjsonUnmarshalBool(fj, "gso")
	o.Inet4Address = FastjsonUnmarshalListableNetipPrefix(fj,"inet4_address")
	o.Inet6Address = FastjsonUnmarshalListableNetipPrefix(fj,"inet6_address")
	o.AutoRoute, _ = FastjsonUnmarshalBool(fj, "auto_route")
	o.StrictRoute, _ = FastjsonUnmarshalBool(fj, "strict_route")
	o.Inet4RouteAddress = FastjsonUnmarshalListableNetipPrefix(fj,"inet4_route_address")
	o.Inet6RouteAddress = FastjsonUnmarshalListableNetipPrefix(fj,"inet6_route_address")
	o.Inet4RouteExcludeAddress = FastjsonUnmarshalListableNetipPrefix(fj,"inet4_route_exclude_address")
	o.Inet6RouteExcludeAddress = FastjsonUnmarshalListableNetipPrefix(fj,"inet6_route_exclude_address")
	o.IncludeInterface = FastjsonUnmarshalListableString(fj, "include_interface")
	o.ExcludeInterface = FastjsonUnmarshalListableString(fj, "exclude_interface")
	o.IncludeUID = FastjsonUnmarshalListableUInt32(fj,"include_uid")
	o.IncludeUIDRange = FastjsonUnmarshalListableString(fj, "include_uid_range")
	o.ExcludeUID = FastjsonUnmarshalListableUInt32(fj,"exclude_uid")
	o.ExcludeUIDRange = FastjsonUnmarshalListableString(fj, "exclude_uid_range")
	o.IncludeAndroidUser = FastjsonUnmarshalListableInt(fj, "include_android_user")
	o.IncludePackage = FastjsonUnmarshalListableString(fj, "include_package")
	o.ExcludePackage = FastjsonUnmarshalListableString(fj, "exclude_package")
	o.EndpointIndependentNat, _ = FastjsonUnmarshalBool(fj, "endpoint_independent_nat")
	udpimeout, _ := FastjsonUnmarshalBadoptionDuration(fj, "udp_timeout")
	o.UDPTimeout = UDPTimeoutCompat(udpimeout)
	o.Stack, _ = FastjsonUnmarshalString(fj, "stack")
	o.Platform ,_ = FastjsonUnmarshalTunPlatformOptions(fj, "platform")
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

	o.Host = FastjsonUnmarshalListableString(fj, "host")
	o.Path, _ = FastjsonUnmarshalString(fj, "path")
	o.Method, _ = FastjsonUnmarshalString(fj, "method")
	o.Headers, _ = FastjsonUnmarshalHTTPHeader(fj, "headers")
	o.IdleTimeout, _ = FastjsonUnmarshalBadoptionDuration(fj, "idle_timeout")
	o.PingTimeout, _ = FastjsonUnmarshalBadoptionDuration(fj, "ping_timeout")
	return nil
}

func (o *V2RayWebsocketOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}

	o.Path, _ = FastjsonUnmarshalString(fj, "path")
	o.Headers, _ = FastjsonUnmarshalHTTPHeader(fj, "headers")
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
	o.Headers, _ = FastjsonUnmarshalHTTPHeader(fj, "headers")
	return nil
}

// v2ray.go
func FastjsonUnmarshalDefaultVLESSUser() VLESSUser {
	return VLESSUser{}
}

func FastjsonUnmarshalConvertVLESSUser(fj *fastjson.Value) (VLESSUser, error) {
	vv := VLESSUser{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalVLESSUser(fj *fastjson.Value, name string) (VLESSUser, error) {
	return FastjsonUnmarshalAny[VLESSUser](FastjsonUnmarshalConvertVLESSUser, FastjsonUnmarshalDefaultVLESSUser, fj, name)
}

func FastjsonUnmarshalArrayVLESSUser(fj *fastjson.Value, name string) []VLESSUser {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalVLESSUser, fj, name, fastjson.TypeObject)
}
 
func FastjsonUnmarshalDefaultV2RayStatsServiceOptions() (V2RayStatsServiceOptions) {
	return V2RayStatsServiceOptions{}
}

func FastjsonUnmarshalConvertV2RayStatsServiceOptions(fj *fastjson.Value) (V2RayStatsServiceOptions, error) {
	vv := V2RayStatsServiceOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalV2RayStatsServiceOptions(fj *fastjson.Value, name string) (*V2RayStatsServiceOptions, error) {
	return FastjsonUnmarshalAnyPtr[V2RayStatsServiceOptions](FastjsonUnmarshalConvertV2RayStatsServiceOptions, fj,name)
}

func (o *V2RayAPIOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	if fj == nil || fj.Type() != fastjson.TypeObject {
		return nil
	}
	o.Listen, _ = FastjsonUnmarshalString(fj, "listen")
	o.Stats, _ = FastjsonUnmarshalV2RayStatsServiceOptions(fj, "stats")
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
	o.Users = FastjsonUnmarshalArrayVLESSUser(fj, "users")
	o.TLS , _ =FastjsonUnmarshalInboundTLSOptions(fj, "tls")
	o.Transport ,_= FastjsonUnmarshalV2RayTransportOptions(fj, "transport")
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
	o.TLS, _ =FastjsonUnmarshalOutboundTLSOptions(fj, "tls")
	o.Multiplex ,_= FastjsonUnmarshalOutboundMultiplexOptions(fj, "multiplex")
	o.Transport ,_= FastjsonUnmarshalV2RayTransportOptions(fj, "transport")
	o.PacketEncoding, _ = FastjsonUnmarshalAnyPtr[string](FastjsonUnmarshalConvertString, fj, "packet_encoding")
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
	o.TLS, _ =FastjsonUnmarshalOutboundTLSOptions(fj, "tls")
	o.PacketEncoding, _ = FastjsonUnmarshalString(fj, "packet_encoding")
	o.Multiplex ,_= FastjsonUnmarshalOutboundMultiplexOptions(fj, "multiplex")
	o.Transport ,_= FastjsonUnmarshalV2RayTransportOptions(fj, "transport")
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
	o.LocalAddress = FastjsonUnmarshalListableNetipPrefix(fj,"local_address")
	o.PrivateKey, _ = FastjsonUnmarshalString(fj, "private_key")
	o.Peers = FastjsonUnmarshalArrayWireGuardPeer(fj,"peers")
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.PeerPublicKey, _ = FastjsonUnmarshalString(fj, "peer_public_key")
	o.PreSharedKey, _ = FastjsonUnmarshalString(fj, "pre_shared_key")
	o.Reserved = FastjsonUnmarshalListableUint8(fj,"reserved")
	o.Workers, _ = FastjsonUnmarshalInt(fj, "workers")
	o.MTU = uint32(fj.GetUint("mtu"))
	networkList, _ := FastjsonUnmarshalString(fj, "network")
	o.Network = NetworkList(networkList)
	o.TurnRelay, _= FastjsonUnmarshalTurnRelayOptions(fj, "turn_relay")
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
	o.AllowedIPs = FastjsonUnmarshalListableNetipPrefix(fj,"allowed_ips")
	o.Reserved = FastjsonUnmarshalListableUint8(fj,"reserved")
	return nil
}

func FastjsonUnmarshalDefaultLegacyWireGuardPeer() LegacyWireGuardPeer {
	return LegacyWireGuardPeer{}
}

func FastjsonUnmarshalConvertLegacyWireGuardPeer(fj *fastjson.Value) (LegacyWireGuardPeer, error) {
	vv := LegacyWireGuardPeer{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalLegacyWireGuardPeer(fj *fastjson.Value, name string) (LegacyWireGuardPeer, error) {
	return FastjsonUnmarshalAny[LegacyWireGuardPeer](FastjsonUnmarshalConvertLegacyWireGuardPeer, FastjsonUnmarshalDefaultLegacyWireGuardPeer, fj, name)
}


func FastjsonUnmarshalArrayWireGuardPeer(fj *fastjson.Value, name string) []LegacyWireGuardPeer {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalLegacyWireGuardPeer, fj, name, fastjson.TypeObject)
}

func FastjsonUnmarshalConvertNetipPrefix(fj *fastjson.Value) (netip.Prefix, error) {
	return FastjsonUnmarshalNetipPrefix(fj, "")
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

func FastjsonUnmarshalListableNetipPrefix(fj *fastjson.Value, name string) badoption.Listable[netip.Prefix] {
	return FastjsonUnmarshalArrayT(FastjsonUnmarshalNetipPrefix, fj, name, fastjson.TypeObject )
}

func (o *TLSFragmentOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	o.Enabled, _ = FastjsonUnmarshalBool(fj, "enabled")
	o.Size, _ = FastjsonUnmarshalString(fj, "size")
	o.Sleep, _ = FastjsonUnmarshalString(fj, "sleep")
	return nil
}

func FastjsonUnmarshalConvertTLSFragmentOptions(fj *fastjson.Value) (TLSFragmentOptions, error) {
	vv := TLSFragmentOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalTLSFragmentOptions(fj *fastjson.Value, name string) (*TLSFragmentOptions, error) {
   return FastjsonUnmarshalAnyPtr[TLSFragmentOptions](FastjsonUnmarshalConvertTLSFragmentOptions, fj,name)
}
 
func (o *TurnRelayOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	o.ServerOptions.FastjsonUnmarshal(fj)
	o.Username, _ = FastjsonUnmarshalString(fj, "username")
	o.Password, _ = FastjsonUnmarshalString(fj, "password")
	o.Realm, _ = FastjsonUnmarshalString(fj, "realm")
	return nil
}

func FastjsonUnmarshalConvertTurnRelayOptions(fj *fastjson.Value) (TurnRelayOptions, error) {
	vv := TurnRelayOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalTurnRelayOptions(fj *fastjson.Value, name string) (*TurnRelayOptions, error) {
	return FastjsonUnmarshalAnyPtr[TurnRelayOptions](FastjsonUnmarshalConvertTurnRelayOptions, fj,name)
}

func (o *TLSTricksOptions) FastjsonUnmarshal(fj *fastjson.Value) error{
	o.MixedCaseSNI, _ = FastjsonUnmarshalBool(fj, "mixedcase_sni")
	o.PaddingMode, _ = FastjsonUnmarshalString(fj, "padding_mode")
	o.PaddingSize, _ = FastjsonUnmarshalString(fj, "padding_size")
	o.PaddingSNI, _ = FastjsonUnmarshalString(fj, "padding_sni")
	return nil
}

func FastjsonUnmarshalConvertTLSTricksOptions(fj *fastjson.Value) (TLSTricksOptions, error) {
	vv := TLSTricksOptions{}
	err := vv.FastjsonUnmarshal(fj)
	return vv, err
}

func FastjsonUnmarshalTLSTricksOptions(fj *fastjson.Value, name string) (*TLSTricksOptions, error) {
	return FastjsonUnmarshalAnyPtr[TLSTricksOptions](FastjsonUnmarshalConvertTLSTricksOptions, fj,name)
}