package route

import (
	"context"
	"net/netip"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/common/geoip"
	"github.com/sagernet/sing-box/common/geosite"
	"github.com/sagernet/sing-box/common/process"
	"github.com/sagernet/sing-box/common/taskmonitor"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/experimental/libbox/platform"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	R "github.com/sagernet/sing-box/route/rule"
	"github.com/sagernet/sing-box/transport/fakeip"
	"github.com/sagernet/sing-dns"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/task"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
)

var _ adapter.Router = (*Router)(nil)

type Router struct {
	ctx                     context.Context
	logger                  log.ContextLogger
	dnsLogger               log.ContextLogger
	inbound                 adapter.InboundManager
	outbound                adapter.OutboundManager
	connection              adapter.ConnectionManager
	network                 adapter.NetworkManager
	rules                   []adapter.Rule
	needGeoIPDatabase       bool
	needGeositeDatabase     bool
	geoIPOptions            option.GeoIPOptions
	geositeOptions          option.GeositeOptions
	geoIPReader             *geoip.Reader
	geositeReader           *geosite.Reader
	geositeCache            map[string]adapter.Rule
	needFindProcess         bool
	dnsClient               *dns.Client
	staticDns               map[string]StaticDNSEntry //hiddify
	defaultDomainStrategy   dns.DomainStrategy
	dnsRules                []adapter.DNSRule
	ruleSetsRemoteWithLocal []adapter.RuleSet //karing
	ruleSets                []adapter.RuleSet
	ruleSetMap              map[string]adapter.RuleSet
	defaultTransport        dns.Transport
	transports              []dns.Transport
	transportMap            map[string]dns.Transport
	transportDomainStrategy map[dns.Transport]dns.DomainStrategy
	dnsReverseMapping       *DNSReverseMapping
	fakeIPStore             adapter.FakeIPStore
	processSearcher         process.Searcher
	pauseManager            pause.Manager
	tracker                 adapter.ConnectionTracker
	platformInterface       platform.Interface
	needWIFIState           bool
	started                 bool
	quitSig                 func()   //karing
}

