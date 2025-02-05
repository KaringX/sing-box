package conntrack

import (
	"net"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing/common/bufio"
	"github.com/sagernet/sing/common/x/list"
)

type PacketConn struct {
	net.PacketConn
	element *list.Element[OutboundConn] //karing
}

func NewPacketConn(conn net.PacketConn, inbound *adapter.InboundContext) (net.PacketConn, error) { //karing
	connAccess.Lock()
	warpper := OutboundConn { //karing
		Closer:conn, 
		CreatedAt:time.Now(), 
		Network: "udp", 
		Destination: inbound.Destination.Fqdn, 
		Outbound: inbound.Outbound,
	}
	element := openConnection.PushBack(warpper) //karing
	connAccess.Unlock()
	if KillerEnabled {
		err := KillerCheck()
		if err != nil {
			conn.Close()
			return nil, err
		}
	}
	return &PacketConn{
		PacketConn: conn,
		element:    element,
	}, nil
}

func (c *PacketConn) Close() error {
	if c.element.Value.Closer != nil { //karing
		connAccess.Lock()
		if c.element.Value.Closer != nil { //karing
			openConnection.Remove(c.element)
			c.element.Value.Closer = nil //karing
		}
		connAccess.Unlock()
	}
	return c.PacketConn.Close()
}

func (c *PacketConn) Upstream() any {
	return bufio.NewPacketConn(c.PacketConn)
}

func (c *PacketConn) ReaderReplaceable() bool {
	return true
}

func (c *PacketConn) WriterReplaceable() bool {
	return true
}
