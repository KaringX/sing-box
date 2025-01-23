package hysteria

import (
	"context"
	"net"
	"os"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/common/humanize"
	"github.com/sagernet/sing-box/common/tls"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/outbound/houtbound" //hiddify
	"github.com/sagernet/sing-box/protocol/tuic"
	"github.com/sagernet/sing-quic/hysteria"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/bufio"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

func RegisterOutbound(registry *outbound.Registry) {
	outbound.Register[option.HysteriaOutboundOptions](registry, C.TypeHysteria, NewOutbound)
}

var (
	_ adapter.Outbound                = (*tuic.Outbound)(nil)
	_ adapter.InterfaceUpdateListener = (*tuic.Outbound)(nil)
)

type Outbound struct {
	outbound.Adapter
	logger logger.ContextLogger
	client *hysteria.Client
	hforwarder *houtbound.Forwarder //hiddify
	parseErr    error               //karing
}

func NewOutbound(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.HysteriaOutboundOptions) (adapter.Outbound, error) {
	empty := &Outbound{  //karing
		Adapter: outbound.NewAdapterWithDialerOptions(C.TypeHysteria, tag, []string{}, options.DialerOptions),
		logger:  logger,
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
	outboundDialer, err := dialer.New(ctx, options.DialerOptions)
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
		if err != nil {
			return nil, E.Cause(err, "invalid down speed format: ", options.Down)
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
		//HopInterval:   options.HopInterval, //https://github.com/morgenanno/sing-box

		ConnReceiveWindow:   options.ReceiveWindowConn,
		StreamReceiveWindow: options.ReceiveWindow,
		DisableMTUDiscovery: options.DisableMTUDiscovery,
	})
	if err != nil {
		return empty, err //karing
	}
	return &Outbound{
		Adapter: outbound.NewAdapterWithDialerOptions(C.TypeHysteria, tag, networkList, options.DialerOptions),
		logger:  logger,
		client:  client,
	}, nil
}

func (h *Outbound) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
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

func (h *Outbound) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if(h.parseErr != nil){ //karing
		return nil, h.parseErr
	}
	h.logger.InfoContext(ctx, "outbound packet connection to ", destination)
	return h.client.ListenPacket(ctx, destination)
}

func (h *Outbound) InterfaceUpdated() {
	h.client.CloseWithError(E.New("network changed"))
}

func (h *Outbound) Close() error {
	if h.client == nil { //karing
		return nil
	}
	return h.client.CloseWithError(os.ErrClosed)
}
func (h *Outbound) SetParseErr(err error){ //karing
	h.parseErr = err
}