func NewRouter(ctx context.Context, logFactory log.Factory, options option.RouteOptions, dnsOptions option.DNSOptions, quitSig func()) (*Router, error) { //karing
	router := &Router{
		ctx:                   ctx,
		logger:                logFactory.NewLogger("router"),
		dnsLogger:             logFactory.NewLogger("dns"),
		inbound:               service.FromContext[adapter.InboundManager](ctx),
		outbound:              service.FromContext[adapter.OutboundManager](ctx),
		connection:            service.FromContext[adapter.ConnectionManager](ctx),
		network:               service.FromContext[adapter.NetworkManager](ctx),
		rules:                 make([]adapter.Rule, 0, len(options.Rules)),
		dnsRules:              make([]adapter.DNSRule, 0, len(dnsOptions.Rules)),
		ruleSetMap:            make(map[string]adapter.RuleSet),
		needGeoIPDatabase:     hasRule(options.Rules, isGeoIPRule) || hasDNSRule(dnsOptions.Rules, isGeoIPDNSRule),
		needGeositeDatabase:   hasRule(options.Rules, isGeositeRule) || hasDNSRule(dnsOptions.Rules, isGeositeDNSRule),
		geoIPOptions:          common.PtrValueOrDefault(options.GeoIP),
		geositeOptions:        common.PtrValueOrDefault(options.Geosite),
		geositeCache:          make(map[string]adapter.Rule),
		needFindProcess:       hasRule(options.Rules, isProcessRule) || hasDNSRule(dnsOptions.Rules, isProcessDNSRule) || options.FindProcess,
		defaultDomainStrategy: dns.DomainStrategy(dnsOptions.Strategy),
		pauseManager:          service.FromContext[pause.Manager](ctx),
		platformInterface:     service.FromContext[platform.Interface](ctx),
		needWIFIState:         hasRule(options.Rules, isWIFIRule) || hasDNSRule(dnsOptions.Rules, isWIFIDNSRule),
		staticDns:             createEntries(dnsOptions.StaticIPs), //hiddify
		quitSig:               quitSig, //karing
	}
	service.MustRegister[adapter.Router](ctx, router)
	router.dnsClient = dns.NewClient(dns.ClientOptions{
		DisableCache:     dnsOptions.DNSClientOptions.DisableCache,
		DisableExpire:    dnsOptions.DNSClientOptions.DisableExpire,
		IndependentCache: dnsOptions.DNSClientOptions.IndependentCache,
		CacheCapacity:    dnsOptions.DNSClientOptions.CacheCapacity,
		RDRC: func() dns.RDRCStore {
			cacheFile := service.FromContext[adapter.CacheFile](ctx)
			if cacheFile == nil {
				return nil
			}
			if !cacheFile.StoreRDRC() {
				return nil
			}
			return cacheFile
		},
		Logger: router.dnsLogger,
	})
	for i, ruleOptions := range options.Rules {
		routeRule, err := R.NewRule(ctx, router.logger, ruleOptions, true)
		if err != nil {
			return nil, E.Cause(err, "parse rule[", i, "]")
		}
		router.rules = append(router.rules, routeRule)
	}
	for i, dnsRuleOptions := range dnsOptions.Rules {
		dnsRule, err := R.NewDNSRule(ctx, router.logger, dnsRuleOptions, true)
		if err != nil {
			return nil, E.Cause(err, "parse dns rule[", i, "]")
		}
		router.dnsRules = append(router.dnsRules, dnsRule)
	}
	for i, ruleSetOptions := range options.RuleSet {
		if _, exists := router.ruleSetMap[ruleSetOptions.Tag]; exists {
			return nil, E.New("duplicate rule-set tag: ", ruleSetOptions.Tag)
		}
		if ruleSetOptions.Type == C.RuleSetTypeRemote { //karing
			if len(ruleSetOptions.LocalOptions.Path) != 0 {
				cacheFile := service.FromContext[adapter.CacheFile](ctx)
				if cacheFile != nil {
					if !cacheFile.HasRuleSet(ruleSetOptions.RemoteOptions.URL) {   //karing
						ruleSet := R.NewRemoteRuleSet(ctx, router, router.logger, ruleSetOptions)
						router.ruleSetsRemoteWithLocal = append(router.ruleSetsRemoteWithLocal, ruleSet)
						ruleSetOptions.Type = C.RuleSetTypeLocal
					}
				}
			}
		}
		ruleSet, err := R.NewRuleSet(ctx, router.logger, ruleSetOptions)
		if err != nil {
			return nil, E.Cause(err, "parse rule-set[", i, "]")
		}
		router.ruleSets = append(router.ruleSets, ruleSet)
		router.ruleSetMap[ruleSetOptions.Tag] = ruleSet
	}

	transports := make([]dns.Transport, len(dnsOptions.Servers))
	dummyTransportMap := make(map[string]dns.Transport)
	transportMap := make(map[string]dns.Transport)
	transportTags := make([]string, len(dnsOptions.Servers))
	transportTagMap := make(map[string]bool)
	transportDomainStrategy := make(map[dns.Transport]dns.DomainStrategy)
	for i, server := range dnsOptions.Servers {
		var tag string
		if server.Tag != "" {
			tag = server.Tag
		} else {
			tag = F.ToString(i)
		}
		if transportTagMap[tag] {
			return nil, E.New("duplicate dns server tag: ", tag)
		}
		transportTags[i] = tag
		transportTagMap[tag] = true
	}
	outboundManager := service.FromContext[adapter.OutboundManager](ctx)
	for {
		lastLen := len(dummyTransportMap)
		for i, server := range dnsOptions.Servers {
			tag := transportTags[i]
			if _, exists := dummyTransportMap[tag]; exists {
				continue
			}
			var detour N.Dialer
			if server.Detour == "" {
				detour = dialer.NewDefaultOutbound(outboundManager)
			} else {
				detour = dialer.NewDetour(outboundManager, server.Detour)
			}
			var serverProtocol string
			if len(server.Addresses) > 0 { //karing
				var toContinue = false
				var detoured = false
				for _, address := range server.Addresses {
					switch address {
					case "local":
					default:
						serverURL, _ := url.Parse(address)
						var serverAddress string
						if serverURL != nil {
							serverAddress = serverURL.Hostname()
						}
						if serverAddress == "" {
							serverAddress = address
						}
						notIpAddress := !M.ParseSocksaddr(serverAddress).Addr.IsValid()
						if server.AddressResolver != "" {
							if !transportTagMap[server.AddressResolver] {
								return nil, E.New("parse dns server[", tag, "]: address resolver not found: ", server.AddressResolver)
							}
							if upstream, exists := dummyTransportMap[server.AddressResolver]; exists {
								if !detoured {
									detoured = true
									detour = dns.NewDialerWrapper(detour, router.dnsClient, upstream, dns.DomainStrategy(server.AddressStrategy), time.Duration(server.AddressFallbackDelay))
								}
							} else {
								toContinue = true
								break
							}
						} else if notIpAddress && strings.Contains(address, ".") {
							return nil, E.New("parse dns server[", tag, "]: missing address_resolver")
						}

					}
				}
				if toContinue{
					continue
				}
			} else {
				switch server.Address {
				case "local":
					serverProtocol = "local"
				default:
					serverURL, _ := url.Parse(server.Address)
					var serverAddress string
					if serverURL != nil {
						if serverURL.Scheme == "" {
							serverProtocol = "udp"
						} else {
							serverProtocol = serverURL.Scheme
						}
						serverAddress = serverURL.Hostname()
					}
					if serverAddress == "" {
						serverAddress = server.Address
					}
					notIpAddress := !M.ParseSocksaddr(serverAddress).Addr.IsValid()
					if server.AddressResolver != "" {
						if !transportTagMap[server.AddressResolver] {
							return nil, E.New("parse dns server[", tag, "]: address resolver not found: ", server.AddressResolver)
							}
							if upstream, exists := dummyTransportMap[server.AddressResolver]; exists {
								detour = dns.NewDialerWrapper(detour, router.dnsClient, upstream, dns.DomainStrategy(server.AddressStrategy), time.Duration(server.AddressFallbackDelay))
							} else {
								continue
							}
						} else if notIpAddress && strings.Contains(server.Address, ".") {
							return nil, E.New("parse dns server[", tag, "]: missing address_resolver")
						}
					}
			}
			
			var clientSubnet netip.Prefix
			if server.ClientSubnet != nil {
				clientSubnet = netip.Prefix(common.PtrValueOrDefault(server.ClientSubnet))
			} else if dnsOptions.ClientSubnet != nil {
				clientSubnet = netip.Prefix(common.PtrValueOrDefault(dnsOptions.ClientSubnet))
			}
			if serverProtocol == "" {
				serverProtocol = "transport"
			}
			transport, err := dns.CreateTransport(dns.TransportOptions{
				Context:      ctx,
				Logger:       logFactory.NewLogger(F.ToString("dns/", serverProtocol, "[", tag, "]")),
				Name:         tag,
				Dialer:       detour,
				Address:      server.Address,
				Addresses:    server.Addresses,//karing
				ClientSubnet: clientSubnet,
			})
			if err != nil {
				return nil, E.Cause(err, "parse dns server[", tag, "]")
			}
			transports[i] = transport
			dummyTransportMap[tag] = transport
			if server.Tag != "" {
				transportMap[server.Tag] = transport
			}
			strategy := dns.DomainStrategy(server.Strategy)
			if strategy != dns.DomainStrategyAsIS {
				transportDomainStrategy[transport] = strategy
			}
		}
		if len(transports) == len(dummyTransportMap) {
			break
		}
		if lastLen != len(dummyTransportMap) {
			continue
		}
		unresolvedTags := common.MapIndexed(common.FilterIndexed(dnsOptions.Servers, func(index int, server option.DNSServerOptions) bool {
			_, exists := dummyTransportMap[transportTags[index]]
			return !exists
		}), func(index int, server option.DNSServerOptions) string {
			return transportTags[index]
		})
		if len(unresolvedTags) == 0 {
			panic(F.ToString("unexpected unresolved dns servers: ", len(transports), " ", len(dummyTransportMap), " ", len(transportMap)))
		}
		return nil, E.New("found circular reference in dns servers: ", strings.Join(unresolvedTags, " "))
	}
	var defaultTransport dns.Transport
	if dnsOptions.Final != "" {
		defaultTransport = dummyTransportMap[dnsOptions.Final]
		if defaultTransport == nil {
			return nil, E.New("default dns server not found: ", dnsOptions.Final)
		}
	}
	if defaultTransport == nil {
		if len(transports) == 0 {
			transports = append(transports, common.Must1(dns.CreateTransport(dns.TransportOptions{
				Context: ctx,
				Name:    "local",
				Address: "local",
				Dialer:  common.Must1(dialer.NewDefault(ctx, option.DialerOptions{})),
			})))
		}
		defaultTransport = transports[0]
	}
	if _, isFakeIP := defaultTransport.(adapter.FakeIPTransport); isFakeIP {
		return nil, E.New("default DNS server cannot be fakeip")
	}
	router.defaultTransport = defaultTransport
	router.transports = transports
	router.transportMap = transportMap
	router.transportDomainStrategy = transportDomainStrategy

	if dnsOptions.ReverseMapping {
		router.dnsReverseMapping = NewDNSReverseMapping()
	}

	if fakeIPOptions := dnsOptions.FakeIP; fakeIPOptions != nil && dnsOptions.FakeIP.Enabled {
		var inet4Range netip.Prefix
		var inet6Range netip.Prefix
		if fakeIPOptions.Inet4Range != nil {
			inet4Range = *fakeIPOptions.Inet4Range
		}
		if fakeIPOptions.Inet6Range != nil {
			inet6Range = *fakeIPOptions.Inet6Range
		}
		router.fakeIPStore = fakeip.NewStore(ctx, router.logger, inet4Range, inet6Range)
	}
	return router, nil
}

