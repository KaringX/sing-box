// karing
package endpoint

import (
	"net"
	"time"
)

func (h *Adapter) SetParseErr(err error) {
	h.parseErr = err
}

func (h *Adapter) GetParseErr() error {
	return h.parseErr
}

func (h *Adapter) Connections() int {
	return int(h.ConnectionsIn.Load())
}

func (h *Adapter) OnNewConnection(con net.Conn) net.Conn {
	h.ConnectionsIn.Add(1)
	return &tcpConnectionCloseAdapter{Conn: con, OnClosed: func() {
		h.ConnectionsIn.Add(-1)
	}}
}

func (h *Adapter) OnNewPacketConnection(con net.PacketConn) net.PacketConn {
	h.ConnectionsIn.Add(1)
	return &udpConnectionCloseAdapter{Conn: con, OnClosed: func() {
		h.ConnectionsIn.Add(-1)
	}}
}

type tcpConnectionCloseAdapter struct {
	Conn     net.Conn
	OnClosed func()
}

func (a *tcpConnectionCloseAdapter) Read(b []byte) (n int, err error) {
	return a.Conn.Read(b)
}

func (a *tcpConnectionCloseAdapter) Write(b []byte) (n int, err error) {
	return a.Conn.Write(b)
}

func (a *tcpConnectionCloseAdapter) Close() error {
	err := a.Conn.Close()
	if a.OnClosed != nil {
		a.OnClosed()
		a.OnClosed = nil
	}
	return err
}

func (a *tcpConnectionCloseAdapter) LocalAddr() net.Addr {
	return a.Conn.LocalAddr()
}

func (a *tcpConnectionCloseAdapter) RemoteAddr() net.Addr {
	return a.Conn.RemoteAddr()
}

func (a *tcpConnectionCloseAdapter) SetDeadline(t time.Time) error {
	return a.Conn.SetDeadline(t)
}

func (a *tcpConnectionCloseAdapter) SetReadDeadline(t time.Time) error {
	return a.Conn.SetReadDeadline(t)
}

func (a *tcpConnectionCloseAdapter) SetWriteDeadline(t time.Time) error {
	return a.Conn.SetWriteDeadline(t)
}

type udpConnectionCloseAdapter struct {
	Conn     net.PacketConn
	OnClosed func()
}

func (a *udpConnectionCloseAdapter) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return a.Conn.ReadFrom(p)
}

func (a *udpConnectionCloseAdapter) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return a.Conn.WriteTo(p, addr)
}

func (a *udpConnectionCloseAdapter) Close() error {
	err := a.Conn.Close()
	if a.OnClosed != nil {
		a.OnClosed()
		a.OnClosed = nil
	}
	return err
}

func (a *udpConnectionCloseAdapter) LocalAddr() net.Addr {
	return a.Conn.LocalAddr()
}

func (a *udpConnectionCloseAdapter) SetDeadline(t time.Time) error {
	return a.Conn.SetDeadline(t)
}

func (a *udpConnectionCloseAdapter) SetReadDeadline(t time.Time) error {
	return a.Conn.SetReadDeadline(t)
}

func (a *udpConnectionCloseAdapter) SetWriteDeadline(t time.Time) error {
	return a.Conn.SetWriteDeadline(t)
}
