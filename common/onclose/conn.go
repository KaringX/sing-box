package onclose

import (
	"net"
	"sync"
)

type CloseHandlerFunc = func()

type Conn struct {
	net.Conn
	onClose func()
	once    sync.Once
}

func NewConn(conn net.Conn, onClose func()) *Conn {
	return &Conn{Conn: conn, onClose: onClose}
}

func (c *Conn) Close() error {
	c.once.Do(c.onClose)
	return c.Conn.Close()
}

type PacketConn struct {
	net.PacketConn
	onClose func()
	once    sync.Once
}

func NewPacketConn(conn net.PacketConn, onClose func()) *PacketConn {
	return &PacketConn{PacketConn: conn, onClose: onClose}
}

func (c *PacketConn) Close() error {
	c.once.Do(c.onClose)
	return c.PacketConn.Close()
}
