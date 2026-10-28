// https://github.com/shtorm-7/sing-box-extended
package masque

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/common/cloudflare"
	"github.com/sagernet/sing-box/common/congestion"
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/common/tls"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/transport/masque"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/bufio"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json/badoption"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/ntp"
	"github.com/sagernet/sing/service"
)

func RegisterOutbound(registry *outbound.Registry) {
	outbound.Register[option.MASQUEOutboundOptions](registry, C.TypeMASQUE, NewOutbound)
}

type tunnelClient interface {
	N.Dialer
	Close() error
}

type Outbound struct {
	outbound.Adapter
	ctx          context.Context
	dnsRouter    adapter.DNSRouter
	logger       logger.ContextLogger
	options      option.MASQUEOutboundOptions
	client       tunnelClient
	startHandler func()

	await chan struct{}
}

func NewOutbound(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.MASQUEOutboundOptions) (adapter.Outbound, error) {
	networks := []string{N.NetworkTCP, N.NetworkUDP}
	if options.UseL4Proxy {
		networks = []string{N.NetworkTCP}
	}
	outbound := &Outbound{
		Adapter:   outbound.NewAdapterWithDialerOptions(C.TypeMASQUE, tag, networks, options.DialerOptions),
		ctx:       ctx,
		dnsRouter: service.FromContext[adapter.DNSRouter](ctx),
		logger:    logger,
		options:   options,
		await:     make(chan struct{}),
	}
	if options.Address == "" { //karing
		return outbound, fmt.Errorf("address is empty")
	}

	if len(options.Profile.IPv4Address) > 0 { //karing
		_, err := netip.ParsePrefix(options.Profile.IPv4Address)
		if err != nil {
			return outbound, fmt.Errorf("invalid IPv4 address: %w %s", err, options.Profile.IPv4Address)
		}
	}
	if len(options.Profile.IPv6Address) > 0 { //karing
		_, err := netip.ParsePrefix(options.Profile.IPv6Address)
		if err != nil {
			return outbound, fmt.Errorf("invalid IPv6 address: %w %s", err, options.Profile.IPv6Address)
		}
	}
	outbound.startHandler = func() {
		defer close(outbound.await)
		//karing begin
		var err error
		defer func() {
			if err != nil {
				outbound.SetParseErr(err)
			}
		}()

		var privKey *ecdsa.PrivateKey
		var peerPubKey *ecdsa.PublicKey
		privKeyBytes, err1 := base64.StdEncoding.DecodeString(options.Profile.PrivateKey)
		if err1 != nil {
			err = err1
			logger.ErrorContext(ctx, E.New("failed to decode private key: ", err))
			return
		}
		privKey, err1 = x509.ParseECPrivateKey(privKeyBytes)
		if err1 != nil {
			err = err1
			logger.ErrorContext(ctx, E.New("failed to parse private key: ", err))
			return
		}
		peerPubKeyBytes, err1 := base64.StdEncoding.DecodeString(options.Profile.PublicKey)
		if err1 != nil {
			err = err1
			logger.ErrorContext(ctx, E.New("failed to decode public key: ", err))
			return
		}
		pubKey, err1 := x509.ParsePKIXPublicKey(peerPubKeyBytes)
		if err1 != nil {
			err = err1
			logger.ErrorContext(ctx, E.New("failed to parse public key: ", err))
			return
		}
		var ok bool
		peerPubKey, ok = pubKey.(*ecdsa.PublicKey)
		if !ok {
			err = fmt.Errorf("public key is not ECDSA")
			logger.ErrorContext(ctx, E.New("public key is not ECDSA"))
			return
		}
		//karing end
		cert, err1 := masque.GenerateCert(privKey, &privKey.PublicKey) //karing
		if err1 != nil {                                               //karing
			err = err1 //karing
			logger.ErrorContext(ctx, E.New("failed to generate cert: ", err))
			return
		}
		serverName := cloudflare.ConnectSNI
		if options.UseL4Proxy {
			serverName = cloudflare.L4ConnectSNI
		}
		if options.TLS != nil && options.TLS.ServerName != "" {
			serverName = options.TLS.ServerName
		}
		tlsConfig, err1 := tls.NewMASQUEClient(ctx, logger, serverName, cert, privKey, peerPubKey, common.PtrValueOrDefault(options.TLS)) //karing
		if err1 != nil {                                                                                                                  //karing                                                                                                     //karing
			err = err1 //karing
			logger.ErrorContext(ctx, E.New("failed to prepare TLS config: ", err))
			return
		}
		useHTTP2 := options.UseHTTP2 && !options.UseL4Proxy
		var endpoint net.Addr
		endpointPort := options.Port
		if endpointPort == 0 {
			endpointPort = 443
		}
		address := M.ParseSocksaddr(options.Address) //karing
		var endpointIP net.IP                        //karing
		if address.IsDomain() {                      //karing
			destinationAddresses, err1 := outbound.dnsRouter.Lookup(ctx, address.Fqdn, adapter.DNSQueryOptions{})
			if err1 != nil {
				err = err1
				logger.ErrorContext(ctx, E.New("failed to lookup destination addresses: ", err))
				return
			}
			if len(destinationAddresses) == 0 {
				err = E.New("no addresses found for: ", address.Fqdn)
				logger.ErrorContext(ctx, err)
				return
			}
			endpointIP = net.IP(destinationAddresses[0].AsSlice())
		} else {
			endpointIP = net.ParseIP(options.Address)
		}

		if endpointIP == nil {
			err = E.New("invalid endpoint address: ", options.Address)
			logger.ErrorContext(ctx, E.New("invalid endpoint address: ", options.Address))
			return
		}
		if useHTTP2 {
			endpoint = &net.TCPAddr{IP: endpointIP, Port: int(endpointPort)}
		} else {
			endpoint = &net.UDPAddr{IP: endpointIP, Port: int(endpointPort)}
		}

		var udpTimeout time.Duration
		if options.UDPTimeout != 0 {
			udpTimeout = time.Duration(options.UDPTimeout)
		} else {
			udpTimeout = C.UDPTimeout
		}
		var udpKeepalivePeriod time.Duration
		if options.UDPKeepalivePeriod != 0 {
			udpKeepalivePeriod = time.Duration(options.UDPKeepalivePeriod)
		} else {
			udpKeepalivePeriod = time.Second * 30
		}
		outboundDialer, err1 := dialer.NewWithOptions(dialer.Options{ //karing
			Context:          ctx,
			Options:          options.DialerOptions,
			RemoteIsDomain:   false,
			ResolverOnDetour: true,
		})
		if err1 != nil { //karing
			err = err1 //karing
			logger.ErrorContext(ctx, err)
			return
		}
		congestionControl, err1 := congestion.NewCongestionControl( //karing
			options.CongestionController,
			options.CWND,
			ntp.TimeFuncFromContext(ctx),
		)
		if err1 != nil { //karing
			err = err1 //karing
			logger.ErrorContext(ctx, err)
			return
		}
		if options.UseL4Proxy {
			outbound.client = masque.NewL4Client(logger, masque.L4ClientOptions{
				Dialer:               outboundDialer,
				TLSConfig:            tlsConfig,
				Endpoint:             endpoint,
				UDPKeepalivePeriod:   udpKeepalivePeriod,
				UDPInitialPacketSize: options.UDPInitialPacketSize,
				HandshakeTimeout:     time.Duration(options.HandshakeTimeout),
				CongestionControl:    congestionControl,
			})
			return
		}
		if options.ReconnectDelay == 0 { //karing
			options.ReconnectDelay = badoption.Duration(time.Duration(3 * time.Second))
		}
		if options.MTU == 0 { //karing
			options.MTU = 1280
		}
		tunnel, err1 := masque.NewTunnel( //karing
			ctx,
			logger,
			masque.TunnelOptions{
				System: options.System,
				Name:   options.Name,
				CreateDialer: func(interfaceName string) N.Dialer {
					return common.Must1(dialer.NewDefault(ctx, option.DialerOptions{
						AbstractDialerOptions: option.AbstractDialerOptions{
							BindInterface: interfaceName,
						},
					}))
				},
				Dialer: outboundDialer,
				Address: []netip.Prefix{
					netip.MustParsePrefix(options.Profile.IPv4Address), //karing
					netip.MustParsePrefix(options.Profile.IPv6Address), //karing
				},
				AllowedAddress:       options.AllowedIPs,
				Endpoint:             endpoint,
				TLSConfig:            tlsConfig,
				UseHTTP2:             options.UseHTTP2,
				UDPTimeout:           udpTimeout,
				UDPKeepalivePeriod:   udpKeepalivePeriod,
				UDPInitialPacketSize: options.UDPInitialPacketSize,
				ReconnectDelay:       options.ReconnectDelay.Build(),
				HandshakeTimeout:     time.Duration(options.HandshakeTimeout),
				CongestionControl:    congestionControl,
				MTU:                  options.MTU, //karing
			},
		)
		if err1 != nil { //karing
			err = err1 //karing
			logger.ErrorContext(ctx, err)
			return
		}
		outbound.client = tunnel
		if err = tunnel.Start(false); err != nil {
			logger.ErrorContext(ctx, err)
			return
		}
		if err = tunnel.Start(true); err != nil {
			logger.ErrorContext(ctx, err)
			return
		}
	}
	return outbound, nil
}

