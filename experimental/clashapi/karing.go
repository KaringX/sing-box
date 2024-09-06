package clashapi

//karing
import (
	"context"
	"net/http"
	"net/netip"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/log"
	dns "github.com/sagernet/sing-dns"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

var (
	dnsClient *dns.Client
)

type DNSServer struct {
	Tag       string `json:"tag"`
	Address   string `json:"address"`
	Addresses []string `json:"addresses"`
	Strategy  string `json:"strategy"`
	Detour    string `json:"detour"`
}
type DNSQueryRequest struct {
	Resolver DNSServer `json:"resolver"`
	Query    DNSServer `json:"query"`
	Domain   string    `json:"domain"`
}

func transStrategy(strategy string) dns.DomainStrategy {
	switch strategy {
	case "", "as_is":
		return dns.DomainStrategy(dns.DomainStrategyAsIS)
	case "prefer_ipv4":
		return dns.DomainStrategy(dns.DomainStrategyPreferIPv4)
	case "prefer_ipv6":
		return dns.DomainStrategy(dns.DomainStrategyPreferIPv6)
	case "ipv4_only":
		return dns.DomainStrategy(dns.DomainStrategyUseIPv4)
	case "ipv6_only":
		return dns.DomainStrategy(dns.DomainStrategyUseIPv6)
	default:
		return dns.DomainStrategy(dns.DomainStrategyPreferIPv4)
	}
}
func LookupWithDefaultRouter(router adapter.Router, logFactory log.Factory, domain string, strategy dns.DomainStrategy) (uint16, []netip.Addr, string, error) {
	ctx := context.Background()
	ctx = adapter.ContextWithRouter(ctx, router)
	start := time.Now()
	addr, tag, err := router.LookupTag(ctx, domain, strategy)
	if err != nil {
		return 0, nil, tag, err
	}
	duration := uint16(time.Since(start) / time.Millisecond)
	return duration, addr, tag, nil
}
func Lookup(router adapter.Router, logFactory log.Factory, req DNSQueryRequest) (uint16, []netip.Addr, error) {
	ctx := context.Background()
	ctx = adapter.ContextWithRouter(ctx, router)
	var resolverTransport dns.Transport

	if len(req.Resolver.Addresses) != 0 {
		tag := req.Resolver.Tag + "_" + req.Resolver.Detour
		var detour N.Dialer
		if req.Resolver.Detour == "" {
			detour = dialer.NewRouter(router)
		} else {
			_, detourExist := router.Outbound(req.Resolver.Detour)
			if !detourExist {
				return 0, nil, E.New("resolver.detour not found: " + req.Resolver.Detour)
			}
			detour = dialer.NewDetour(router, req.Resolver.Detour)
		}

		transport, err := dns.CreateTransport(dns.TransportOptions{
			Context: ctx,
			Logger:  logFactory.NewLogger(F.ToString("dns_query_resolver/transport[", tag, "]")),
			Name:    req.Resolver.Tag,
			Dialer:  detour,
			Address:  req.Resolver.Address,
			Addresses: req.Resolver.Addresses,
		})
		if err != nil {
			return 0, nil, err
		}
		resolverTransport = transport
	}

	tag := req.Query.Tag + "_" + req.Query.Detour
	var detour N.Dialer
	if req.Query.Detour == "" {
		detour = dialer.NewRouter(router)
	} else {
		_, detourExist := router.Outbound(req.Query.Detour)
		if !detourExist {
			return 0, nil, E.New("query.detour not found: " + req.Query.Detour)
		}
		detour = dialer.NewDetour(router, req.Query.Detour)
	}
	if len(req.Resolver.Addresses) != 0 {
		detour = dns.NewDialerWrapper(detour, dnsClient, resolverTransport, transStrategy(req.Query.Strategy), time.Duration(0))
	}

	transport, err := dns.CreateTransport(dns.TransportOptions{
		Context: ctx,
		Logger:  logFactory.NewLogger(F.ToString("dns_query/transport[", tag, "]")),
		Name:    req.Query.Tag,
		Dialer:  detour,
		Address:  req.Query.Address,
		Addresses: req.Query.Addresses,
	})

	if err != nil {
		return 0, nil, err
	}

	start := time.Now()
	addr, err := dnsClient.Lookup(ctx, transport, req.Domain, transStrategy(req.Query.Strategy))
	if err != nil {
		return 0, nil, err
	}
	duration := uint16(time.Since(start) / time.Millisecond)
	return duration, addr, nil
}

func karingRouter(router adapter.Router, logFactory log.Factory) http.Handler {
	dnsClient = dns.NewClient(dns.ClientOptions{
		DisableCache:     true,
		DisableExpire:    false,
		IndependentCache: true,
		//Logger:           router.dns,
	})

	r := chi.NewRouter()
	r.Get("/stop", stop(router, logFactory))
	r.Get("/dnsQueryWithDefaultRouter", dnsQueryWithDefaultRouter(router, logFactory))
	r.Post("/dnsQuery", dnsQuery(router, logFactory))
	r.Get("/outboundQuery", outboundQuery(router, logFactory))
	r.Get("/directDelay", directDelay(router, logFactory))
	r.Get("/remoteRuleSetRulesCount", remoteRuleSetRulesCount(router, logFactory))
	return r
}

func stop(router adapter.Router, logFactory log.Factory) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
	    router.SingalQuit()
		render.JSON(w, r, render.M{
			"pid": os.Getpid(),
		})
	}
}

