package direct

import (
	"context"
	"net/netip"
	"testing"

	"github.com/sagernet/sing-box/adapter"
	inboundmanager "github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/option"
	tun "github.com/sagernet/sing-tun"
	"github.com/sagernet/sing/service"
	"github.com/stretchr/testify/require"
)

type emptyNetworkManager struct {
	adapter.NetworkManager
}

func (emptyNetworkManager) InterfaceMonitor() tun.DefaultInterfaceMonitor {
	return emptyInterfaceMonitor{}
}

type emptyInterfaceMonitor struct {
	tun.DefaultInterfaceMonitor
}

func (emptyInterfaceMonitor) MyInterfaces() []string {
	return nil
}

type emptyDNSTransportManager struct {
	adapter.DNSTransportManager
}

func (emptyDNSTransportManager) Transports() []adapter.DNSTransport {
	return nil
}

func (emptyDNSTransportManager) Default() adapter.DNSTransport {
	return nil
}

func newContextWithTUNPrefix(prefixes ...netip.Prefix) context.Context {
	ctx := service.ContextWithDefaultRegistry(context.Background())
	service.MustRegister[adapter.DNSTransportManager](ctx, emptyDNSTransportManager{})
	inboundManager := inboundmanager.NewManager(nil, nil, nil)
	inboundManager.SetTunAddressPrefix(prefixes)
	service.MustRegister[adapter.InboundManager](ctx, inboundManager)
	return ctx
}

func TestNewOutboundUsesConfiguredTUNPrefixBeforeInterfaceStart(t *testing.T) {
	ctx := newContextWithTUNPrefix(netip.MustParsePrefix("10.20.0.1/30"))

	createdOutbound, err := NewOutbound(ctx, nil, nil, "direct", option.DirectOutboundOptions{})
	require.NoError(t, err)
	outbound := createdOutbound.(*Outbound)

	require.True(t, outbound.isMyLoopbackAddress(netip.MustParseAddr("10.20.0.2")))
	require.False(t, outbound.isMyLoopbackAddress(netip.MustParseAddr("192.168.1.254")))
}

func TestFetchMyAddressesUsesConfiguredTUNPrefixBeforeInterfaceStart(t *testing.T) {
	ctx := newContextWithTUNPrefix(netip.MustParsePrefix("10.20.0.1/30"))

	outbound := &Outbound{
		ctx:     ctx,
		network: emptyNetworkManager{},
	}
	outbound.fetchMyAddresses()

	require.True(t, outbound.isMyLoopbackAddress(netip.MustParseAddr("10.20.0.2")))
	require.False(t, outbound.isMyLoopbackAddress(netip.MustParseAddr("192.168.1.254")))
}

func TestFetchMyAddressesWithoutConfiguredTUNPrefix(t *testing.T) {
	outbound := &Outbound{
		ctx:     context.Background(),
		network: emptyNetworkManager{},
	}
	outbound.fetchMyAddresses()

	require.False(t, outbound.isMyLoopbackAddress(netip.MustParseAddr("10.20.0.2")))
}