func (r *Router) Start(stage adapter.StartStage) error {
	monitor := taskmonitor.New(r.logger, C.StartTimeout)
	switch stage {
	case adapter.StartStateInitialize:
		if r.fakeIPStore != nil {
			monitor.Start("initialize fakeip store")
			err := r.fakeIPStore.Start()
			monitor.Finish()
			if err != nil {
				return err
			}
		}
	case adapter.StartStateStart:
		/*if r.timeService != nil {// karing
			go func(){ // karing
				monitor := taskmonitor.New(r.logger, C.StartTimeout)
				monitor.Start("initialize time service")
				err := r.timeService.Start()
				monitor.Finish()
				if err != nil {
					time.Sleep(time.Second * 3)
					err := r.timeService.Start()
					if err != nil {
						r.logger.ErrorContext(r.ctx, "initialize time service: ", err)
					}
				}
			}()
		}*/
		if r.needGeoIPDatabase {
			monitor.Start("initialize geoip database")
			err := r.prepareGeoIPDatabase()
			monitor.Finish()
			if err != nil {
				return err
			}
		}
		if r.needGeositeDatabase {
			monitor.Start("initialize geosite database")
			err := r.prepareGeositeDatabase()
			monitor.Finish()
			if err != nil {
				return err
			}
		}
		if r.needGeositeDatabase {
			for _, rule := range r.rules {
				err := rule.UpdateGeosite()
				if err != nil {
					r.logger.Error("failed to initialize geosite: ", err)
				}
			}
			for _, rule := range r.dnsRules {
				err := rule.UpdateGeosite()
				if err != nil {
					r.logger.Error("failed to initialize geosite: ", err)
				}
			}
			err := common.Close(r.geositeReader)
			if err != nil {
				return err
			}
			r.geositeCache = nil
			r.geositeReader = nil
		}

		monitor.Start("initialize DNS client")
		r.dnsClient.Start()
		monitor.Finish()

		for _, rule := range r.dnsRules { //karing
			monitor.Start("initialize DNS rule[", rule, "]") //karing
			err := rule.Start()
			monitor.Finish()
			if err != nil {
				return E.Cause(err, "initialize DNS rule[", rule, "]") //karing
			}
		}
		for _, transport := range r.transports { //karing
			monitor.Start("initialize DNS transport[", transport.Name(), "]")  //karing
			err := transport.Start()
			monitor.Finish()
			if err != nil {
				return E.Cause(err, "initialize DNS server[", transport.Name(), "]")//karing
			}
		}
		var cacheContext *adapter.HTTPStartContext
		if len(r.ruleSets) > 0 {
			monitor.Start("initialize rule-set")
			cacheContext = adapter.NewHTTPStartContext()
			var ruleSetStartGroup task.Group
			for i, ruleSet := range r.ruleSets {
				ruleSetInPlace := ruleSet
				ruleSetStartGroup.Append0(func(ctx context.Context) error {
					err := ruleSetInPlace.StartContext(ctx, cacheContext)
					if err != nil {
						return E.Cause(err, "initialize rule-set[", i, "]")
					}
					return nil
				})
			}
			ruleSetStartGroup.Concurrency(5)
			ruleSetStartGroup.FastFail()
			err := ruleSetStartGroup.Run(r.ctx)
			monitor.Finish()
			if err != nil {
				return err
			}
		}
		if cacheContext != nil {
			cacheContext.Close()
		}
		needFindProcess := r.needFindProcess
		for _, ruleSet := range r.ruleSets {
			metadata := ruleSet.Metadata()
			if metadata.ContainsProcessRule {
				needFindProcess = true
			}
			if metadata.ContainsWIFIRule {
				r.needWIFIState = true
			}
		}
		if needFindProcess {
			if r.platformInterface != nil && !C.IsDarwin{ //karing
				r.processSearcher = r.platformInterface
			} else {
				monitor.Start("initialize process searcher")
				searcher, err := process.NewSearcher(process.Config{
					Logger:         r.logger,
					PackageManager: r.network.PackageManager(),
				})
				monitor.Finish()
				if err != nil {
					if err != os.ErrInvalid {
						r.logger.Warn(E.Cause(err, "create process searcher"))
					}
				} else {
					r.processSearcher = searcher
				}
			}
		}
	case adapter.StartStatePostStart:
		for i, rule := range r.rules {
			monitor.Start("initialize rule[", i, "]")
			err := rule.Start()
			monitor.Finish()
			if err != nil {
				return E.Cause(err, "initialize rule[", i, "]")
			}
		}
		for _, ruleSet := range r.ruleSets {
			monitor.Start("post start rule_set[", ruleSet.Name(), "]")
			err := ruleSet.PostStart()
			monitor.Finish()
			if err != nil {
				return E.Cause(err, "post start rule_set[", ruleSet.Name(), "]")
			}
		}
		r.started = true
		return nil
	case adapter.StartStateStarted:
		for _, ruleSet := range r.ruleSetMap {
			ruleSet.Cleanup()
		}
		runtime.GC()
	}
	return nil
}

