// karing
package dns

import (
	"context"
	"net/netip"
	"sync"
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
	var response4 []netip.Addr
	var response6 []netip.Addr
	dnsQueryTypes := []uint16{dns.TypeA, dns.TypeAAAA}
	var count atomic.Int64
	var once sync.Once
	var errOnce sync.Once
	var returnError error
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(ctx)
	count.Add(int64(len(dnsQueryTypes)))
	for _, queryType := range dnsQueryTypes {
		go func() {
			response, err := c.lookupToExchange(ctx, transport, dnsName, queryType, options, responseChecker)
			if err == nil {
				if len(response) > 0 {
					switch queryType {
					case dns.TypeA:
						response4 = response
						if strategy == C.DomainStrategyPreferIPv4 {
							once.Do(func() {
								done <- struct{}{}
							})
						}
					case dns.TypeAAAA:
						response6 = response
						if strategy == C.DomainStrategyPreferIPv6 {
							once.Do(func() {
								done <- struct{}{}
							})
						}
					}
				}
			} else {
				errOnce.Do(func() {
					returnError = E.Cause(err, "dns exchange type: "+dns.TypeToString[queryType])
				})
			}
			count.Add(-1)
			if count.Load() == 0 {
				once.Do(func() {
					done <- struct{}{}
				})
			}
		}()
	}
	<-done
	cancel()
	close(done)
	len4 := len(response4)
	len6 := len(response6)
	if len4 == 0 && len6 == 0 {
		return nil, nil, returnError
	}
	if len4 != 0 && len6 == 0 {
		return response4, nil, nil
	}
	if len4 == 0 && len6 != 0 {
		return nil, response6, nil
	}
	return response4, response6, nil
}
