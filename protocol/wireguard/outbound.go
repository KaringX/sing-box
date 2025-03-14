package wireguard

import (
	"context"
	"net"
	"net/netip"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/common/dialer"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/experimental/deprecated"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option" //hiddify
	"github.com/sagernet/sing-box/protocol/wireguard/houtbound"
	"github.com/sagernet/sing-box/transport/wireguard"
	dns "github.com/sagernet/sing-dns"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

func RegisterOutbound(registry *outbound.Registry) {
	outbound.Register[option.LegacyWireGuardOutboundOptions](registry, C.TypeWireGuard, NewOutbound)
}

var (
	_ adapter.Endpoint                = (*Endpoint)(nil)
	_ adapter.InterfaceUpdateListener = (*Endpoint)(nil)
)

type Outbound struct {
	outbound.Adapter
	ctx            context.Context
	router         adapter.Router
	logger         logger.ContextLogger
	localAddresses []netip.Prefix
	endpoint       *wireguard.Endpoint
	hforwarder     *houtbound.Forwarder //hiddify
	parseErr       error                //karing
}

func NewOutbound(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.LegacyWireGuardOutboundOptions) (adapter.Outbound, error) {
	empty := &Outbound{ //karing
		Adapter: outbound.NewAdapterWithDialerOptions(C.TypeWireGuard, tag, []string{}, options.DialerOptions),
		logger:  logger,
	}
	deprecated.Report(ctx, deprecated.OptionWireGuardOutbound)
	if options.GSO {
		deprecated.Report(ctx, deprecated.OptionWireGuardGSO)
	}
	if len(options.LocalAddress) == 0 { //karing
		return empty, E.New("missing local address")
	}
	for _, prefix := range options.LocalAddress { //karing
		if !prefix.IsValid() {
			return empty, E.New("invalid local address")
		}
	}
	hforwarder := houtbound.ApplyTurnRelay(houtbound.CommonTurnRelayOptions{ServerOptions: options.ServerOptions, TurnRelayOptions: options.TurnRelay}) //hiddify
	outbound := &Outbound{
		Adapter:        outbound.NewAdapterWithDialerOptions(C.TypeWireGuard, tag, []string{N.NetworkTCP, N.NetworkUDP}, options.DialerOptions),
		ctx:            ctx,
		router:         router,
		logger:         logger,
		localAddresses: options.LocalAddress,
		hforwarder:     hforwarder, //hiddify
	}

	if options.Detour == "" {
		options.IsWireGuardListener = true
	} else if options.GSO {
		return empty, E.New("gso is conflict with detour") //karing
	}
	outboundDialer, err := dialer.New(ctx, options.DialerOptions)
	if err != nil {
		return empty, err //karing
	}
	peers := common.Map(options.Peers, func(it option.LegacyWireGuardPeer) wireguard.PeerOptions {
		return wireguard.PeerOptions{
			Endpoint:     it.ServerOptions.Build(),
			PublicKey:    it.PublicKey,
			PreSharedKey: it.PreSharedKey,
			AllowedIPs:   it.AllowedIPs,
			// PersistentKeepaliveInterval: time.Duration(it.PersistentKeepaliveInterval),
			Reserved: it.Reserved,
		}
	})
	if len(peers) == 0 {
		peers = []wireguard.PeerOptions{{
			Endpoint:     options.ServerOptions.Build(),
			PublicKey:    options.PeerPublicKey,
			PreSharedKey: options.PreSharedKey,
			AllowedIPs:   []netip.Prefix{netip.PrefixFrom(netip.IPv4Unspecified(), 0), netip.PrefixFrom(netip.IPv6Unspecified(), 0)},
			Reserved:     options.Reserved,
		}}
	}
	wgEndpoint, err := wireguard.NewEndpoint(wireguard.EndpointOptions{
		Context: ctx,
		Logger:  logger,
		System:  options.SystemInterface,
		Dialer:  outboundDialer,
		CreateDialer: func(interfaceName string) N.Dialer {
			return common.Must1(dialer.NewDefault(ctx, option.DialerOptions{
				BindInterface: interfaceName,
			}))
		},
		Name:       options.InterfaceName,
		MTU:        options.MTU,
		Address:    options.LocalAddress,
		PrivateKey: options.PrivateKey,
		ResolvePeer: func(domain string) (netip.Addr, error) {
			endpointAddresses, _, lookupErr := router.Lookup(ctx, domain, dns.DomainStrategy(options.DomainStrategy)) //karing
			if lookupErr != nil {
				return netip.Addr{}, lookupErr
			}
			return endpointAddresses[0], nil
		},
		Peers:            peers,
		Workers:          options.Workers,
		FakePackets:      options.FakePackets,      //hiddify
		FakePacketsSize:  options.FakePacketsSize,  //hiddify
		FakePacketsDelay: options.FakePacketsDelay, //hiddify
		FakePacketsMode:  options.FakePacketsMode,  //hiddify
	})
	if err != nil {
		return empty, err //karing
	}
	outbound.endpoint = wgEndpoint
	return outbound, nil
}

func (o *Outbound) Start(stage adapter.StartStage) error {
	if o.endpoint == nil { //karing
		return nil
	}
	switch stage {
	case adapter.StartStateStart:
		return o.endpoint.Start(false)
	case adapter.StartStatePostStart:
		return o.endpoint.Start(true)
	}
	return nil
}

func (o *Outbound) Close() error {
	if o.endpoint == nil { //karing
		return nil
	}
	return o.endpoint.Close()
}

func (h *Outbound) SetParseErr(err error) { //karing
	h.parseErr = err
}

func (o *Outbound) InterfaceUpdated() {
	if o.endpoint == nil { //karing
		return
	}
	o.endpoint.BindUpdate()
	return
}

func (o *Outbound) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	if o.parseErr != nil { //karing
		return nil, o.parseErr
	}
	switch network {
	case N.NetworkTCP:
		o.logger.InfoContext(ctx, "outbound connection to ", destination)
	case N.NetworkUDP:
		o.logger.InfoContext(ctx, "outbound packet connection to ", destination)
	}
	if destination.IsFqdn() {
		destinationAddresses, err := o.router.LookupDefault(ctx, destination.Fqdn)
		if err != nil {
			return nil, err
		}
		return N.DialSerial(ctx, o.endpoint, network, destination, destinationAddresses)
	} else if !destination.Addr.IsValid() {
		return nil, E.New("invalid destination: ", destination)
	}
	return o.endpoint.DialContext(ctx, network, destination)
}

func (o *Outbound) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if o.parseErr != nil { //karing
		return nil, o.parseErr
	}
	o.logger.InfoContext(ctx, "outbound packet connection to ", destination)
	if destination.IsFqdn() {
		destinationAddresses, err := o.router.LookupDefault(ctx, destination.Fqdn)
		if err != nil {
			return nil, err
		}
		packetConn, _, err := N.ListenSerial(ctx, o.endpoint, destination, destinationAddresses)
		if err != nil {
			return nil, err
		}
		return packetConn, err
	}
	return o.endpoint.ListenPacket(ctx, destination)
}
