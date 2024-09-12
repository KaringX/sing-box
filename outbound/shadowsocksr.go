//go:build with_shadowsocksr

package outbound

//karing
import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/dialer"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/transport/clashssr/obfs"
	"github.com/sagernet/sing-box/transport/clashssr/protocol"
	"github.com/sagernet/sing/common/bufio"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"

	"github.com/Dreamacro/clash/transport/shadowsocks/core"
	"github.com/Dreamacro/clash/transport/shadowsocks/shadowstream"
	"github.com/Dreamacro/clash/transport/socks5"
)

var _ adapter.Outbound = (*ShadowsocksR)(nil)

type ShadowsocksR struct {
	myOutboundAdapter
	dialer     N.Dialer
	serverAddr M.Socksaddr
	cipher     core.Cipher
	obfs       obfs.Obfs
	protocol   protocol.Protocol
	parseErr   error                //karing
}

func NewShadowsocksR(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.ShadowsocksROutboundOptions) (*ShadowsocksR, error) {
	empty := &ShadowsocksR{ //karing
		myOutboundAdapter: myOutboundAdapter{
			protocol:     C.TypeShadowsocksR,
			network:      options.Network.Build(),
			router:       router,
			logger:       logger,
			tag:          tag,
			dependencies: withDialerDependency(options.DialerOptions),
		},
	}
	//logger.Warn("ShadowsocksR is deprecated, see https://sing-box.sagernet.org/deprecated") // logFactory.Start() 调用之前调用logger会导致崩溃
	outboundDialer, err := dialer.New(router, options.DialerOptions)
	if err != nil {
		return empty, err //karing
	}
	outbound := &ShadowsocksR{
		myOutboundAdapter: myOutboundAdapter{
			protocol:     C.TypeShadowsocksR,
			network:      options.Network.Build(),
			router:       router,
			logger:       logger,
			tag:          tag,
			dependencies: withDialerDependency(options.DialerOptions),
		},
		dialer:     outboundDialer,
		serverAddr: options.ServerOptions.Build(),
	}
	var cipher string
	switch options.Method {
	case "none":
		cipher = "dummy"
	default:
		cipher = options.Method
	}
	outbound.cipher, err = core.PickCipher(cipher, nil, options.Password)
	if err != nil {
		return empty, err //karing
	}
	var (
		ivSize int
		key    []byte
	)
	if cipher == "dummy" {
		ivSize = 0
		key = core.Kdf(options.Password, 16)
	} else {
		streamCipher, ok := outbound.cipher.(*core.StreamCipher)
		if !ok {
			return empty, fmt.Errorf("%s is not none or a supported stream cipher in ssr", cipher) //karing
		}
		ivSize = streamCipher.IVSize()
		key = streamCipher.Key
	}
	obfs, obfsOverhead, err := obfs.PickObfs(options.Obfs, &obfs.Base{
		Host:   options.Server,
		Port:   int(options.ServerPort),
		Key:    key,
		IVSize: ivSize,
		Param:  options.ObfsParam,
	})
	if err != nil {
		return empty, E.Cause(err, "initialize obfs") //karing
	}
	protocol, err := protocol.PickProtocol(options.Protocol, &protocol.Base{
		Key:      key,
		Overhead: obfsOverhead,
		Param:    options.ProtocolParam,
	})
	if err != nil {
		return empty, E.Cause(err, "initialize protocol") //karing
	}
	outbound.obfs = obfs
	outbound.protocol = protocol
	return outbound, nil
}

func (h *ShadowsocksR) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	if(h.parseErr != nil){ //karing
		return nil, h.parseErr
	}
	ctx, metadata := adapter.AppendContext(ctx)
	metadata.Outbound = h.tag
	metadata.Destination = destination
	switch network {
	case N.NetworkTCP:
		h.logger.InfoContext(ctx, "outbound connection to ", destination)
		conn, err := h.dialer.DialContext(ctx, network, h.serverAddr)
		if err != nil {
			return nil, err
		}
		conn = h.cipher.StreamConn(h.obfs.StreamConn(conn))
		writeIv, err := conn.(*shadowstream.Conn).ObtainWriteIV()
		if err != nil {
			conn.Close()
			return nil, err
		}
		conn = h.protocol.StreamConn(conn, writeIv)
		err = M.SocksaddrSerializer.WriteAddrPort(conn, destination)
		if err != nil {
			conn.Close()
			return nil, E.Cause(err, "write request")
		}
		return conn, nil
	case N.NetworkUDP:
		conn, err := h.ListenPacket(ctx, destination)
		if err != nil {
			return nil, err
		}
		return bufio.NewBindPacketConn(conn, destination), nil
	default:
		return nil, E.Extend(N.ErrUnknownNetwork, network)
	}
}

func (h *ShadowsocksR) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if(h.parseErr != nil){ //karing
		return nil, h.parseErr
	}
	ctx, metadata := adapter.AppendContext(ctx)
	metadata.Outbound = h.tag
	metadata.Destination = destination
	h.logger.InfoContext(ctx, "outbound packet connection to ", destination)
	outConn, err := h.dialer.DialContext(ctx, N.NetworkUDP, h.serverAddr)
	if err != nil {
		return nil, err
	}
	packetConn := h.cipher.PacketConn(bufio.NewUnbindPacketConn(outConn))
	packetConn = h.protocol.PacketConn(packetConn)
	packetConn = &ssPacketConn{packetConn, outConn.RemoteAddr()}
	return packetConn, nil
}

func (h *ShadowsocksR) NewConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	if(h.parseErr != nil){ //karing
		return h.parseErr
	}
	return NewConnection(ctx, h, conn, metadata)
}

func (h *ShadowsocksR) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	if(h.parseErr != nil){ //karing
		return h.parseErr
	}
	return NewPacketConnection(ctx, h, conn, metadata)
}
func (h *ShadowsocksR) SetParseErr(err error){ //karing
	h.parseErr = err
}
type ssPacketConn struct {
	net.PacketConn
	rAddr net.Addr
}

func (spc *ssPacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	packet, err := socks5.EncodeUDPPacket(socks5.ParseAddrToSocksAddr(addr), b)
	if err != nil {
		return
	}
	return spc.PacketConn.WriteTo(packet[3:], spc.rAddr)
}

func (spc *ssPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, _, e := spc.PacketConn.ReadFrom(b)
	if e != nil {
		return 0, nil, e
	}

	addr := socks5.SplitAddr(b[:n])
	if addr == nil {
		return 0, nil, errors.New("parse addr error")
	}

	udpAddr := addr.UDPAddr()
	if udpAddr == nil {
		return 0, nil, errors.New("parse addr error")
	}

	copy(b, b[len(addr):])
	return n - len(addr), udpAddr, e
}