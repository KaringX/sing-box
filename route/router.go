package route

import (
	"context"
	"net/netip"

	"os"
	"runtime"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/process"
	"github.com/sagernet/sing-box/common/taskmonitor"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/experimental/clashapi/trafficontrol"
	"github.com/sagernet/sing-box/experimental/libbox/platform"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	R "github.com/sagernet/sing-box/route/rule"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/task"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
)

var _ adapter.Router = (*Router)(nil)

type Router struct {
	ctx                     context.Context
	logger                  log.ContextLogger
	inbound                 adapter.InboundManager
	outbound                adapter.OutboundManager
	dns                     adapter.DNSRouter
	dnsTransport            adapter.DNSTransportManager
	connection              adapter.ConnectionManager
	network                 adapter.NetworkManager
	rules                   []adapter.Rule
	needFindProcess         bool
	ruleSetsRemoteWithLocal []adapter.RuleSet //karing
	ruleSets                []adapter.RuleSet
	ruleSetMap              map[string]adapter.RuleSet
	processSearcher         process.Searcher
	pauseManager            pause.Manager
	trackers                []adapter.ConnectionTracker
	platformInterface       platform.Interface
	needWIFIState           bool
	started                 bool
}

func NewRouter(ctx context.Context, logFactory log.Factory, options option.RouteOptions, dnsOptions option.DNSOptions) *Router {
	return &Router{
		ctx:               ctx,
		logger:            logFactory.NewLogger("router"),
		inbound:           service.FromContext[adapter.InboundManager](ctx),
		outbound:          service.FromContext[adapter.OutboundManager](ctx),
		dns:               service.FromContext[adapter.DNSRouter](ctx),
		dnsTransport:      service.FromContext[adapter.DNSTransportManager](ctx),
		connection:        service.FromContext[adapter.ConnectionManager](ctx),
		network:           service.FromContext[adapter.NetworkManager](ctx),
		rules:             make([]adapter.Rule, 0, len(options.Rules)),
		ruleSetMap:        make(map[string]adapter.RuleSet),
		needFindProcess:   hasRule(options.Rules, isProcessRule) || hasDNSRule(dnsOptions.Rules, isProcessDNSRule) || options.FindProcess,
		pauseManager:      service.FromContext[pause.Manager](ctx),
		platformInterface: service.FromContext[platform.Interface](ctx),
		needWIFIState:     hasRule(options.Rules, isWIFIRule) || hasDNSRule(dnsOptions.Rules, isWIFIDNSRule),
	}
}

func (r *Router) Initialize(rules []option.Rule, ruleSets []option.RuleSet) error {
	for i, options := range rules {
		rule, err := R.NewRule(r.ctx, r.logger, options, false)
		if err != nil {
			return E.Cause(err, "parse rule[", i, "]")
		}
		r.rules = append(r.rules, rule)
	}
	for i, options := range ruleSets {
		if _, exists := r.ruleSetMap[options.Tag]; exists {
			return E.New("duplicate rule-set tag: ", options.Tag)
		}
		if options.Type == C.RuleSetTypeRemote { //karing
			if len(options.RemoteOptions.Path) != 0 {
				cacheFile := service.FromContext[adapter.CacheFile](r.ctx)
				if cacheFile != nil {
					if !cacheFile.HasRuleSet(options.RemoteOptions.URL) {
						ruleSet := R.NewRemoteRuleSet(r.ctx, r.logger, options)
						r.ruleSetsRemoteWithLocal = append(r.ruleSetsRemoteWithLocal, ruleSet)

						options.Type = C.RuleSetTypeLocal
						options.LocalOptions.Path = options.RemoteOptions.Path
						options.LocalOptions.IsAsset = options.RemoteOptions.IsAsset
					}
				}
			}
		}
		ruleSet, err := R.NewRuleSet(r.ctx, r.logger, options)
		if err != nil {
			return E.Cause(err, "parse rule-set[", i, "]")
		}
		r.ruleSets = append(r.ruleSets, ruleSet)
		r.ruleSetMap[options.Tag] = ruleSet
	}
	return nil
}

