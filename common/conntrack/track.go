package conntrack

import (
	"io"
	"sync"
	"time"

	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/x/list"
)

type OutboundConn struct { //karing
	Closer      io.Closer
	CreatedAt   time.Time
	Network     string
	Destination string
	Outbound    string
}

var (
	connAccess     sync.RWMutex
	openConnection list.List[OutboundConn] //karing
)

func Count() int {
	if !Enabled {
		return 0
	}
	return openConnection.Len()
}

func List() []io.Closer {
	if !Enabled {
		return nil
	}
	connAccess.RLock()
	defer connAccess.RUnlock()
	connList := make([]io.Closer, 0, openConnection.Len())
	for element := openConnection.Front(); element != nil; element = element.Next() {
		connList = append(connList, element.Value.Closer) //karing
	}
	return connList
}

func Connections() []OutboundConn { //karing
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

func Close() {
	if !Enabled {
		return
	}
	connAccess.Lock()
	defer connAccess.Unlock()
	for element := openConnection.Front(); element != nil; element = element.Next() {
		common.Close(element.Value.Closer)
		element.Value.Closer = nil
		element.Value.Outbound = "" //karing
	}
	openConnection.Init()
}