func (r *Router) Close() error {
	monitor := taskmonitor.New(r.logger, C.StopTimeout)
	var err error
	for i, rule := range r.rules {
		monitor.Start("close rule[", i, "]")
		err = E.Append(err, rule.Close(), func(err error) error {
			return E.Cause(err, "close rule[", i, "]")
		})
		monitor.Finish()
	}
	for i, rule := range r.dnsRules {
		monitor.Start("close dns rule[", i, "]")
		err = E.Append(err, rule.Close(), func(err error) error {
			return E.Cause(err, "close dns rule[", i, "]")
		})
		monitor.Finish()
	}
	for i, transport := range r.transports {
		monitor.Start("close dns transport[", i, "]")
		err = E.Append(err, transport.Close(), func(err error) error {
			return E.Cause(err, "close dns transport[", i, "]")
		})
		monitor.Finish()
	}
	if r.geoIPReader != nil {
		monitor.Start("close geoip reader")
		err = E.Append(err, r.geoIPReader.Close(), func(err error) error {
			return E.Cause(err, "close geoip reader")
		})
		monitor.Finish()
	}
	if r.fakeIPStore != nil {
		monitor.Start("close fakeip store")
		err = E.Append(err, r.fakeIPStore.Close(), func(err error) error {
			return E.Cause(err, "close fakeip store")
		})
		monitor.Finish()
	}
	//karing
	r.processSearcher = nil
	r.fakeIPStore = nil
	r.rules = make([]adapter.Rule, 0)
	r.geositeCache = make(map[string]adapter.Rule)
	r.staticDns = make(map[string]StaticDNSEntry)
	r.defaultDomainStrategy = dns.DomainStrategyAsIS 
	r.dnsRules = make([]adapter.DNSRule, 0)
	r.ruleSetsRemoteWithLocal = make([]adapter.RuleSet, 0)
	for _, ruleSet := range r.ruleSets {
		ruleSet.Close()
	}
	r.ruleSets = make([]adapter.RuleSet, 0)
	r.ruleSetMap = make(map[string]adapter.RuleSet)
	r.transports = make([]dns.Transport, 0)
	r.transportMap = make(map[string]dns.Transport)
	r.transportDomainStrategy = make(map[dns.Transport]dns.DomainStrategy)
	//karing

	return err
}

