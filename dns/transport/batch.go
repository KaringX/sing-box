// karing
package transport

import (
	"context"
	"sync"
	"sync/atomic"

	mDNS "github.com/miekg/dns"
	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/dns"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	"github.com/sagernet/sing/service"
)

var _ adapter.DNSTransport = (*BatchTransport)(nil)

func RegisterBatch(registry *dns.TransportRegistry) {
	dns.RegisterTransport[option.BatchDNSServerOptions](registry, C.DNSTypeBatch, NewBatch)
}

type BatchTransport struct {
	dns.TransportAdapter
	logger  logger.ContextLogger
	servers []string
}

func NewBatch(ctx context.Context, logger log.ContextLogger, tag string, options option.BatchDNSServerOptions) (adapter.DNSTransport, error) {
	transportManager := service.FromContext[adapter.DNSTransportManager](ctx)
	for _, server := range options.Servers {
		_, loaded := transportManager.Transport(server)
		if !loaded {
			return nil, E.New("dns dependencies server not found: " + server)
		}
	}
	return &BatchTransport{
		TransportAdapter: dns.NewTransportAdapter(C.DNSTypeBatch, tag, nil),
		logger:           logger,
		servers:          options.Servers,
	}, nil
}

func (t *BatchTransport) Start(stage adapter.StartStage) error {
	return nil
}

func (t *BatchTransport) Close() error {
	return nil
}

func (t *BatchTransport) Exchange(ctx context.Context, message *mDNS.Msg) (*mDNS.Msg, error) {
	var transports []adapter.DNSTransport
	transportManager := service.FromContext[adapter.DNSTransportManager](ctx)
	if transportManager == nil {
		return nil, E.New("dns transportManager is nil:", t.Tag())
	}
	for _, server := range t.servers {
		transport, loaded := transportManager.Transport(server)
		if loaded {
			transports = append(transports, transport)
		}
	}
	if len(transports) == 0 {
		return nil, E.New("dns transport empty :", t.Tag())
	}
	question := message.Question[0]
	domain := question.Name
	if mDNS.IsFqdn(domain) {
		domain = domain[:len(domain)-1]
	}
	var count atomic.Int64
	var result *mDNS.Msg
	var errResult error
	var once sync.Once
	var errOnce sync.Once
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(ctx)
	for _, transport := range transports {
		count.Add(1)
		transport := transport
		go func() {
			copydMessage := message.Copy()
			ret, err := transport.Exchange(ctx, copydMessage)
			count.Add(-1)
			if err == nil {
				once.Do(func() {
					result = ret
					done <- struct{}{}
					t.logger.InfoContext(ctx, "exchanged ["+domain+"] by: ", transport.Tag())
				})
			} else {
				errOnce.Do(func() {
					errResult = err
				})
				if count.Load() == 0 {
					once.Do(func() {
						done <- struct{}{}
					})
				}
			}
		}()
	}
	<-done
	cancel()
	close(done)
	if result == nil && errResult == nil {
		errResult = E.New("exchange: all failed")
	} else if result != nil {
		errResult = nil
	}
	return result, errResult
}
