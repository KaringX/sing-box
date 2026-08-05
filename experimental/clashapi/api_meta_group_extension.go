//go:build with_karing

package clashapi

import (
	"net/http"

	"github.com/sagernet/sing-box/adapter"

	"github.com/go-chi/render"
)

func updateGroupDelayCheck(server *Server) func(w http.ResponseWriter, r *http.Request) {
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

func getGroupDelayHistory(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		delayHistory := server.urlTestHistory.GetURLTestHistory()
		render.JSON(w, r, render.M{
			"history": delayHistory,
		})
	}
}
