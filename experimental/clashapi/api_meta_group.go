package clashapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/gofree" //karing
	"github.com/sagernet/sing-box/common/urltest"
	"github.com/sagernet/sing-box/protocol/group"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/batch"
	"github.com/sagernet/sing/common/json/badjson"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func groupRouter(server *Server) http.Handler {
	r := chi.NewRouter()
	r.Get("/", getGroups(server))
	r.Route("/{name}", func(r chi.Router) {
		r.Use(parseProxyName, findProxyByName(server))
		r.Get("/", getGroup(server))
		r.Get("/delay", getGroupDelay(server))
		r.Get("/delayUpdateCheck", updateGroupDelayCheck(server)) //karing
	})
	r.Get("/delayhistory", getProxyDelayHistory(server)) //karing
	return r
}

func getGroups(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		groups := common.Map(common.Filter(server.outbound.Outbounds(), func(it adapter.Outbound) bool {
			_, isGroup := it.(adapter.OutboundGroup)
			return isGroup
		}), func(it adapter.Outbound) *badjson.JSONObject {
			return proxyInfo(server, it)
		})
		render.JSON(w, r, render.M{
			"proxies": groups,
		})
	}
}

func getGroup(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy := r.Context().Value(CtxKeyProxy).(adapter.Outbound)
		if _, ok := proxy.(adapter.OutboundGroup); ok {
			render.JSON(w, r, proxyInfo(server, proxy))
			return
		}
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, ErrNotFound)
	}
}

func getGroupDelay(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy := r.Context().Value(CtxKeyProxy).(adapter.Outbound)
		outboundGroup, ok := proxy.(adapter.OutboundGroup)
		if !ok {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, ErrNotFound)
			return
		}

		query := r.URL.Query()
		url := query.Get("url")
		if strings.HasPrefix(url, "http://") {
			url = ""
		}
		timeout, err := strconv.ParseInt(query.Get("timeout"), 10, 32)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*time.Duration(timeout))
		defer cancel()

		var result map[string]urltest.URLTestResult //karing
		if urlTestGroup, isURLTestGroup := outboundGroup.(adapter.URLTestGroup); isURLTestGroup {
			result, err = urlTestGroup.URLTest(ctx)
		} else {
			outbounds := common.FilterNotNil(common.Map(outboundGroup.All(), func(it string) adapter.Outbound {
				itOutbound, _ := server.outbound.Outbound(it)
				return itOutbound
			}))
			b, _ := batch.New(ctx, batch.WithConcurrencyNum[any](10))
			checked := make(map[string]bool)
			result = make(map[string]urltest.URLTestResult) //karing
			var resultAccess sync.Mutex
			for _, detour := range outbounds {
				tag := detour.Tag()
				realTag := group.RealTag(detour)
				if checked[realTag] {
					continue
				}
				checked[realTag] = true
				p, loaded := server.outbound.Outbound(realTag)
				if !loaded {
					continue
				}
				b.Go(realTag, func() (any, error) {
					t, _, err := urltest.URLTest(ctx, url, p) //karing
					if err != nil {
						server.logger.Debug("outbound ", tag, " unavailable: ", err)
						//server.urlTestHistory.DeleteURLTestHistory(realTag)
						server.urlTestHistory.StoreURLTestHistory(realTag, &urltest.History{ //karing
							Time:  time.Now(),
							Delay: 0,
							Err:   err.Error(),
						})
					} else {
						server.logger.Debug("outbound ", tag, " available: ", t, "ms")
						server.urlTestHistory.StoreURLTestHistory(realTag, &urltest.History{
							Time:  time.Now(),
							Delay: t,
							Err:   "",
						})
					}
					resultAccess.Lock() //karing
					if err == nil {
						result[tag] = urltest.URLTestResult{Delay: t, Err: ""}
					} else {
						result[tag] = urltest.URLTestResult{Delay: t, Err: err.Error()}
					}
					resultAccess.Unlock() //karing
					return nil, nil
				})
			}
			b.Wait()
			gofree.FreeIdleThread()  //karing
		}

		if err != nil {
			render.Status(r, http.StatusGatewayTimeout)
			render.JSON(w, r, newError(err.Error()))
			return
		}

		render.JSON(w, r, result)
	}
}

func updateGroupDelayCheck(server *Server) func(w http.ResponseWriter, r *http.Request) { //karing
	return func(w http.ResponseWriter, r *http.Request) {
		proxy := r.Context().Value(CtxKeyProxy).(adapter.Outbound)
		group, ok := proxy.(adapter.OutboundGroup)
		if !ok {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, ErrNotFound)
			return
		}

		if urlTestGroup, isURLTestGroup := group.(adapter.URLTestGroup); isURLTestGroup {
			urlTestGroup.UpdateCheck()
		}

		render.JSON(w, r, "")
	}
}

func getProxyDelayHistory(server *Server) func(w http.ResponseWriter, r *http.Request) { //karing
	return func(w http.ResponseWriter, r *http.Request) {
		delayHistory := server.urlTestHistory.GetURLTestHistory()
		render.JSON(w, r, render.M{
			"history": delayHistory,
		})
	}
}
