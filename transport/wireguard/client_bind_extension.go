// karing
package wireguard

import (
	"net/netip"
	"time"

	"github.com/sagernet/wireguard-go/conn"
)

func (c *ClientBind) SendWithoutModify(bufs [][]byte, ep conn.Endpoint) error { //hiddify
	udpConn, err := c.connect()
	if err != nil {
		c.pauseManager.WaitActive()
		time.Sleep(time.Second)
		return err
	}
	destination := netip.AddrPort(ep.(remoteEndpoint))
	for _, b := range bufs {
		if false && len(b) > 3 { //do not change to reserved
			reserved, loaded := c.reservedForEndpoint[destination]
			if !loaded {
				reserved = c.reserved
			}
			copy(b[1:4], reserved[:])
		}
		_, err = udpConn.WriteToUDPAddrPort(b, destination)
		if err != nil {
			udpConn.Close()
			return err
		}
	}
	return nil
}
