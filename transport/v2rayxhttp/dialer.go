package xhttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"github.com/sagernet/quic-go/http3"
	"github.com/sagernet/sing-box/common/vision"
	common "github.com/sagernet/sing-box/common/xray"
	"github.com/sagernet/sing-box/common/xray/buf"
	"github.com/sagernet/sing-box/common/xray/signal/done"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	"golang.org/x/net/http2"
)

// interface to abstract between use of browser dialer, vs net/http
type DialerClient interface {
	IsClosed() bool
	Close()

	OpenStream(context.Context, string, string, io.Reader, bool) (io.ReadCloser, net.Addr, net.Addr, error)
	PostPacket(context.Context, string, string, string, buf.MultiBuffer) error
}

// implements xhttp.DialerClient in terms of direct network connections
type DefaultDialerClient struct {
	options     *option.V2RayXHTTPBaseOptions
	client      *http.Client
	closed      bool
	httpVersion string
	// pool of net.Conn, created using dialUploadConn
	uploadRawPool  *sync.Pool
	dialUploadConn func(ctxInner context.Context) (net.Conn, error)

	mtx sync.RWMutex
}

type clientConnPool struct {
	t     *http2.Transport
	mu    sync.Mutex
	conns map[string][]*http2.ClientConn
}

type efaceWords struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

//go:linkname transportConnPool golang.org/x/net/http2.(*Transport).connPool
func transportConnPool(t *http2.Transport) http2.ClientConnPool

func (c *DefaultDialerClient) Close() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	switch transport := c.client.Transport.(type) {
	case *http.Transport:
		transport.CloseIdleConnections()
	case *http2.Transport:
		connPool := transportConnPool(transport)
		p := (*clientConnPool)((*efaceWords)(unsafe.Pointer(&connPool)).data)
		p.mu.Lock()
		defer p.mu.Unlock()
		for _, vv := range p.conns {
			for _, cc := range vv {
				cc.Close()
			}
		}
	case *http3.Transport:
		transport.Close()
	default:
		panic(E.New("unknown transport type: ", reflect.TypeOf(transport)))
	}
}

func (c *DefaultDialerClient) IsClosed() bool {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.closed
}

func (c *DefaultDialerClient) OpenStream(ctx context.Context, url string, sessionId string, body io.Reader, uploadOnly bool) (wrc io.ReadCloser, remoteAddr, localAddr net.Addr, err error) {
	gotConn := done.New()
	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			remoteAddr = connInfo.Conn.RemoteAddr()
			localAddr = connInfo.Conn.LocalAddr()
			if hook, ok := vision.HookFromContext(ctx); ok {
				hook(connInfo.Conn)
			}
			gotConn.Close()
		},
	})
	method := "GET"
	if body != nil {
		method = c.options.GetNormalizedUplinkHTTPMethod()
	}
	req, err := http.NewRequestWithContext(context.WithoutCancel(ctx), method, url, body) //karing
	if err != nil {                                                                       //karing
		return
	}
	FillStreamRequest(req, sessionId, "", c.options)
	wrc = &WaitReadCloser{Wait: make(chan struct{})}
	go func() {
		resp, err := c.client.Do(req)
		if err != nil {
			if !uploadOnly {
				c.Close()
			}
			gotConn.Close()
			common.Close(body)
			wrc.Close()
			return
		}
		if resp.StatusCode != 200 || uploadOnly {
			if resp.StatusCode != 200 {
				c.Close()
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			common.Close(body)
			wrc.Close()
			return
		}
		wrc.(*WaitReadCloser).Set(resp.Body)
	}()
	<-gotConn.Wait()
	return
}

