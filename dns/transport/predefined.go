// karing
package transport

import (
	"context"

	mDNS "github.com/miekg/dns"
	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/dns"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
)

var _ adapter.DNSTransport = (*PredefinedTransport)(nil)

func RegisterPredefined(registry *dns.TransportRegistry) {
	dns.RegisterTransport[option.PredefinedDNSServerOptions](registry, C.DNSTypePredefined, NewPredefinedTransport)
}

type PredefinedTransport struct {
	dns.TransportAdapter
	Rcode int
}

func NewPredefinedTransport(ctx context.Context, logger log.ContextLogger, tag string, options option.PredefinedDNSServerOptions) (adapter.DNSTransport, error) {
	return &PredefinedTransport{
		TransportAdapter: dns.NewTransportAdapter(C.DNSTypePredefined, tag, nil),
		Rcode:            options.Rcode.Build(),
	}, nil
}

func (t *PredefinedTransport) Start(stage adapter.StartStage) error {
	return nil
}

func (t *PredefinedTransport) Close() error {
	return nil
}

func (t *PredefinedTransport) Exchange(ctx context.Context, message *mDNS.Msg) (*mDNS.Msg, error) {
	question := message.Question[0]
	return &mDNS.Msg{
		MsgHdr: mDNS.MsgHdr{
			Id:                 message.Id,
			Rcode:              t.Rcode,
			Response:           true,
			Authoritative:      true,
			RecursionDesired:   true,
			RecursionAvailable: true,
		},
		Question: []mDNS.Question{question},
	}, nil
}
