// karing
package dns

import (
	"context"
	"net/netip"
	"sync/atomic"

	"github.com/miekg/dns"
	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	E "github.com/sagernet/sing/common/exceptions"
)

func (c *Client) Close() {
	c.ClearCache()
	c.cache = nil
	c.transportCache = nil
	c.rdrc = nil
}

func (c *Client) lookupToExchange_A_AAAA(ctx context.Context, transport adapter.DNSTransport, dnsName string, strategy C.DomainStrategy, options adapter.DNSQueryOptions, responseChecker func(responseAddrs []netip.Addr) bool) ([]netip.Addr, []netip.Addr, error) {
	var response4 []netip.Addr = []netip.Addr{}
	var response6 []netip.Addr = []netip.Addr{}
	dnsQueryTypes := []uint16{dns.TypeA, dns.TypeAAAA}
	var returnError error
	var count atomic.Int64
	var once atomic.Bool
	var errOnce atomic.Bool
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(ctx)
	count.Add(int64(len(dnsQueryTypes)))
	for _, queryType := range dnsQueryTypes {
		go func(qtype uint16) {
			response, err := c.lookupToExchange(ctx, transport, dnsName, qtype, options, responseChecker)
			if err == nil {
				if len(response) > 0 {
					switch qtype {
					case dns.TypeA:
						response4 = response
						if strategy == C.DomainStrategyPreferIPv4 {
							if once.CompareAndSwap(false, true) {
								done <- struct{}{}
							}
						}
					case dns.TypeAAAA:
						response6 = response
						if strategy == C.DomainStrategyPreferIPv6 {
							if once.CompareAndSwap(false, true) {
								done <- struct{}{}
							}
						}
					}
				}
			} else {
				if errOnce.CompareAndSwap(false, true) {
					returnError = E.Cause(err, "dns exchange type: "+dns.TypeToString[qtype])
				}
			}

			if count.Add(-1) == 0 {
				if once.CompareAndSwap(false, true) {
					done <- struct{}{}
				}
			}
		}(queryType)
	}
	<-done
	cancel()
	close(done)
	if len(response4) == 0 && len(response6) == 0 {
		return nil, nil, returnError
	}
	return response4, response6, nil
}
