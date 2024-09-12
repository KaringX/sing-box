//go:build with_quic

package outbound

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/common/tls"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-quic/tuic"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/bufio"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/uot"

	"github.com/gofrs/uuid/v5"

	"github.com/sagernet/sing-box/outbound/houtbound" //hiddify
)

var (
	_ adapter.Outbound                = (*TUIC)(nil)
	_ adapter.InterfaceUpdateListener = (*TUIC)(nil)
)

type TUIC struct {
	myOutboundAdapter
	client     *tuic.Client
	udpStream  bool
	hforwarder *houtbound.Forwarder //hiddify
	parseErr   error                //karing
}

func NewTUIC(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.TUICOutboundOptions) (*TUIC, error) {
	empty := &TUIC{ //karing
		myOutboundAdapter: myOutboundAdapter{
			protocol:     C.TypeTUIC,
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
	outboundDialer, err := dialer.New(router, options.DialerOptions)
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
	return &TUIC{
		myOutboundAdapter: myOutboundAdapter{
			protocol:     C.TypeTUIC,
			network:      options.Network.Build(),
			router:       router,
			logger:       logger,
			tag:          tag,
			dependencies: withDialerDependency(options.DialerOptions),
		},
		client:     client,
		udpStream:  options.UDPOverStream,
		hforwarder: hforwarder, //hiddify
	}, nil
}

func (h *TUIC) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
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

func (h *TUIC) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
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

func (h *TUIC) NewConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	if(h.parseErr != nil){ //karing
		return h.parseErr
	}
	return NewConnection(ctx, h, conn, metadata)
}

func (h *TUIC) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	if(h.parseErr != nil){ //karing
		return h.parseErr
	}
	return NewPacketConnection(ctx, h, conn, metadata)
}

func (h *TUIC) InterfaceUpdated() {
	_ = h.client.CloseWithError(E.New("network changed"))
}

func (h *TUIC) Close() error {
	if h.hforwarder != nil { //hiddify
		h.hforwarder.Close()
	}
	return h.client.CloseWithError(os.ErrClosed)
}
func (h *TUIC) SetParseErr(err error){ //karing
	h.parseErr = err
}