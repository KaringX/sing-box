package conntrack

import (
	"net"
	"time"

	"github.com/sagernet/sing-box/adapter"
	M "github.com/sagernet/sing/common/metadata"
	"github.com/sagernet/sing/common/x/list"
)

type Conn struct {
	net.Conn
	element *list.Element[OutboundConn] //karing
}

func NewConn(conn net.Conn, destination M.Socksaddr, inbound *adapter.InboundContext) (net.Conn, error) { //karing
	connAccess.Lock()
	warpper := OutboundConn { //karing
		Closer:      conn, 
		CreatedAt:   time.Now(), 
		Network:     "tcp", 
		Source:      inbound.Source,
		Destination: destination,
		Fqdn:        inbound.Destination.Fqdn, 
		Outbound:    inbound.Outbound,
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
	return &Conn{
		Conn:    conn,
		element: element,
	}, nil
}

func (c *Conn) Close() error {
	if c.element.Value.Closer != nil { //karing
		connAccess.Lock()
		if c.element.Value.Closer != nil { //karing
			openConnection.Remove(c.element)
			c.element.Value.Closer = nil //karing
		}
		connAccess.Unlock()
	}
	return c.Conn.Close()
}

func (c *Conn) Upstream() any {
	return c.Conn
}

func (c *Conn) ReaderReplaceable() bool {
	return true
}

func (c *Conn) WriterReplaceable() bool {
	return true
}
