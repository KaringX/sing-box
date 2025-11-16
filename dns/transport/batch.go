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
	logger    logger.ContextLogger
	transport adapter.DNSTransportManager
	servers   []string
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
		transport:        transportManager,
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
	for _, server := range t.servers {
		transport, loaded := t.transport.Transport(server)
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
	var resultEmpty *mDNS.Msg
	var errResult error
	var once sync.Once
	var errOnce sync.Once
	var emptyOnce sync.Once
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(ctx)

	count.Add(int64(len(transports)))
	for _, transport := range transports {
		transport := transport
		go func() {
			copydMessage := message.Copy()
			ret, err := transport.Exchange(ctx, copydMessage)
			if err == nil {
				if len(ret.Answer) == 0 {
					emptyOnce.Do(func() {
						resultEmpty = ret
						t.logger.InfoContext(ctx, "exchanged empty result ["+domain+"] by: ", transport.Tag())
					})
				} else {
					once.Do(func() {
						result = ret
						done <- struct{}{}
						t.logger.InfoContext(ctx, "exchanged ["+domain+"] by: ", transport.Tag())
					})
				}
			} else {
				errOnce.Do(func() {
					errResult = err
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
	if result != nil {
		return result, nil
	}
	if resultEmpty != nil {
		return resultEmpty, nil
	}
	if errResult != nil {
		return nil, errResult
	}

	return nil, E.New("batch exchange: unknown error")
}