func (r *Router) PostStart() error {
	monitor := taskmonitor.New(r.logger, C.StopTimeout)
	if len(r.ruleSets) > 0 {
		monitor.Start("initialize rule-set")
		ruleSetStartContext := NewRuleSetStartContext()
		var ruleSetStartGroup task.Group
		for i, ruleSet := range r.ruleSets {
			ruleSetInPlace := ruleSet
			ruleSetStartGroup.Append0(func(ctx context.Context) error {
				err := ruleSetInPlace.StartContext(ctx, ruleSetStartContext)
				if err != nil {
					return E.Cause(err, "initialize rule-set[", i, "]")
				}
				return nil
			})
		}
		ruleSetStartGroup.Concurrency(5)
		ruleSetStartGroup.FastFail()
		err := ruleSetStartGroup.Run(r.ctx)
		monitor.Finish()
		if err != nil {
			return err
		}
		ruleSetStartContext.Close()
	}

	if len(r.ruleSetsRemoteWithLocal) > 0 { //karing
		go func() {
			ruleSetStartContext := NewRuleSetStartContext()
			var ruleSetStartGroup task.Group
			for i, ruleSet := range r.ruleSetsRemoteWithLocal {
				ruleSetInPlace := ruleSet
				ruleSetStartGroup.Append0(func(ctx context.Context) error {
					err := ruleSetInPlace.StartContext(ctx, ruleSetStartContext)
					if err != nil {
						return E.Cause(err, "initialize rule-set-remote[", i, "]")
					}
					return nil
				})
			}
			ruleSetStartGroup.Concurrency(5)
			ruleSetStartGroup.FastFail()
			ruleSetStartGroup.Run(r.ctx)

			ruleSetStartContext.Close()
		}()
	}
 
	for _, rule := range r.rules { //karing
		monitor.Start("initialize rule[", rule, "]") //karing
		err := rule.Start()
		monitor.Finish()
		if err != nil {
			return E.Cause(err, "initialize rule[", rule, "]") //karing
		}
	}
	r.started = true
	return nil
}