func (w *Outbound) Start(stage adapter.StartStage, scope *adapter.Scope) error {
	if w.GetParseErr() != nil { //karing
		return nil
	}
	if stage != adapter.StartStatePostStart {
		return nil
	}
	scope.Add(func() error {
		if err := w.isTunnelInitialized(w.ctx); err != nil {
			return err
		}
		return w.client.Close()
	})
	go w.startHandler()
	return nil
}

func (w *Outbound) DialContext(ctx context.Context, network string, destination M.Socksaddr) (conn net.Conn, err error) { //karing
	if w.GetParseErr() != nil { //karing
		return nil, w.GetParseErr()
	}
	/*defer func() { //karing
		if err == nil {
			conn = w.OnNewConnection(conn, func() {
				w.Adapter.ConnectionsIn.Store(0)
				if err := w.isTunnelInitialized(w.ctx); err != nil {
					return
				}
				w.client.Close()
			})
		}
	}()*/
	if err := w.isTunnelInitialized(ctx); err != nil {
		return nil, err
	}
	switch network {
	case N.NetworkTCP:
		w.logger.InfoContext(ctx, "outbound connection to ", destination)
	case N.NetworkUDP:
		w.logger.InfoContext(ctx, "outbound packet connection to ", destination)
	}
	if destination.IsDomain() {
		destinationAddresses, err := w.dnsRouter.Lookup(ctx, destination.Fqdn, adapter.DNSQueryOptions{})
		if err != nil {
			return nil, err
		}
		return N.DialSerial(ctx, w.client, network, destination, destinationAddresses)
	} else if !destination.Addr.IsValid() {
		return nil, E.New("invalid destination: ", destination)
	}
	return w.client.DialContext(ctx, network, destination)
}

