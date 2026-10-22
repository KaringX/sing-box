//go:build !go1.27

package force_close

import (
	"sync"
	"unsafe"

	"golang.org/x/net/http2"
)

func CloseHTTP2Connections(transport *http2.Transport) {
	connPool := transportConnPool(transport)
	p := (*clientConnPool)((*efaceWords)(unsafe.Pointer(&connPool)).data)
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, vv := range p.conns {
		for _, cc := range vv {
			cc.Close()
		}
	}
}

type clientConnPool struct {
	t     *http2.Transport
	mu    sync.Mutex
	conns map[string][]*http2.ClientConn // key is host:port
}

//go:linkname transportConnPool golang.org/x/net/http2.(*Transport).connPool
func transportConnPool(t *http2.Transport) http2.ClientConnPool
