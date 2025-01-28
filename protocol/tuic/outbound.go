package tuic

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/common/tls"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-quic/tuic"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/bufio"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/uot"

	"github.com/gofrs/uuid/v5"
	"github.com/sagernet/sing-box/protocol/wireguard/houtbound" //hiddify
)

func RegisterOutbound(registry *outbound.Registry) {
	outbound.Register[option.TUICOutboundOptions](registry, C.TypeTUIC, NewOutbound)
}

var _ adapter.InterfaceUpdateListener = (*Outbound)(nil)

type Outbound struct {
	outbound.Adapter
	logger    logger.ContextLogger
	client    *tuic.Client
	udpStream bool
	hforwarder *houtbound.Forwarder //hiddify
	parseErr  error                //karing
}

func NewOutbound(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.TUICOutboundOptions) (adapter.Outbound, error) {
	empty := &Outbound{ //karing
		Adapter: outbound.NewAdapterWithDialerOptions(C.TypeTUIC, tag, []string{}, options.DialerOptions),
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
	userUUID, err := uuid.FromString(options.UUID)
	if err != nil {
		return empty, E.Cause(err, "invalid uuid") //karing
	}
	var tuicUDPStream bool
	if options.UDPOverStream && options.UDPRelayMode != "" {
		return empty, E.New("udp_over_stream is conflict with udp_relay_mode") //karing
	}
	switch options.UDPRelayMode {
	case "native":
	case "quic":
		tuicUDPStream = true
	}
	outboundDialer, err := dialer.New(ctx, options.DialerOptions)
	if err != nil {
		return empty, err //karing
	}
	client, err := tuic.NewClient(tuic.ClientOptions{
		Context:           ctx,
		Dialer:            outboundDialer,
		ServerAddress:     options.ServerOptions.Build(),
		TLSConfig:         tlsConfig,
		UUID:              userUUID,
		Password:          options.Password,
		CongestionControl: options.CongestionControl,
		UDPStream:         tuicUDPStream,
		ZeroRTTHandshake:  options.ZeroRTTHandshake,
		Heartbeat:         time.Duration(options.Heartbeat),
	})
	if err != nil {
		return empty, err //karing
	}
	return &Outbound{
		Adapter:   outbound.NewAdapterWithDialerOptions(C.TypeTUIC, tag, options.Network.Build(), options.DialerOptions),
		logger:    logger,
		client:    client,
		udpStream: options.UDPOverStream,
		hforwarder: hforwarder, //hiddify
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
		if h.udpStream {
			h.logger.InfoContext(ctx, "outbound stream packet connection to ", destination)
			streamConn, err := h.client.DialConn(ctx, uot.RequestDestination(uot.Version))
			if err != nil {
				return nil, err
			}
			return uot.NewLazyConn(streamConn, uot.Request{
				IsConnect:   true,
				Destination: destination,
			}), nil
		} else {
			conn, err := h.ListenPacket(ctx, destination)
			if err != nil {
				return nil, err
			}
			return bufio.NewBindPacketConn(conn, destination), nil
		}
	default:
		return nil, E.New("unsupported network: ", network)
	}
}

func (h *Outbound) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if(h.parseErr != nil){ //karing
		return nil, h.parseErr
	}
	if h.udpStream {
		h.logger.InfoContext(ctx, "outbound stream packet connection to ", destination)
		streamConn, err := h.client.DialConn(ctx, uot.RequestDestination(uot.Version))
		if err != nil {
			return nil, err
		}
		return uot.NewLazyConn(streamConn, uot.Request{
			IsConnect:   false,
			Destination: destination,
		}), nil
	} else {
		h.logger.InfoContext(ctx, "outbound packet connection to ", destination)
		return h.client.ListenPacket(ctx)
	}
}

func (h *Outbound) InterfaceUpdated() {
	if h.client == nil { //karing
		return
	}
	_ = h.client.CloseWithError(E.New("network changed"))
}

func (h *Outbound) Close() error {
	if h.hforwarder != nil { //hiddify
		h.hforwarder.Close()
	}
	if h.client == nil { //karing
		return nil
	}
	return h.client.CloseWithError(os.ErrClosed)
}
func (h *Outbound) SetParseErr(err error){ //karing
	h.parseErr = err
}