func (w *Outbound) ListenPacketWithDestination(ctx context.Context, destination M.Socksaddr) (conn net.PacketConn, addr netip.Addr, err error) { //karing
	if w.GetParseErr() != nil { //karing
		return nil, netip.Addr{}, w.GetParseErr()
	}
	/*defer func() { //karing
		if err == nil {
			conn = w.OnNewPacketConnection(conn, func() {
				w.Adapter.ConnectionsIn.Store(0)
				if err := w.isTunnelInitialized(w.ctx); err != nil {
					return
				}
				w.client.Close()
			})
		}
	}()*/
	if err := w.isTunnelInitialized(ctx); err != nil {
		return nil, netip.Addr{}, err
	}
	w.logger.InfoContext(ctx, "outbound packet connection to ", destination)
	if destination.IsDomain() {
		destinationAddresses, err := w.dnsRouter.Lookup(ctx, destination.Fqdn, adapter.DNSQueryOptions{})
		if err != nil {
			return nil, netip.Addr{}, err
		}
		return N.ListenSerial(ctx, w.client, destination, destinationAddresses)
	}
	packetConn, err := w.client.ListenPacket(ctx, destination)
	if err != nil {
		return nil, netip.Addr{}, err
	}
	if destination.IsIP() {
		return packetConn, destination.Addr, nil
	}
	return packetConn, netip.Addr{}, nil
}

func (w *Outbound) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if w.GetParseErr() != nil { //karing
		return nil, w.GetParseErr()
	}
	packetConn, destinationAddress, err := w.ListenPacketWithDestination(ctx, destination)
	if err != nil {
		return nil, err
	}
	if destinationAddress.IsValid() && destination != M.SocksaddrFrom(destinationAddress, destination.Port) {
		return bufio.NewNATPacketConn(bufio.NewPacketConn(packetConn), M.SocksaddrFrom(destinationAddress, destination.Port), destination), nil
	}
	return packetConn, nil
}

func (w *Outbound) isTunnelInitialized(ctx context.Context) error {
	select {
	case <-w.await:
	case <-ctx.Done():
		return ctx.Err()
	}
	if w.client == nil {
		if w.GetParseErr() != nil { //karing
			return w.GetParseErr()
		}
		return E.New("tunnel not initialized")
	}
	return nil
}
