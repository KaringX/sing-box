//go:build with_quic

package outbound

import (
	"context"
	"net"
	"os"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/common/humanize"
	"github.com/sagernet/sing-box/common/tls"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/outbound/houtbound" //hiddify
	"github.com/sagernet/sing-quic/hysteria"
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

type Hysteria struct {
	myOutboundAdapter
	client     *hysteria.Client
	hforwarder *houtbound.Forwarder //hiddify
	parseErr    error               //karing
}

func NewHysteria(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.HysteriaOutboundOptions) (*Hysteria, error) {
	empty := &Hysteria{ //karing
		myOutboundAdapter: myOutboundAdapter{
			protocol:     C.TypeHysteria,
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
	outboundDialer, err := dialer.New(router, options.DialerOptions)
	if err != nil {
		return empty, err //karing
	}
	networkList := options.Network.Build()
	if options.HopInterval < 5 { //https://github.com/morgenanno/sing-box
		options.HopInterval = 5
	}
	var password string
	if options.AuthString != "" {
		password = options.AuthString
	} else {
		password = string(options.Auth)
	}
	var sendBps, receiveBps uint64
	if len(options.Up) > 0 {
		sendBps, err = humanize.ParseBytes(options.Up)
		if err != nil {
			return empty, E.Cause(err, "invalid up speed format: ", options.Up) //karing
		}
	} else {
		sendBps = uint64(options.UpMbps) * hysteria.MbpsToBps
	}
	if len(options.Down) > 0 {
		receiveBps, err = humanize.ParseBytes(options.Down)
		if receiveBps == 0 {
			return empty, E.New("invalid down speed format: ", options.Down) //karing
		}
	} else {
		receiveBps = uint64(options.DownMbps) * hysteria.MbpsToBps
	}
	client, err := hysteria.NewClient(hysteria.ClientOptions{
		Context:       ctx,
		Dialer:        outboundDialer,
		Logger:        logger,
		ServerAddress: options.ServerOptions.Build(),
		SendBPS:       sendBps,
		ReceiveBPS:    receiveBps,
		XPlusPassword: options.Obfs,
		Password:      password,
		TLSConfig:     tlsConfig,
		UDPDisabled:   !common.Contains(networkList, N.NetworkUDP),
		HopPorts:      options.HopPorts, //https://github.com/morgenanno/sing-box
		HopInterval:   options.HopInterval, //https://github.com/morgenanno/sing-box

		ConnReceiveWindow:   options.ReceiveWindowConn,
		StreamReceiveWindow: options.ReceiveWindow,
		DisableMTUDiscovery: options.DisableMTUDiscovery,
	})
	if err != nil {
		return empty, err //karing
	}
	return &Hysteria{
		myOutboundAdapter: myOutboundAdapter{
			protocol:     C.TypeHysteria,
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

func (h *Hysteria) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
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

func (h *Hysteria) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if(h.parseErr != nil){ //karing
		return nil, h.parseErr
	}
	h.logger.InfoContext(ctx, "outbound packet connection to ", destination)
	return h.client.ListenPacket(ctx, destination)
}

func (h *Hysteria) NewConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	if(h.parseErr != nil){ //karing
		return h.parseErr
	}
	return NewConnection(ctx, h, conn, metadata)
}

func (h *Hysteria) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	if(h.parseErr != nil){ //karing
		return h.parseErr
	}
	return NewPacketConnection(ctx, h, conn, metadata)
}

func (h *Hysteria) InterfaceUpdated() error {
	return h.client.CloseWithError(E.New("network changed"))
}

func (h *Hysteria) Close() error {
	return h.client.CloseWithError(os.ErrClosed)
}
func (h *Hysteria) SetParseErr(err error){ //karing
	h.parseErr = err
}