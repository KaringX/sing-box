package clashapi

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/urltest"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/outbound"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/json/badjson"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

func proxyRouter(server *Server, router adapter.Router) http.Handler {
	r := chi.NewRouter()
	r.Get("/", getProxies(server, router))

	r.Route("/{name}", func(r chi.Router) {
		r.Use(parseProxyName, findProxyByName(router))
		r.Get("/", getProxy(server))
		r.Get("/delay", getProxyDelay(server))
		r.Get("/httprequest", httpRequestByProxy(server))//karing
		r.Put("/", updateProxy)
	})
	return r
}

func parseProxyName(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := getEscapeParam(r, "name")
		ctx := context.WithValue(r.Context(), CtxKeyProxyName, name)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func findProxyByName(router adapter.Router) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			name := r.Context().Value(CtxKeyProxyName).(string)
			proxy, exist := router.Outbound(name)
			if !exist {
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, ErrNotFound)
				return
			}
			ctx := context.WithValue(r.Context(), CtxKeyProxy, proxy)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func proxyInfo(server *Server, detour adapter.Outbound) *badjson.JSONObject {
	var info badjson.JSONObject
	var clashType string
	switch detour.Type() {
	case C.TypeBlock:
		clashType = "Reject"
	default:
		clashType = C.ProxyDisplayName(detour.Type())
	}
	info.Put("type", clashType)
	info.Put("name", detour.Tag())
	info.Put("udp", common.Contains(detour.Network(), N.NetworkUDP))
	delayHistory := server.urlTestHistory.LoadURLTestHistory(adapter.OutboundTag(detour))
	if delayHistory != nil {
		info.Put("history", []*urltest.History{delayHistory})
	} else {
		info.Put("history", []*urltest.History{})
	}
	if group, isGroup := detour.(adapter.OutboundGroup); isGroup {
		info.Put("now", group.Now())
		info.Put("all", group.All())
	}
	return &info
}

func getProxies(server *Server, router adapter.Router) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var proxyMap badjson.JSONObject
		outbounds := common.Filter(router.Outbounds(), func(detour adapter.Outbound) bool {
			return detour.Tag() != ""
		})

		allProxies := make([]string, 0, len(outbounds))

		for _, detour := range outbounds {
			switch detour.Type() {
			case C.TypeDirect, C.TypeBlock, C.TypeDNS:
				continue
			}
			allProxies = append(allProxies, detour.Tag())
		}

		var defaultTag string
		if defaultOutbound, err := router.DefaultOutbound(N.NetworkTCP); err == nil {
			defaultTag = defaultOutbound.Tag()
		} else {
			defaultTag = allProxies[0]
		}

		sort.SliceStable(allProxies, func(i, j int) bool {
			return allProxies[i] == defaultTag
		})

		// fix clash dashboard
		proxyMap.Put("GLOBAL", map[string]any{
			"type":    "Fallback",
			"name":    "GLOBAL",
			"udp":     true,
			"history": []*urltest.History{},
			"all":     allProxies,
			"now":     defaultTag,
		})

		for i, detour := range outbounds {
			var tag string
			if detour.Tag() == "" {
				tag = F.ToString(i)
			} else {
				tag = detour.Tag()
			}
			proxyMap.Put(tag, proxyInfo(server, detour))
		}
		var responseMap badjson.JSONObject
		responseMap.Put("proxies", &proxyMap)
		response, err := responseMap.MarshalJSON()
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, newError(err.Error()))
			return
		}
		w.Write(response)
	}
}

func getProxy(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy := r.Context().Value(CtxKeyProxy).(adapter.Outbound)
		response, err := proxyInfo(server, proxy).MarshalJSON()
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, newError(err.Error()))
			return
		}
		w.Write(response)
	}
}

type UpdateProxyRequest struct {
	Name string `json:"name"`
}

func updateProxy(w http.ResponseWriter, r *http.Request) {
	req := UpdateProxyRequest{}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrBadRequest)
		return
	}

	proxy := r.Context().Value(CtxKeyProxy).(adapter.Outbound)
	selector, ok := proxy.(*outbound.Selector)
	if !ok {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError("Must be a Selector"))
		return
	}

	if !selector.SelectOutbound(req.Name) {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError("Selector update error: not found"))
		return
	}

	render.NoContent(w, r)
}

func getProxyDelay(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		url := query.Get("url")
		if strings.HasPrefix(url, "http://") {
			url = ""
		}
		timeout, err := strconv.ParseInt(query.Get("timeout"), 10, 32) //karing
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrBadRequest)
			return
		}

		proxy := r.Context().Value(CtxKeyProxy).(adapter.Outbound)
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
		defer cancel()

		delay, delay2, err := urltest.URLTest(ctx, url, proxy)
		defer func() {
			realTag := outbound.RealTag(proxy)
			if err != nil {
				//server.urlTestHistory.DeleteURLTestHistory(realTag)
				server.urlTestHistory.StoreURLTestHistory(realTag, &urltest.History{ //karing
					Time:  time.Now(),
					Delay: 0,
					Err:   err.Error(),
				})
			} else {
				server.urlTestHistory.StoreURLTestHistory(realTag, &urltest.History{
					Time:  time.Now(),
					Delay: delay,
					Err:   "", //karing
				})
			}
		}()

		if ctx.Err() != nil {
			//render.Status(r, http.StatusGatewayTimeout) //karing
			//render.JSON(w, r, ErrRequestTimeout) //karing
			render.JSON(w, r, newError(ctx.Err().Error())) //karing
			return
		}

		if err != nil || delay == 0 {
			//render.Status(r, http.StatusServiceUnavailable) //karing
			//render.JSON(w, r, newError("An error occurred in the delay test")) //karing
			render.JSON(w, r, newError(err.Error())) //karing
			return
		}

		render.JSON(w, r, render.M{
			"delay": delay,
			"delay2": delay2,
		})
	}
}
func httpRequestByProxy(server *Server) func(w http.ResponseWriter, r *http.Request) {  //karing
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		url := query.Get("url")
		timeout, err := strconv.ParseInt(query.Get("timeout"), 10, 32) //karing
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrBadRequest)
			return
		}

		proxy := r.Context().Value(CtxKeyProxy).(adapter.Outbound)
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
		defer cancel()

		statusCode,header, body, err := URLRequest(ctx, url, proxy)
	 
		if ctx.Err() != nil {
			//render.Status(r, http.StatusGatewayTimeout) 
			//render.JSON(w, r, ErrRequestTimeout) //karing
			render.JSON(w, r, newError(ctx.Err().Error())) 
			return
		}

		if err != nil   {
			//render.Status(r, http.StatusServiceUnavailable) //karing
			//render.JSON(w, r, newError("An error occurred in the delay test")) //karing
			render.JSON(w, r, newError(err.Error())) //karing
			return
		}

		render.JSON(w, r, render.M{
			"status_code" :statusCode,
			"header": header,
			"body": body,
		})
	}
}
func URLRequest(ctx context.Context, link string, detour N.Dialer) (statusCode int, header map[string][]string, content [] byte, err error) {//karing
	if link == "" {
		return 0, nil, nil, E.New("request url is empty" )
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
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
	
	return resp.StatusCode,resp.Header, content, err
}