package force_close

import (
	"net/http"
	"reflect"
	"unsafe"

	"github.com/sagernet/quic-go/http3"
	E "github.com/sagernet/sing/common/exceptions"

	"golang.org/x/net/http2"
)

func ResetTransport(rawTransport http.RoundTripper) http.RoundTripper {
	switch transport := rawTransport.(type) {
	case *http.Transport:
		transport.CloseIdleConnections()
		return transport.Clone()
	case *http2.Transport:
		CloseHTTP2Connections(transport)
		return transport
	case *http3.Transport:
		transport.Close()
		return transport
	default:
		panic(E.New("unknown transport type: ", reflect.TypeOf(transport)))
	}
}

type efaceWords struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}
