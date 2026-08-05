package route

import (
	"context"

	"os"
	"runtime"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/process"
	"github.com/sagernet/sing-box/common/taskmonitor"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	R "github.com/sagernet/sing-box/route/rule"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/task"
	"github.com/sagernet/sing/contrab/freelru"
	"github.com/sagernet/sing/contrab/maphash"
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
	ruleSets                []adapter.RuleSet
	ruleSetsRemoteWithLocal []adapter.RuleSet //karing
	ruleSetMap              map[string]adapter.RuleSet
	processSearcher         process.Searcher
	processCache            freelru.Cache[processCacheKey, processCacheEntry]
	pauseManager            pause.Manager
	trackers                []adapter.ConnectionTracker
	platformInterface       adapter.PlatformInterface
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
		platformInterface: service.FromContext[adapter.PlatformInterface](ctx),
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
		r.network.Initialize(r.ruleSets)
		needFindProcess := r.needFindProcess
		for _, ruleSet := range r.ruleSets {
			metadata := ruleSet.Metadata()
			if metadata.ContainsProcessRule {
				needFindProcess = true
			}
		}
		if C.IsAndroid && r.platformInterface != nil {
			needFindProcess = true
		}
		r.needFindProcess = needFindProcess
		if needFindProcess && !C.IsIos { //karing
			if r.platformInterface != nil && !C.IsDarwin && r.platformInterface.UsePlatformConnectionOwnerFinder() { //karing
				r.processSearcher = newPlatformSearcher(r.platformInterface)
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
		if r.processSearcher != nil {
			processCache := common.Must1(freelru.NewSharded[processCacheKey, processCacheEntry](256, maphash.NewHasher[processCacheKey]().Hash32))
			cacheLifetime := 200 * time.Millisecond //karing
			if C.IsWindows {                        //karing
				cacheLifetime = 1 * time.Second
			}
			processCache.SetLifetime(cacheLifetime)
			r.processCache = processCache
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
	if r.processSearcher != nil {
		monitor.Start("close process searcher")
		err = E.Append(err, r.processSearcher.Close(), func(err error) error {
			return E.Cause(err, "close process searcher")
		})
		monitor.Finish()
	}
	return err
}

func (r *Router) RuleSet(tag string) (adapter.RuleSet, bool) {
	ruleSet, loaded := r.ruleSetMap[tag]
	return ruleSet, loaded
}

func (r *Router) Rules() []adapter.Rule {
	return r.rules
}

func (r *Router) AppendTracker(tracker adapter.ConnectionTracker) {
	r.trackers = append(r.trackers, tracker)
}

func (r *Router) NeedFindProcess() bool {
	return r.needFindProcess
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
