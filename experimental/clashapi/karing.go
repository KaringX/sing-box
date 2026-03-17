package clashapi

//karing
import (
	"context"
	"io"
	"net/http"
	"net/netip"
	"runtime"
	runtimeDebug "runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sagernet/sing-box/adapter"
	D "github.com/sagernet/sing-box/common/debug"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/dns"
	"github.com/sagernet/sing-box/dns/transport/local"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/json"
	M "github.com/sagernet/sing/common/metadata"
	"github.com/sagernet/sing/service"
	//"github.com/sagernet/sing-box/log"
)

type DNSQueryRequest struct {
	Servers  []option.DNSServerOptions `json:"servers,omitempty"`
	Tag      string                    `json:"tag"`
	Domain   string                    `json:"domain"`
	Strategy option.DomainStrategy     `json:"strategy"`
}

var (
	dnsClient *dns.Client
)

func init() {
	dnsClient = dns.NewClient(dns.ClientOptions{
		DisableCache:     true,
		DisableExpire:    false,
		IndependentCache: true,
		//Logger:           router.dns,
	})
}

func LookupWithDefaultRouter(ctx context.Context, logFactory log.Factory, domain string, strategy C.DomainStrategy) (uint16, []netip.Addr, string, error) {
	ctx, cancel := context.WithTimeout(ctx, C.DNSTimeout)
	defer cancel()
	dnsRouter := service.FromContext[adapter.DNSRouter](ctx)
	start := time.Now()
	addr, tag, err := dnsRouter.LookupTag(ctx, domain, adapter.DNSQueryOptions{
		Strategy: strategy,
	})
	if err != nil {
		return 0, nil, tag, err
	}
	duration := uint16(time.Since(start) / time.Millisecond)
	return duration, addr, tag, nil
}

func Lookup(ctx context.Context, router adapter.Router, logFactory log.Factory, req DNSQueryRequest) (uint16, []netip.Addr, error) {
	//var dnsClient = router.GetDNSClient()
	ctxClone := service.Clone(ctx)
	defer service.UnRegisterAll(ctxClone)
	ctxClone, cancel := context.WithTimeout(ctxClone, C.DNSTimeout)
	defer cancel()
	outboundManager := service.FromContext[adapter.OutboundManager](ctxClone)
	dnsTransportRegistry := service.FromContext[adapter.DNSTransportRegistry](ctxClone)
	dnsTransportManager := dns.NewTransportManager(logFactory.NewLogger("dns_Lookup/transport"), dnsTransportRegistry, outboundManager, "")
	service.MustRegister[adapter.DNSTransportManager](ctxClone, dnsTransportManager)
	defer dnsTransportManager.Close()
	err := dnsTransportManager.Start(adapter.StartStateInitialize)
	if err != nil {
		return 0, nil, E.Cause(err, "dnsTransportManager start failed")
	}
	for i, transportOptions := range req.Servers {
		var tag string
		if transportOptions.Tag != "" {
			tag = transportOptions.Tag
		} else {
			tag = F.ToString(i)
		}
		err := dnsTransportManager.Create(
			ctxClone,
			logFactory.NewLogger(F.ToString("dns_Lookup/", transportOptions.Type, "[", tag, "]")),
			tag,
			transportOptions.Type,
			transportOptions.Options,
		)
		if err != nil {
			return 0, nil, E.Cause(err, "initialize DNS server[", i, "]")
		}
	}
	dnsTransportManager.Initialize(common.Must1(
		local.NewTransport(
			ctxClone,
			logFactory.NewLogger("dns_Lookup/local"),
			"local",
			option.LocalDNSServerOptions{},
		)))
	transport, ok := dnsTransportManager.Transport(req.Tag)
	if !ok {
		return 0, nil, E.New("server tag[", req.Tag, "] not found")
	}
	start := time.Now()
	addr, err := dnsClient.Lookup(ctxClone, transport, req.Domain, adapter.DNSQueryOptions{
		Strategy: C.DomainStrategy(req.Strategy),
	}, nil)
	if err != nil {
		return 0, nil, err
	}
	duration := uint16(time.Since(start) / time.Millisecond)
	return duration, addr, nil
}

func karingRouter(ctx context.Context, router adapter.Router, logFactory log.Factory) http.Handler {
	r := chi.NewRouter()
	r.Get("/dnsQueryWithDefaultRouter", dnsQueryWithDefaultRouter(ctx, logFactory))
	r.Post("/dnsQuery", dnsQuery(ctx, router, logFactory))
	r.Get("/outboundQuery", outboundQuery(ctx, router))
	r.Get("/remoteRuleSetRulesCount", remoteRuleSetRulesCount(router))
	r.Get("/remoteRuleSetStates", remoteRuleSetRulesStates(ctx))
	r.Get("/resetOutboundConnections", resetOutboundConnections())
	r.Get("/mainStack", mainStack())
	return r
}

