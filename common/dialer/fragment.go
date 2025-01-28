package dialer

//hiddify
import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"time"

	opt "github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/bufio"
	M "github.com/sagernet/sing/common/metadata"
)

type TLSFragment struct {
	Enabled bool
	Sleep   opt.IntRange
	Size    opt.IntRange
}

type fragmentConn struct {
	dialer      net.Dialer
	fragment    TLSFragment
	network     string
	destination M.Socksaddr
	conn        net.Conn
	err         error
}

// isClientHelloPacket checks if data resembles a TLS clientHello packet
func isClientHelloPacket(b []byte) bool {
	// Check if the packet is at least 5 bytes long and the content type is 22 (TLS handshake)
	if len(b) < 5 || b[0] != 22 {
		return false
	}

	// Check if the protocol version is TLS 1.0 or higher (0x0301 or greater)
	version := uint16(b[1])<<8 | uint16(b[2])
	if version < 0x0301 {
		return false
	}

	// Check if the handshake message type is ClientHello (1)
	if b[5] != 1 {
		return false
	}

	return true
}

// parseSniInfo parses the ClientHello message and extracts the SNI info
func parseSniInfo(data []byte) (sni string, startIndex int, length int, err error) {
	// Skip the first 5 bytes of the TLS record header
	data = data[5:]

	messageLen := int(data[1])<<16 | int(data[2])<<8 | int(data[3])
	if len(data) < 4+messageLen {
		return "", 0, 0, errors.New("data too short for complete handshake message")
	}

	// Skip the handshake message header
	data = data[4 : 4+messageLen]

	if len(data) < 34 {
		return "", 0, 0, errors.New("data too short for ClientHello fixed part")
	}

	sessionIDLen := int(data[34])
	offset := 35 + sessionIDLen

	if len(data) < offset+2 {
		return "", 0, 0, errors.New("data too short for cipher suites length")
	}

	cipherSuitesLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2 + cipherSuitesLen

	if len(data) < offset+1 {
		return "", 0, 0, errors.New("data too short for compression methods length")
	}

	compressionMethodsLen := int(data[offset])
	offset += 1 + compressionMethodsLen

	if len(data) < offset+2 {
		return "", 0, 0, errors.New("data too short for extensions length")
	}

	extensionsLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2

	if len(data) < offset+extensionsLen {
		return "", 0, 0, errors.New("data too short for complete extensions")
	}

	extensions := data[offset : offset+extensionsLen]
	for len(extensions) >= 4 {
		extType := binary.BigEndian.Uint16(extensions[:2])
		extLen := int(binary.BigEndian.Uint16(extensions[2:4]))
		if len(extensions) < 4+extLen {
			return "", 0, 0, errors.New("extension length mismatch")
		}
		if extType == 0x00 { // SNI extension
			sniData := extensions[4 : 4+extLen]
			if len(sniData) < 2 {
				return "", 0, 0, errors.New("invalid SNI extension data")
			}
			serverNameListLen := int(binary.BigEndian.Uint16(sniData[:2]))
			if len(sniData) < 2+serverNameListLen {
				return "", 0, 0, errors.New("SNI list length mismatch")
			}
			serverNameList := sniData[2 : 2+serverNameListLen]
			for len(serverNameList) >= 3 {
				nameType := serverNameList[0]
				nameLen := int(binary.BigEndian.Uint16(serverNameList[1:3]))
				if len(serverNameList) < 3+nameLen {
					return "", 0, 0, errors.New("server name length mismatch")
				}
				if nameType == 0 { // host_name
					sni = string(serverNameList[3 : 3+nameLen])
					startIndex = offset + 4 + 2 + 3
					length = nameLen
					return sni, startIndex, length, nil
				}
				serverNameList = serverNameList[3+nameLen:]
			}
		}
		extensions = extensions[4+extLen:]
	}

	return "", 0, 0, errors.New("SNI not found")
}

// selectRandomIndices selects random indices to chunk data into fragments based on a given range
func selectRandomIndices(dataLen int, sizeRange opt.IntRange) []int {
	var indices []int

	for current := 0; current < dataLen; {
		// Ensure the chunk size does not exceed the remaining length
		chunkSize := int(sizeRange.UniformRand())
		if current+chunkSize > dataLen {
			chunkSize = dataLen - current
		}

		current += chunkSize
		indices = append(indices, current)
	}

	return indices
}

