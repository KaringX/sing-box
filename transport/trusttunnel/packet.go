package trusttunnel

import (
	"encoding/binary"
	"math"
	"net"

	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/buf"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/rw"
)

var (
	_ N.NetPacketConn = (*clientPacketConn)(nil)
	_ N.FrontHeadroom = (*clientPacketConn)(nil)
)

type clientPacketConn struct {
	httpConn
}

func (u *clientPacketConn) FrontHeadroom() int {
	return 4 + 16 + 2 + 16 + 2 + 1 + math.MaxUint8
}

func (u *clientPacketConn) ReadPacket(buffer *buf.Buffer) (destination M.Socksaddr, err error) {
	err = u.waitCreated()
	if err != nil {
		return M.Socksaddr{}, err
	}
	return u.readPacketFromServer(buffer)
}

func (u *clientPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	buffer := buf.With(p)
	destination, err := u.ReadPacket(buffer)
	if err != nil {
		return 0, nil, err
	}
	return buffer.Len(), destination.UDPAddr(), nil
}

func (u *clientPacketConn) WritePacket(buffer *buf.Buffer, destination M.Socksaddr) error {
	return u.writePacketToServer(buffer, destination)
}

func (u *clientPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	err = u.WritePacket(buf.As(p), M.SocksaddrFromNet(addr))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (u *clientPacketConn) readPacketFromServer(buffer *buf.Buffer) (destination M.Socksaddr, err error) {
	header := buf.NewSize(4 + 16 + 2 + 16 + 2)
	defer header.Release()
	_, err = header.ReadFullFrom(u.body, header.FreeLen())
	if err != nil {
		return
	}
	var length uint32
	common.Must(binary.Read(header, binary.BigEndian, &length))
	var sourceAddressBuffer [16]byte
	common.Must1(header.Read(sourceAddressBuffer[:]))
	destination.Addr = parse16BytesIP(sourceAddressBuffer)
	common.Must(binary.Read(header, binary.BigEndian, &destination.Port))
	common.Must(rw.SkipN(header, 16+2))
	payloadLen := int(length) - (16 + 2 + 16 + 2)
	if payloadLen < 0 {
		return M.Socksaddr{}, E.New("invalid udp length: ", length)
	}
	_, err = buffer.ReadFullFrom(u.body, payloadLen)
	return
}

func (u *clientPacketConn) writePacketToServer(buffer *buf.Buffer, source M.Socksaddr) error {
	defer buffer.Release()
	if !source.IsIP() {
		return E.New("only support IP")
	}
	payloadLen := buffer.Len()
	headerLen := 4 + 16 + 2 + 16 + 2 + 1 + len(appName)
	lengthField := uint32(16 + 2 + 16 + 2 + 1 + len(appName) + payloadLen)
	destinationAddress := buildPaddingIP(source.Addr)
	header := buf.NewSize(headerLen)
	defer header.Release()
	common.Must(binary.Write(header, binary.BigEndian, lengthField))
	common.Must(header.WriteZeroN(16 + 2))
	common.Must1(header.Write(destinationAddress[:]))
	common.Must(binary.Write(header, binary.BigEndian, source.Port))
	common.Must(binary.Write(header, binary.BigEndian, uint8(len(appName))))
	common.Must1(header.WriteString(appName))
	return u.writeChunks(header.Bytes(), buffer.Bytes())
}

var (
	_ N.NetPacketConn = (*serverPacketConn)(nil)
	_ N.FrontHeadroom = (*serverPacketConn)(nil)
)

type serverPacketConn struct {
	httpConn
}

func (u *serverPacketConn) FrontHeadroom() int {
	return 4 + 16 + 2 + 16 + 2
}

func (u *serverPacketConn) ReadPacket(buffer *buf.Buffer) (destination M.Socksaddr, err error) {
	err = u.waitCreated()
	if err != nil {
		return M.Socksaddr{}, err
	}
	return u.readPacketFromClient(buffer)
}

func (u *serverPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	buffer := buf.With(p)
	destination, err := u.ReadPacket(buffer)
	if err != nil {
		return 0, nil, err
	}
	return buffer.Len(), destination.UDPAddr(), nil
}

func (u *serverPacketConn) WritePacket(buffer *buf.Buffer, destination M.Socksaddr) error {
	return u.writePacketToClient(buffer, destination)
}

func (u *serverPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	err = u.WritePacket(buf.As(p), M.SocksaddrFromNet(addr))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (u *serverPacketConn) readPacketFromClient(buffer *buf.Buffer) (destination M.Socksaddr, err error) {
	header := buf.NewSize(4 + 16 + 2 + 16 + 2 + 1)
	defer header.Release()
	_, err = header.ReadFullFrom(u.body, header.FreeLen())
	if err != nil {
		return
	}
	var length uint32
	common.Must(binary.Read(header, binary.BigEndian, &length))
	common.Must(rw.SkipN(header, 16+2))
	var destinationAddressBuffer [16]byte
	common.Must1(header.Read(destinationAddressBuffer[:]))
	destination.Addr = parse16BytesIP(destinationAddressBuffer)
	common.Must(binary.Read(header, binary.BigEndian, &destination.Port))
	var appNameLen uint8
	common.Must(binary.Read(header, binary.BigEndian, &appNameLen))
	if appNameLen > 0 {
		err = rw.SkipN(u.body, int(appNameLen))
		if err != nil {
			return M.Socksaddr{}, err
		}
	}
	payloadLen := int(length) - (16 + 2 + 16 + 2 + 1 + int(appNameLen))
	if payloadLen < 0 {
		return M.Socksaddr{}, E.New("invalid udp length: ", length)
	}
	_, err = buffer.ReadFullFrom(u.body, payloadLen)
	return
}

func (u *serverPacketConn) writePacketToClient(buffer *buf.Buffer, source M.Socksaddr) error {
	defer buffer.Release()
	if !source.IsIP() {
		return E.New("only support IP")
	}
	payloadLen := buffer.Len()
	headerLen := 4 + 16 + 2 + 16 + 2
	lengthField := uint32(16 + 2 + 16 + 2 + payloadLen)
	sourceAddress := buildPaddingIP(source.Addr)
	header := buf.NewSize(headerLen)
	defer header.Release()
	common.Must(binary.Write(header, binary.BigEndian, lengthField))
	common.Must1(header.Write(sourceAddress[:]))
	common.Must(binary.Write(header, binary.BigEndian, source.Port))
	common.Must(header.WriteZeroN(16 + 2))
	return u.writeChunks(header.Bytes(), buffer.Bytes())
}
