package sudoku

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/sagernet/sing-box/common/tls"
	"github.com/sagernet/sing-box/transport/sudoku/obfs/httpmask"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

type ClientOptions struct {
	Dialer    N.Dialer
	TLSConfig tls.Config
	Config    ProtocolConfig
}

type Client struct {
	dialer    N.Dialer
	tlsConfig tls.Config
	baseConf  ProtocolConfig

	muxMu     sync.Mutex
	muxClient *MultiplexClient

	httpMaskMu     sync.Mutex
	httpMaskClient *httpmask.TunnelClient
	httpMaskKey    string
}

func NewClient(options ClientOptions) *Client {
	return &Client{
		dialer:    options.Dialer,
		tlsConfig: options.TLSConfig,
		baseConf:  options.Config,
	}
}

func (c *Client) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	cfg := c.baseConf
	cfg.TargetAddress = destination.String()

	muxMode := normalizeHTTPMaskMultiplex(cfg.HTTPMaskMultiplex)
	if muxMode == "on" && !cfg.DisableHTTPMask && httpTunnelModeEnabled(cfg.HTTPMaskMode) {
		stream, err := c.dialMultiplex(ctx, cfg.TargetAddress)
		if err == nil {
			return stream, nil
		}
		return nil, err
	}

	conn, err := c.dialAndHandshake(ctx, &cfg)
	if err != nil {
		return nil, err
	}

	switch N.NetworkName(network) {
	case N.NetworkUDP:
		if err = WriteKIPMessage(conn, KIPTypeStartUoT, nil); err != nil {
			conn.Close()
			return nil, E.Cause(err, "start uot")
		}
		return conn, nil
	default:
		addrBuf, err := EncodeAddress(cfg.TargetAddress)
		if err != nil {
			conn.Close()
			return nil, E.Cause(err, "encode target address")
		}
		if err = WriteKIPMessage(conn, KIPTypeOpenTCP, addrBuf); err != nil {
			conn.Close()
			return nil, E.Cause(err, "send target address")
		}
		return conn, nil
	}
}

func (c *Client) Close() {
	c.resetMuxClient()
	c.resetHTTPMaskClient()
}

func (c *Client) dialAndHandshake(ctx context.Context, cfg *ProtocolConfig) (net.Conn, error) {
	handshakeCfg := *cfg
	if !handshakeCfg.DisableHTTPMask && httpTunnelModeEnabled(handshakeCfg.HTTPMaskMode) {
		handshakeCfg.DisableHTTPMask = true
	}

	upgrade := func(raw net.Conn) (net.Conn, error) {
		return ClientHandshake(raw, &handshakeCfg)
	}

	var conn net.Conn
	var err error
	var handshakeDone bool

	if !cfg.DisableHTTPMask && httpTunnelModeEnabled(cfg.HTTPMaskMode) {
		muxMode := normalizeHTTPMaskMultiplex(cfg.HTTPMaskMultiplex)
		if muxMode == "auto" && strings.ToLower(strings.TrimSpace(cfg.HTTPMaskMode)) != "ws" {
			if client, cerr := c.getOrCreateHTTPMaskClient(cfg); cerr == nil && client != nil {
				conn, err = client.DialTunnel(ctx, httpmask.TunnelDialOptions{
					Mode:         cfg.HTTPMaskMode,
					TLSConfig:    c.httpMaskTLSConfig(),
					HostOverride: cfg.HTTPMaskHost,
					PathRoot:     cfg.HTTPMaskPathRoot,
					AuthKey:      ClientAEADSeed(cfg.Key),
					Upgrade:      upgrade,
					Multiplex:    cfg.HTTPMaskMultiplex,
					DialContext:  c.dialRaw,
				})
				if err != nil {
					c.resetHTTPMaskClient()
				}
			}
		}
		if conn == nil && err == nil {
			conn, err = c.dialHTTPMaskTunnel(ctx, cfg, upgrade)
		}
		if err == nil && conn != nil {
			handshakeDone = true
		}
	}
	if conn == nil && err == nil {
		conn, err = c.dialer.DialContext(ctx, N.NetworkTCP, M.ParseSocksaddr(cfg.ServerAddress))
	}
	if err != nil {
		return nil, E.Cause(err, "connect to ", cfg.ServerAddress)
	}

	if !handshakeDone {
		conn, err = ClientHandshake(conn, &handshakeCfg)
		if err != nil {
			common.Close(conn)
			return nil, err
		}
	}

	return conn, nil
}

func (c *Client) dialRaw(ctx context.Context, network, addr string) (net.Conn, error) {
	return c.dialer.DialContext(ctx, network, M.ParseSocksaddr(addr))
}