func (r *Router) FakeIPStore() adapter.FakeIPStore {
	return r.fakeIPStore
}

func (r *Router) RuleSet(tag string) (adapter.RuleSet, bool) {
	ruleSet, loaded := r.ruleSetMap[tag]
	return ruleSet, loaded
}

func (r *Router) GetRemoteRuleSetRulesCount() map[string]int{  //karing
	counts :=  make( map[string]int)
	for _, ruleSet := range r.ruleSets {
		if ruleset, isRemote := ruleSet.(*R.RemoteRuleSet); isRemote {
			counts[ruleset.options.RemoteOptions.URL]= len(ruleset.rules)
		}
	}
	return counts
}

func (r *Router) NeedWIFIState() bool {
	return r.needWIFIState
}

func (r *Router) Rules() []adapter.Rule {
	return r.rules
}

func (r *Router) SetTracker(tracker adapter.ConnectionTracker) {
	r.tracker = tracker
}

func (r *Router) ResetNetwork() {
	r.network.ResetNetwork()
	for _, transport := range r.transports {
		transport.Reset()
	}
}
func (r *Router) FindProcessInfo(ctx context.Context, network string, source netip.AddrPort)(*process.Info, error){ //karing
	if r.processSearcher != nil {
		var originDestination netip.AddrPort
		return  process.FindProcessInfo(r.processSearcher, ctx, network, source, originDestination)
	}
	return nil, E.New("processSearcher not impl")
}
func (r *Router) GetMatchRuleChain(rule adapter.Rule) []string { //karing
	var chain []string
	var next string
	if rule == nil {
		if defaultOutbound, err := r.DefaultOutbound(N.NetworkTCP); err == nil {
			next = defaultOutbound.Tag()
		}
	} else {
		next = rule.Outbound()
	}
	for {
		chain = append(chain, next)
		detour, loaded := r.Outbound(next)
		if !loaded {
			break
		}
		group, isGroup := detour.(adapter.OutboundGroup)
		if !isGroup {
			break
		}
		next = group.Now()
	}
	return chain
}
func (r *Router) GetMatchRule(ctx context.Context, metadata *adapter.InboundContext) (adapter.Rule, string, error) { //karing
	_, rule, detour, err := r.match(ctx, metadata, r.defaultOutboundForConnection)
	if err != nil {
		return nil, "", err
	}

	return rule, detour.Tag(), err
}
func (r *Router) GetAssetContent(path string)([]byte, error) {//karing
	if r.platformInterface == nil{
		return nil, E.New("platform interface not set")
	}
	return r.platformInterface.GetAssetContent(path)
}
func (r *Router) SingalQuit(){ //karing
	r.quitSig()
}
