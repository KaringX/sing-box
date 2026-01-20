// karing
package clashapi

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/sagernet/sing-box/adapter"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

func httpRequestByProxy(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		url := query.Get("url")
		timeout, err := strconv.ParseInt(query.Get("timeout"), 10, 32)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrBadRequest)
			return
		}

		proxy := r.Context().Value(CtxKeyProxy).(adapter.Outbound)
		ctx, cancel := context.WithTimeout(server.ctx, time.Millisecond*time.Duration(timeout))
		defer cancel()

		statusCode, header, body, err := URLRequest(ctx, url, proxy)

		if ctx.Err() != nil {
			//render.Status(r, http.StatusGatewayTimeout)
			//render.JSON(w, r, ErrRequestTimeout)
			render.JSON(w, r, newError(ctx.Err().Error()))
			return
		}

		if err != nil {
			//render.Status(r, http.StatusServiceUnavailable)
			//render.JSON(w, r, newError("An error occurred in the delay test"))
			render.JSON(w, r, newError(err.Error()))
			return
		}

		render.JSON(w, r, render.M{
			"status_code": statusCode,
			"header":      header,
			"body":        body,
		})
	}
}

func getProxyDelayHistory(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.Context().Value(CtxKeyProxyName).(string)
		delayHistory := server.urlTestHistory.LoadURLTestHistory(name)
		render.JSON(w, r, render.M{
			"history": delayHistory,
		})
	}
}

func URLRequest(ctx context.Context, link string, detour N.Dialer) (statusCode int, header map[string][]string, content []byte, err error) {
	if link == "" {
		return 0, nil, nil, E.New("request url is empty")
	}
	linkURL, err := url.Parse(link)
	if err != nil {
		return
	}
	hostname := linkURL.Hostname()
	port := linkURL.Port()
	if port == "" {
		switch linkURL.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		}
	}

	instance, err := detour.DialContext(ctx, "tcp", M.ParseSocksaddrHostPortStr(hostname, port))
	if err != nil {
		return
	}
	defer instance.Close()
	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return
	}
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return instance, nil
			},
			//DisableKeepAlives:   true,// karing
			//TLSHandshakeTimeout: C.TCPTimeout,// karing
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	defer client.CloseIdleConnections()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return
	}

	content, err = io.ReadAll(resp.Body)
	resp.Body.Close()

	return resp.StatusCode, resp.Header, content, err
}