func (c *Client) dialHTTPMaskTunnel(ctx context.Context, cfg *ProtocolConfig, upgrade func(net.Conn) (net.Conn, error)) (net.Conn, error) {
	var earlyHandshake *httpmask.ClientEarlyHandshake
	var err error
	if upgrade != nil {
		earlyHandshake, err = newClientHTTPMaskEarlyHandshake(cfg)
		if err != nil {
			return nil, err
		}
	}
	return httpmask.DialTunnel(ctx, cfg.ServerAddress, httpmask.TunnelDialOptions{
		Mode:           cfg.HTTPMaskMode,
		TLSConfig:      c.httpMaskTLSConfig(),
		HostOverride:   cfg.HTTPMaskHost,
		PathRoot:       cfg.HTTPMaskPathRoot,
		AuthKey:        ClientAEADSeed(cfg.Key),
		EarlyHandshake: earlyHandshake,
		Upgrade:        upgrade,
		Multiplex:      cfg.HTTPMaskMultiplex,
		DialContext:    c.dialRaw,
	})
}

func (c *Client) httpMaskTLSConfig() tls.Config {
	if c.tlsConfig == nil {
		return nil
	}
	return c.tlsConfig
}

func (c *Client) dialMultiplex(ctx context.Context, targetAddress string) (net.Conn, error) {
	for attempt := 0; attempt < 2; attempt++ {
		client, err := c.getOrCreateMuxClient(ctx)
		if err != nil {
			return nil, err
		}
		stream, err := client.Dial(ctx, targetAddress)
		if err != nil {
			c.resetMuxClient()
			continue
		}
		return stream, nil
	}
	return nil, fmt.Errorf("multiplex open stream failed")
}

func (c *Client) getOrCreateMuxClient(ctx context.Context) (*MultiplexClient, error) {
	c.muxMu.Lock()
	defer c.muxMu.Unlock()

	if c.muxClient != nil && !c.muxClient.IsClosed() {
		return c.muxClient, nil
	}

	baseCfg := c.baseConf
	baseConn, err := c.dialAndHandshake(ctx, &baseCfg)
	if err != nil {
		return nil, err
	}

	client, err := StartMultiplexClient(baseConn)
	if err != nil {
		baseConn.Close()
		return nil, err
	}
	c.muxClient = client
	return client, nil
}

func (c *Client) resetMuxClient() {
	c.muxMu.Lock()
	defer c.muxMu.Unlock()
	if c.muxClient != nil {
		c.muxClient.Close()
		c.muxClient = nil
	}
}

func (c *Client) getOrCreateHTTPMaskClient(cfg *ProtocolConfig) (*httpmask.TunnelClient, error) {
	key := cfg.ServerAddress + "|" + fmt.Sprint(c.tlsConfig != nil) + "|" + strings.TrimSpace(cfg.HTTPMaskHost)

	c.httpMaskMu.Lock()
	if c.httpMaskClient != nil && c.httpMaskKey == key {
		client := c.httpMaskClient
		c.httpMaskMu.Unlock()
		return client, nil
	}
	c.httpMaskMu.Unlock()

	client, err := httpmask.NewTunnelClient(cfg.ServerAddress, httpmask.TunnelClientOptions{
		TLSConfig:    c.httpMaskTLSConfig(),
		HostOverride: cfg.HTTPMaskHost,
		DialContext:  c.dialRaw,
		MaxIdleConns: 32,
	})
	if err != nil {
		return nil, err
	}

	c.httpMaskMu.Lock()
	defer c.httpMaskMu.Unlock()
	if c.httpMaskClient != nil && c.httpMaskKey == key {
		client.CloseIdleConnections()
		return c.httpMaskClient, nil
	}
	if c.httpMaskClient != nil {
		c.httpMaskClient.CloseIdleConnections()
	}
	c.httpMaskClient = client
	c.httpMaskKey = key
	return client, nil
}

func (c *Client) resetHTTPMaskClient() {
	c.httpMaskMu.Lock()
	defer c.httpMaskMu.Unlock()
	if c.httpMaskClient != nil {
		c.httpMaskClient.CloseIdleConnections()
		c.httpMaskClient = nil
		c.httpMaskKey = ""
	}
}

func normalizeHTTPMaskMultiplex(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "off":
		return "off"
	case "auto":
		return "auto"
	case "on":
		return "on"
	default:
		return "off"
	}
}

func httpTunnelModeEnabled(mode string) bool {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "stream", "poll", "auto", "ws":
		return true
	default:
		return false
	}
}
