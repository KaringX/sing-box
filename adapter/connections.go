package adapter

import (
	"context"
	"net"

	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

type ConnectionManager interface {
	Lifecycle
	Count() int
	CloseAll()
	TrackConn(conn net.Conn, destination M.Socksaddr, inbound *InboundContext) net.Conn                   //karing
	TrackPacketConn(conn net.PacketConn, destination M.Socksaddr, inbound *InboundContext) net.PacketConn //karing
	NewConnection(ctx context.Context, this N.Dialer, conn net.Conn, metadata InboundContext, onClose N.CloseHandlerFunc)
	NewPacketConnection(ctx context.Context, this N.Dialer, conn N.PacketConn, metadata InboundContext, onClose N.CloseHandlerFunc)
	Connections() []OutboundContext //karing
}
