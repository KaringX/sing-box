//go:build with_quic

package outbound

import (
	"context"
	"net"
	"os"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/common/tls"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/outbound/houtbound" //hiddify
	"github.com/sagernet/sing-quic/hysteria"
	"github.com/sagernet/sing-quic/hysteria2"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/bufio"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

var (
	_ adapter.Outbound                = (*TUIC)(nil)
	_ adapter.InterfaceUpdateListener = (*TUIC)(nil)
)

type Hysteria2 struct {
	myOutboundAdapter
	client     *hysteria2.Client
	hforwarder *houtbound.Forwarder //hiddify
	parseErr    error               //karing
}

func NewHysteria2(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.Hysteria2OutboundOptions) (*Hysteria2, error) {
	empty := &Hysteria2{ //karing
		myOutboundAdapter: myOutboundAdapter{
			protocol:     C.TypeHysteria2,
			network:      options.Network.Build(),
			router:       router,
			logger:       logger,
			tag:          tag,
			dependencies: withDialerDependency(options.DialerOptions),
		},
	}
	options.UDPFragmentDefault = true
	if options.TLS == nil || !options.TLS.Enabled {
		return empty, C.ErrTLSRequired //karing
	}
	hforwarder := houtbound.ApplyTurnRelay(houtbound.CommonTurnRelayOptions{ServerOptions: options.ServerOptions, TurnRelayOptions: options.TurnRelay}) //hiddify

	tlsConfig, err := tls.NewClient(ctx, options.Server, common.PtrValueOrDefault(options.TLS))
	if err != nil {
		return empty, err //karing
	}
	var salamanderPassword string
	if options.Obfs != nil {
		if options.Obfs.Password == "" {
			return empty, E.New("missing obfs password") //karing
		}
		switch options.Obfs.Type {
		case hysteria2.ObfsTypeSalamander:
			salamanderPassword = options.Obfs.Password
		default:
			return empty, E.New("unknown obfs type: ", options.Obfs.Type) //karing
		}
	}
	outboundDialer, err := dialer.New(router, options.DialerOptions)
	if err != nil {
		return empty, err //karing
	}
	networkList := options.Network.Build()
	if options.HopInterval < 5 { //https://github.com/morgenanno/sing-box
		options.HopInterval = 5
	}
	client, err := hysteria2.NewClient(hysteria2.ClientOptions{
		Context:            ctx,
		Dialer:             outboundDialer,
		Logger:             logger,
		BrutalDebug:        options.BrutalDebug,
		ServerAddress:      options.ServerOptions.Build(),
		SendBPS:            uint64(options.UpMbps * hysteria.MbpsToBps),
		ReceiveBPS:         uint64(options.DownMbps * hysteria.MbpsToBps),
		SalamanderPassword: salamanderPassword,
		Password:           options.Password,
		TLSConfig:          tlsConfig,
		UDPDisabled:        !common.Contains(networkList, N.NetworkUDP),
		HopPorts:           options.HopPorts, //https://github.com/morgenanno/sing-box
		HopInterval:        options.HopInterval, //https://github.com/morgenanno/sing-box
	})
	if err != nil {
		return empty, err //karing
	}
	return &Hysteria2{
		myOutboundAdapter: myOutboundAdapter{
			protocol:     C.TypeHysteria2,
			network:      networkList,
			router:       router,
			logger:       logger,
			tag:          tag,
			dependencies: withDialerDependency(options.DialerOptions),
		},
		client:     client,
		hforwarder: hforwarder, //hiddify
	}, nil
}

func (h *Hysteria2) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	if(h.parseErr != nil){ //karing
		return nil, h.parseErr
	}
	switch N.NetworkName(network) {
	case N.NetworkTCP:
		h.logger.InfoContext(ctx, "outbound connection to ", destination)
		return h.client.DialConn(ctx, destination)
	case N.NetworkUDP:
		conn, err := h.ListenPacket(ctx, destination)
		if err != nil {
			return nil, err
		}
		return bufio.NewBindPacketConn(conn, destination), nil
	default:
		return nil, E.New("unsupported network: ", network)
	}
}

func (h *Hysteria2) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if(h.parseErr != nil){ //karing
		return nil, h.parseErr
	}
	h.logger.InfoContext(ctx, "outbound packet connection to ", destination)
	return h.client.ListenPacket(ctx)
}

func (h *Hysteria2) NewConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	if(h.parseErr != nil){ //karing
		return h.parseErr
	}
	return NewConnection(ctx, h, conn, metadata)
}

func (h *Hysteria2) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	if(h.parseErr != nil){ //karing
		return h.parseErr
	}
	return NewPacketConnection(ctx, h, conn, metadata)
}

func (h *Hysteria2) InterfaceUpdated() {
	if h.client == nil { //karing
		return
	}
	h.client.CloseWithError(E.New("network changed"))
}

func (h *Hysteria2) Close() error {
	if h.hforwarder != nil { //hiddify
		h.hforwarder.Close()
	}
	if h.client == nil { //karing
		return nil
	}
	return h.client.CloseWithError(os.ErrClosed)
}
func (h *Hysteria2) SetParseErr(err error){ //karing
	h.parseErr = err
}