func (r *Router) Start(stage adapter.StartStage) error {
	monitor := taskmonitor.New(r.logger, C.StartTimeout)
	switch stage {
	case adapter.StartStateStart:
		var cacheContext *adapter.HTTPStartContext
		if len(r.ruleSets) > 0 {
			monitor.Start("initialize rule-set")
			cacheContext = adapter.NewHTTPStartContext(r.ctx)
			var ruleSetStartGroup task.Group
			for _, ruleSet := range r.ruleSets { //karing
				ruleSetInPlace := ruleSet
				ruleSetStartGroup.Append0(func(ctx context.Context) error {
					err := ruleSetInPlace.StartContext(ctx, cacheContext)
					if err != nil {
						return E.Cause(err, "initialize rule-set[", ruleSet.Name(), "]") //karing
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
		if len(r.ruleSetsRemoteWithLocal) > 0 { //karing
			cacheRemoteContext := adapter.NewHTTPStartContext(r.ctx)
			var ruleSetStartGroup task.Group
			for _, ruleSet := range r.ruleSetsRemoteWithLocal { //karing
				ruleSetInPlace := ruleSet
				ruleSetStartGroup.Append0(func(ctx context.Context) error {
					err := ruleSetInPlace.StartContext(ctx, cacheRemoteContext)
					if err != nil {
						return E.Cause(err, "initialize rule-set-remote[", ruleSet.Name(), "]") //karing
					}
					return nil
				})
			}
			ruleSetStartGroup.Concurrency(5)
			ruleSetStartGroup.FastFail()
			ruleSetStartGroup.Run(r.ctx)

			cacheRemoteContext.Close()
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
		if needFindProcess && !C.IsIos { //karing
			if r.platformInterface != nil && !C.IsDarwin { //karing
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
		for _, rule := range r.rules { //karing
			monitor.Start("initialize rule[", rule.Name(), "]") //karing
			err := rule.Start()
			monitor.Finish()
			if err != nil {
				return E.Cause(err, "initialize rule[", rule.Name(), "]") //karing
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
		for _, ruleSet := range r.ruleSetsRemoteWithLocal { //karing
			monitor.Start("post start rule_set_remote_with_local[", ruleSet.Name(), "]")
			err := ruleSet.PostStart()
			monitor.Finish()
			if err != nil {
				return E.Cause(err, "post start rule_set_remote_with_local[", ruleSet.Name(), "]")
			}
		}
		r.started = true
		return nil
	case adapter.StartStateStarted:
		for _, ruleSet := range r.ruleSets {
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
	for i, ruleSet := range r.ruleSets {
		monitor.Start("close rule-set[", i, "]")
		err = E.Append(err, ruleSet.Close(), func(err error) error {
			return E.Cause(err, "close rule-set[", i, "]")
		})
		monitor.Finish()
	}
	r.inbound = nil                                        //karing
	r.outbound = nil                                       //karing
	r.connection = nil                                     //karing
	r.network = nil                                        //karing
	r.rules = make([]adapter.Rule, 0)                      //karing
	r.ruleSetsRemoteWithLocal = make([]adapter.RuleSet, 0) //karing
	r.ruleSets = make([]adapter.RuleSet, 0)                //karing
	r.ruleSetMap = make(map[string]adapter.RuleSet)        //karing
	r.processSearcher = nil                                //karing
	r.pauseManager = nil                                   //karing
	r.platformInterface = nil                              //karing

	return err
}

func (r *Router) RuleSet(tag string) (adapter.RuleSet, bool) {
	ruleSet, loaded := r.ruleSetMap[tag]
	return ruleSet, loaded
}

func (r *Router) NeedWIFIState() bool {
	return r.needWIFIState
}

func (r *Router) Rules() []adapter.Rule {
	return r.rules
}

func (r *Router) AppendTracker(tracker adapter.ConnectionTracker) {
	r.trackers = append(r.trackers, tracker)
}

func (r *Router) ResetNetwork() {
	//r.network.ResetNetwork() //karing
	//r.dns.ResetNetwork() //karing

	if r.network != nil { //karing
		r.network.ResetNetwork()
	}
	if r.dns != nil { //karing
		r.dns.ResetNetwork()
	}
}

func (r *Router) ResetOutboundNetwork(tags []string) { //karing
	if r.network != nil {
		r.network.ResetOutboundNetwork(tags)
	}
}

func (r *Router) GetRemoteRuleSetRulesCount() map[string]int { //karing
	counts := make(map[string]int)
	for _, ruleSet := range r.ruleSets {
		if ruleset, isRemote := ruleSet.(*R.RemoteRuleSet); isRemote {
			counts[ruleset.Url()] = ruleset.RulesCount()
		}
	}
	return counts
}

func (r *Router) FindProcessInfo(ctx context.Context, network string, source netip.AddrPort) (*process.Info, error) { //karing
	if r.processSearcher != nil {
		var originDestination netip.AddrPort
		return process.FindProcessInfo(r.processSearcher, ctx, network, source, originDestination)
	}
	return nil, E.New("processSearcher not impl")
}

func (r *Router) GetMatchRuleChain(outboundManager adapter.OutboundManager, matchOutboundTag string) ([]string, string, string) { //karing
	return trafficontrol.GetMatchRuleChain(outboundManager, matchOutboundTag)
}

func (r *Router) GetMatchRule(ctx context.Context, metadata *adapter.InboundContext) (adapter.Rule, error) { //karing
	rule, _, _, _, err := r.matchRule(ctx, metadata, false, nil, nil)
	if err != nil {
		return nil, err
	}

	return rule, err
}

func (r *Router) GetAssetContent(path string) ([]byte, error) { //karing
	if r.platformInterface == nil {
		return nil, E.New("platform interface not set")
	}
	return r.platformInterface.GetAssetContent(path)
}
