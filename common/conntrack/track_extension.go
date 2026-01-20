// karing
package conntrack

import (
	"io"
	"time"

	M "github.com/sagernet/sing/common/metadata"
)

type OutboundConn struct { //karing
	Closer      io.Closer
	CreatedAt   time.Time
	Network     string
	Source      M.Socksaddr
	Destination M.Socksaddr
	Fqdn        string
	Outbound    string
}

func Connections() []OutboundConn {
	if !Enabled {
		return nil
	}
	connAccess.RLock()
	defer connAccess.RUnlock()
	connList := make([]OutboundConn, 0, openConnection.Len())
	for element := openConnection.Front(); element != nil; element = element.Next() {
		connList = append(connList, element.Value)
	}
	return connList
}
