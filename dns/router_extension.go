// karing
package dns

import (
	"context"
	"errors"
	"net/netip"
	"strings"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	R "github.com/sagernet/sing-box/route/rule"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	M "github.com/sagernet/sing/common/metadata"

	mDNS "github.com/miekg/dns"
)

func (r *Router) LookupTag(ctx context.Context, domain string, options adapter.DNSQueryOptions) ([]netip.Addr, string, error) {
	if r.transport == nil { //karing
		return nil, "", E.New("router closed")
	}
	var (
		responseAddrs []netip.Addr
		err           error
		transportTag  string //karing
	)
	printResult := func() {
		if err == nil && len(responseAddrs) == 0 {
			err = E.New("empty result")
		}
		if err != nil {
			if errors.Is(err, ErrResponseRejectedCached) {
				r.logger.DebugContext(ctx, "response rejected for ", domain, " (cached)")
			} else if errors.Is(err, ErrResponseRejected) {
				r.logger.DebugContext(ctx, "response rejected for ", domain)
			} else {
				r.logger.ErrorContext(ctx, E.Cause(err, "lookup failed for ", domain))
			}
		}
		if err != nil {
			err = E.Cause(err, "lookup ", domain)
		}
	}
	r.logger.DebugContext(ctx, "lookup domain ", domain)
	ctx, metadata := adapter.ExtendContext(ctx)
	metadata.Destination = M.Socksaddr{}
	metadata.Domain = FqdnToDomain(domain)
	if options.Transport != nil {
		transport := options.Transport
		transportTag = transport.Tag() //karing
		if legacyTransport, isLegacy := transport.(adapter.LegacyDNSTransport); isLegacy {
			if options.Strategy == C.DomainStrategyAsIS {
				options.Strategy = r.defaultDomainStrategy
			}
			if !options.ClientSubnet.IsValid() {
				options.ClientSubnet = legacyTransport.LegacyClientSubnet()
			}
		}
		if options.Strategy == C.DomainStrategyAsIS {
			options.Strategy = r.defaultDomainStrategy
		}
		responseAddrs, err = r.client.Lookup(ctx, transport, domain, options, nil)
	} else {
		var (
			transport adapter.DNSTransport
			rule      adapter.DNSRule
			ruleIndex int
		)
		ruleIndex = -1
		for {
			dnsCtx := adapter.OverrideContext(ctx)
			dnsOptions := options
			transport, rule, ruleIndex = r.matchDNS(ctx, false, ruleIndex, true, &dnsOptions)
			if rule != nil {
				switch action := rule.Action().(type) {
				case *R.RuleActionReject:
					return nil, "", &R.RejectedError{Cause: action.Error(ctx)}
				case *R.RuleActionPredefined:
					if action.Rcode != mDNS.RcodeSuccess {
						err = RcodeError(action.Rcode)
					} else {
						for _, answer := range action.Answer {
							switch record := answer.(type) {
							case *mDNS.A:
								responseAddrs = append(responseAddrs, M.AddrFromIP(record.A))
							case *mDNS.AAAA:
								responseAddrs = append(responseAddrs, M.AddrFromIP(record.AAAA))
							}
						}
					}
					goto response
				}
			}
			if transport != nil { //karing
				transportTag = transport.Tag()
			} else { //karing
				transportTag = ""
			}
			var responseCheck func(responseAddrs []netip.Addr) bool
			if rule != nil && rule.WithAddressLimit() {
				responseCheck = func(responseAddrs []netip.Addr) bool {
					metadata.DestinationAddresses = responseAddrs
					return rule.MatchAddressLimit(metadata)
				}
			}
			if dnsOptions.Strategy == C.DomainStrategyAsIS {
				dnsOptions.Strategy = r.defaultDomainStrategy
			}
			responseAddrs, err = r.client.Lookup(dnsCtx, transport, domain, dnsOptions, responseCheck)
			if responseCheck == nil || err == nil {
				break
			}
			printResult()
		}
	}
response:
	printResult()
	if len(responseAddrs) > 0 {
		lateString := F.MakeLaterString(func() string { //karing
			return strings.Join(F.MapToString(responseAddrs), " ")
		})
		r.logger.InfoContext(ctx, "lookup succeed for ", domain, ": ", lateString) //karing
	}
	return responseAddrs, transportTag, err //karing
}
