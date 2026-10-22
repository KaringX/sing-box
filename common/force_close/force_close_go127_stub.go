//go:build go1.27 && !badlinkname

package force_close

import "golang.org/x/net/http2"

func CloseHTTP2Connections(transport *http2.Transport) {
	transport.CloseIdleConnections()
}