func fragmentTLSClientHello(b []byte, sizeRange opt.IntRange) [][]byte {
	var fragments [][]byte
	var fragmentIndices []int
	clientHelloLen := int(binary.BigEndian.Uint16(b[3:5]))
	clientHelloData := b[5:]

	_, sniStartIdx, sniLen, err := parseSniInfo(b)
	if err != nil {
		fragmentIndices = selectRandomIndices(clientHelloLen, sizeRange)
	} else {
		// select random indices in two parts, 0-randomIndexOfSni and randomIndexOfSni-packetEnd, ensuring the SNI ext is fragmented
		sniExtFragmentIdx := opt.IntRange{Min: uint64(sniStartIdx), Max: uint64(sniStartIdx + sniLen)}.UniformRand()
		preSniExtIdx := selectRandomIndices(int(sniExtFragmentIdx), sizeRange)
		postSniExtIdx := selectRandomIndices(clientHelloLen-(sniStartIdx+sniLen), sizeRange)
		for i := range postSniExtIdx {
			postSniExtIdx[i] += sniStartIdx + sniLen
		}
		fragmentIndices = append(fragmentIndices, preSniExtIdx...)
		fragmentIndices = append(fragmentIndices, postSniExtIdx...)
	}

	fragmentStart := 0
	for _, fragmentEnd := range fragmentIndices {
		header := make([]byte, 5)
		header[0] = b[0]
		binary.BigEndian.PutUint16(header[1:], binary.BigEndian.Uint16(b[1:3]))
		binary.BigEndian.PutUint16(header[3:], uint16(fragmentEnd-fragmentStart))
		payload := append(header, clientHelloData[fragmentStart:fragmentEnd]...)
		fragments = append(fragments, payload)
		fragmentStart = fragmentEnd
	}

	return fragments
}

func (c *fragmentConn) writeFragments(fragments [][]byte) (n int, err error) {
	var totalWrittenBytes int
	for _, fragment := range fragments {
		lastWrittenBytes, err := c.conn.Write(fragment)
		if err != nil {
			c.err = err
			return totalWrittenBytes, c.err
		}
		totalWrittenBytes += lastWrittenBytes

		if c.fragment.Sleep.Max != 0 {
			time.Sleep(time.Duration(c.fragment.Sleep.UniformRand()) * time.Millisecond)
		}
	}
	return totalWrittenBytes, nil
}

func (c *fragmentConn) Write(b []byte) (n int, err error) {
	if c.conn == nil {
		return 0, c.err
	}

	if isClientHelloPacket(b) {
		fragments := fragmentTLSClientHello(b, c.fragment.Size)
		return c.writeFragments(fragments)
	}

	return c.conn.Write(b)
}

func (c *fragmentConn) Read(b []byte) (n int, err error) {
	if c.conn == nil {
		return 0, c.err
	}
	return c.conn.Read(b)
}

func (c *fragmentConn) Close() error {
	return common.Close(c.conn)
}

func (c *fragmentConn) LocalAddr() net.Addr {
	if c.conn == nil {
		return M.Socksaddr{}
	}
	return c.conn.LocalAddr()
}

func (c *fragmentConn) RemoteAddr() net.Addr {
	if c.conn == nil {
		return M.Socksaddr{}
	}
	return c.conn.RemoteAddr()
}

func (c *fragmentConn) SetDeadline(t time.Time) error {
	if c.conn == nil {
		return os.ErrInvalid
	}
	return c.conn.SetDeadline(t)
}

func (c *fragmentConn) SetReadDeadline(t time.Time) error {
	if c.conn == nil {
		return os.ErrInvalid
	}
	return c.conn.SetReadDeadline(t)
}

func (c *fragmentConn) SetWriteDeadline(t time.Time) error {
	if c.conn == nil {
		return os.ErrInvalid
	}
	return c.conn.SetWriteDeadline(t)
}

func (c *fragmentConn) Upstream() any {
	return c.conn
}

func (c *fragmentConn) ReaderReplaceable() bool {
	return c.conn != nil
}

func (c *fragmentConn) WriterReplaceable() bool {
	return c.conn != nil
}

func (c *fragmentConn) LazyHeadroom() bool {
	return c.conn == nil
}

func (c *fragmentConn) NeedHandshake() bool {
	return c.conn == nil
}

func (c *fragmentConn) WriteTo(w io.Writer) (n int64, err error) {
	if c.conn == nil {
		return 0, c.err
	}
	return bufio.Copy(w, c.conn)
}