func dnsQueryWithDefaultRouter(ctx context.Context, logFactory log.Factory) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		domain := r.URL.Query().Get("domain")
		strategy := r.URL.Query().Get("strategy")
		var domainStrategy option.DomainStrategy
		switch strategy {
		case "", "as_is":
			domainStrategy = option.DomainStrategy(C.DomainStrategyAsIS)
		case "prefer_ipv4":
			domainStrategy = option.DomainStrategy(C.DomainStrategyPreferIPv4)
		case "prefer_ipv6":
			domainStrategy = option.DomainStrategy(C.DomainStrategyPreferIPv6)
		case "ipv4_only":
			domainStrategy = option.DomainStrategy(C.DomainStrategyIPv4Only)
		case "ipv6_only":
			domainStrategy = option.DomainStrategy(C.DomainStrategyIPv6Only)
		default:
			render.JSON(w, r, render.M{
				"err":     E.New("unknown domain strategy: ", domainStrategy).Error(),
				"latency": nil,
				"addr":    nil,
				"tag":     "",
			})
			return
		}

		duration, addr, tag, err := LookupWithDefaultRouter(ctx, logFactory, domain, C.DomainStrategy(domainStrategy))
		if err != nil {
			render.JSON(w, r, render.M{
				"err":     err.Error(),
				"latency": nil,
				"addr":    nil,
				"tag":     tag,
			})
		} else {
			render.JSON(w, r, render.M{
				"err":     nil,
				"latency": duration,
				"addr":    addr,
				"tag":     tag,
			})
		}
	}
}

func dnsQuery(ctx context.Context, router adapter.Router, logFactory log.Factory) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		rawMessage, err := io.ReadAll(r.Body)
		if err != nil {
			render.JSON(w, r, render.M{
				"err": err.Error(),
			})
			return
		}
		req, err := json.UnmarshalExtendedContext[DNSQueryRequest](ctx, rawMessage)
		if err != nil {
			render.JSON(w, r, render.M{
				"err": err.Error(),
			})
			return
		}

		duration, addr, err := Lookup(ctx, router, logFactory, req)
		if err != nil {
			render.JSON(w, r, render.M{
				"err":     err.Error(),
				"latency": nil,
				"addr":    nil,
			})
		} else {
			render.JSON(w, r, render.M{
				"err":     nil,
				"latency": duration,
				"addr":    addr,
			})
		}
	}
}

func outboundQuery(ctx context.Context, router adapter.Router) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		domain := r.URL.Query().Get("domain")
		ip := r.URL.Query().Get("ip")
		meta := adapter.InboundContext{Domain: domain, Destination: M.ParseSocksaddr(ip)}
		rule, err := router.GetMatchRule(ctx, &meta)

		if err != nil {
			render.JSON(w, r, render.M{
				"err":         err.Error(),
				"rule":        nil,
				"chain":       nil,
				"action_type": nil,
				"outbound":    nil,
			})
		} else {
			outboundManager := service.FromContext[adapter.OutboundManager](ctx)
			var ruleName string
			var chain []string
			var outbound string
			var actionType string
			if rule != nil {
				ruleName = rule.String()
				actionType = rule.Action().Type()
				outbound = rule.Action().Target()
			} else {
				ruleName = "final"
			}
			if len(actionType) == 0 || actionType == C.RuleActionTypeRoute {
				chain, outbound, _ = router.GetMatchRuleChain(outboundManager, outbound)
			}
			render.JSON(w, r, render.M{
				"err":         nil,
				"rule":        ruleName,
				"chain":       chain,
				"action_type": actionType,
				"outbound":    outbound,
			})
		}
	}
}

func remoteRuleSetRulesCount(router adapter.Router) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, render.M{
			"result": router.GetRemoteRuleSetRulesCount(),
		})
	}
}

func remoteRuleSetRulesStates(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var cached map[string]time.Time
		var failed map[string]string
		cacheFile := service.FromContext[adapter.CacheFile](ctx)
		if cacheFile != nil {
			cached = cacheFile.GetAllRuleSetCachedLastUpdated()
			failed = cacheFile.GetAllRuleSetFailed()
		}
		render.JSON(w, r, render.M{
			"cached": cached,
			"failed": failed,
		})
	}
}

func resetOutboundConnections() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conntrack.Close()
		go func() {
			runtime.GC()
			runtimeDebug.FreeOSMemory()
		}()
		render.JSON(w, r, render.M{})
	}
}

func mainStack() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		stack := D.GetGoroutineStack(D.MainGoroutineId)
		render.JSON(w, r, render.M{
			"mainGoId": D.MainGoroutineId,
			"result":   stack,
		})
	}
}