func (c *DefaultDialerClient) PostPacket(ctx context.Context, url string, sessionId string, seqStr string, payload buf.MultiBuffer) error {
	method := c.options.GetNormalizedUplinkHTTPMethod()
	req, err := http.NewRequestWithContext(context.WithoutCancel(ctx), method, url, nil)
	if err != nil {
		return err
	}
	FillPacketRequest(req, sessionId, seqStr, payload, c.options)
	if c.httpVersion != "1.1" {
		resp, err := c.client.Do(req)
		if err != nil {
			c.Close()
			return err
		}
		_, copyErr := io.Copy(io.Discard, resp.Body)
		closeErr := resp.Body.Close()
		if resp.StatusCode != 200 {
			c.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			return E.New("bad status code: ", resp.Status)
		}
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	} else {
		requestBuff := new(bytes.Buffer)
		requestBuff.Grow(512 + int(req.ContentLength))
		common.Must(req.Write(requestBuff))
		var uploadConn any
		var h1UploadConn *H1Conn
		for {
			uploadConn = c.uploadRawPool.Get()
			newConnection := uploadConn == nil
			if newConnection {
				newConn, err := c.dialUploadConn(context.WithoutCancel(ctx))
				if err != nil {
					return err
				}
				h1UploadConn = NewH1Conn(newConn)
				uploadConn = h1UploadConn
			} else {
				h1UploadConn = uploadConn.(*H1Conn)
				if h1UploadConn.UnreadedResponsesCount > 0 {
					resp, err := http.ReadResponse(h1UploadConn.RespBufReader, req)
					if err != nil {
						c.Close()
						return fmt.Errorf("error while reading response: %s", err.Error())
					}
					_, copyErr := io.Copy(io.Discard, resp.Body)
					closeErr := resp.Body.Close()
					if resp.StatusCode != 200 {
						c.Close()
						return fmt.Errorf("got non-200 error response code: %d", resp.StatusCode)
					}
					if copyErr != nil {
						return copyErr
					}
					if closeErr != nil {
						return closeErr
					}
				}
			}
			_, err := h1UploadConn.Write(requestBuff.Bytes())
			if err == nil {
				break
			} else if newConnection {
				return err
			}
		}
		c.uploadRawPool.Put(uploadConn)
	}
	return nil
}

type WaitReadCloser struct {
	Wait chan struct{}
	io.ReadCloser
	mu     sync.Mutex
	once   sync.Once
	closed bool
}

func (w *WaitReadCloser) notify() {
	w.once.Do(func() {
		close(w.Wait)
	})
}

func (w *WaitReadCloser) Set(rc io.ReadCloser) {
	w.mu.Lock()
	if w.closed || w.ReadCloser != nil {
		w.mu.Unlock()
		rc.Close()
		return
	}
	w.ReadCloser = rc
	w.mu.Unlock()
	w.notify()
}

func (w *WaitReadCloser) Read(b []byte) (int, error) {
	w.mu.Lock()
	rc := w.ReadCloser
	w.mu.Unlock()

	if rc == nil {
		<-w.Wait
		w.mu.Lock()
		rc = w.ReadCloser
		w.mu.Unlock()
		if rc == nil {
			return 0, io.ErrClosedPipe
		}
	}
	return rc.Read(b)
}

func (w *WaitReadCloser) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	rc := w.ReadCloser
	w.ReadCloser = nil
	w.mu.Unlock()

	if rc != nil {
		return rc.Close()
	}

	w.notify()
	return nil
}

func ApplyMetaToRequest(options *option.V2RayXHTTPBaseOptions, req *http.Request, sessionId string, seqStr string) {
	sessionPlacement := options.GetNormalizedSessionPlacement()
	seqPlacement := options.GetNormalizedSeqPlacement()
	sessionKey := options.GetNormalizedSessionKey()
	seqKey := options.GetNormalizedSeqKey()
	if sessionId != "" {
		switch sessionPlacement {
		case option.PlacementPath:
			req.URL.Path = appendToPath(req.URL.Path, sessionId)
		case option.PlacementQuery:
			q := req.URL.Query()
			q.Set(sessionKey, sessionId)
			req.URL.RawQuery = q.Encode()
		case option.PlacementHeader:
			req.Header.Set(sessionKey, sessionId)
		case option.PlacementCookie:
			req.AddCookie(&http.Cookie{Name: sessionKey, Value: sessionId})
		}
	}
	if seqStr != "" {
		switch seqPlacement {
		case option.PlacementPath:
			req.URL.Path = appendToPath(req.URL.Path, seqStr)
		case option.PlacementQuery:
			q := req.URL.Query()
			q.Set(seqKey, seqStr)
			req.URL.RawQuery = q.Encode()
		case option.PlacementHeader:
			req.Header.Set(seqKey, seqStr)
		case option.PlacementCookie:
			req.AddCookie(&http.Cookie{Name: seqKey, Value: seqStr})
		}
	}
}

func appendToPath(path, value string) string {
	if strings.HasSuffix(path, "/") {
		return path + value
	}
	return path + "/" + value
}