func dnsQueryWithDefaultRouter(router adapter.Router, logFactory log.Factory) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		domain := r.URL.Query().Get("domain")
		strategy := r.URL.Query().Get("strategy")

		duration, addr, tag, err := LookupWithDefaultRouter(router, logFactory, domain, transStrategy(strategy))
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

func dnsQuery(router adapter.Router, logFactory log.Factory) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := DNSQueryRequest{}
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			render.JSON(w, r, render.M{
				"err": "invalid json data",
			})
			return
		}
		duration, addr, err := Lookup(router, logFactory, req)
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

func outboundQuery(router adapter.Router, logFactory log.Factory) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		domain := r.URL.Query().Get("domain")
		ip := r.URL.Query().Get("ip")
		ctx := context.Background()
		ctx = adapter.ContextWithRouter(ctx, router)
		meta := adapter.InboundContext{Domain: domain, Destination: M.ParseSocksaddr(ip)}
		rule, outbound, err := router.GetMatchRule(ctx, &meta)
		rulechain := router.GetMatchRuleChain(rule)
		if err != nil {
			render.JSON(w, r, render.M{
				"err":       err.Error(),
				"rule":      nil,
				"rulechain": nil,
				"outbound":  nil,
			})
		} else {
			if rule != nil {
				render.JSON(w, r, render.M{
					"err":       nil,
					"rule":      rule.String(),
					"rulechain": rulechain,
					"outbound":  outbound,
				})
			} else {
				render.JSON(w, r, render.M{
					"err":       nil,
					"rule":      "final",
					"rulechain": rulechain,
					"outbound":  outbound,
				})
			}
		}
	}
}
func directDelay(router adapter.Router, logFactory log.Factory) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		if len(url) == 0 {
			render.JSON(w, r, render.M{
				"err": "invalid url",
			})
			return
		}

		delay, err := httpDirectDelay(url)
		if err != nil {
			render.JSON(w, r, render.M{
				"err": err.Error(),
			})
			return
		}
		render.JSON(w, r, render.M{
			"delay": delay,
		})
	}
}
func remoteRuleSetRulesCount(router adapter.Router, logFactory log.Factory) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, render.M{
			"result": router.GetRemoteRuleSetRulesCount(),
		})
	}
}
func httpDirectDelay(url string) (uint16, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	start := time.Now()
	req, _ := http.NewRequest("HEAD", url, nil)
	response, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return 0, err
	}

	response.Body.Close()
	delay := uint16(time.Since(start).Milliseconds())
	return delay, nil
}