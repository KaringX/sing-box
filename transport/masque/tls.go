package masque

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/sagernet/quic-go"
	"github.com/sagernet/quic-go/http3"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	aTLS "github.com/sagernet/sing/common/tls"
)

type TLSConfig struct {
	aTLS.Config
	connectionIDLength int
}

func NewTLSConfig(config aTLS.Config, connectionIDLength int) aTLS.Config {
	return &TLSConfig{config, connectionIDLength}
}

func (c *TLSConfig) Clone() aTLS.Config {
	return NewTLSConfig(c.Config.Clone(), c.connectionIDLength)
}

func (c *TLSConfig) Dial(ctx context.Context, conn net.Conn, quicConfig *quic.Config) (*quic.Conn, error) {
	return c.dial(ctx, conn, quicConfig, false)
}

func (c *TLSConfig) DialEarly(ctx context.Context, conn net.Conn, quicConfig *quic.Config) (*quic.Conn, error) {
	return c.dial(ctx, conn, quicConfig, true)
}

func (c *TLSConfig) CreateTransport(conn net.Conn, quicConnPtr **quic.Conn, quicConfig *quic.Config) http.RoundTripper {
	packetConn, isPacketConn := common.Cast[net.PacketConn](conn)
	if !isPacketConn {
		return roundTripperError{E.New("masque: underlying connection does not support a custom QUIC transport")}
	}
	return c.CreatePacketTransport(packetConn, conn.RemoteAddr(), quicConnPtr, quicConfig)
}

func (c *TLSConfig) CreatePacketTransport(packetConn net.PacketConn, remoteAddr net.Addr, quicConnPtr **quic.Conn, quicConfig *quic.Config) http.RoundTripper {
	stdConfig, err := c.STDConfig()
	if err != nil {
		return roundTripperError{err}
	}
	transport := c.newTransport(packetConn)
	return &http3.Transport{
		TLSClientConfig: stdConfig,
		QUICConfig:      quicConfig,
		Dial: func(ctx context.Context, addr string, tlsConfig *tls.Config, config *quic.Config) (*quic.Conn, error) {
			quicConn, dialErr := transport.DialEarly(ctx, remoteAddr, tlsConfig, config)
			if dialErr != nil {
				return nil, dialErr
			}
			*quicConnPtr = quicConn
			return quicConn, nil
		},
	}
}

func (c *TLSConfig) dial(ctx context.Context, conn net.Conn, quicConfig *quic.Config, early bool) (*quic.Conn, error) {
	packetConn, isPacketConn := common.Cast[net.PacketConn](conn)
	if !isPacketConn {
		return nil, E.New("masque: underlying connection does not support a custom QUIC transport")
	}
	stdConfig, err := c.STDConfig()
	if err != nil {
		return nil, err
	}
	transport := c.newTransport(packetConn)
	var quicConn *quic.Conn
	if early {
		quicConn, err = transport.DialEarly(ctx, conn.RemoteAddr(), stdConfig, quicConfig)
	} else {
		quicConn, err = transport.Dial(ctx, conn.RemoteAddr(), stdConfig, quicConfig)
	}
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	return quicConn, nil
}

func (c *TLSConfig) newTransport(packetConn net.PacketConn) *quic.Transport {
	transport := &quic.Transport{
		Conn:               packetConn,
		ConnectionIDLength: c.connectionIDLength,
	}
	transport.SetSingleUse(true)
	return transport
}

type roundTripperError struct {
	err error
}

func (r roundTripperError) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, r.err
}
