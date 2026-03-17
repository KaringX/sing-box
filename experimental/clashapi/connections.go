package clashapi

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/experimental/clashapi/trafficontrol"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
	"github.com/sagernet/ws"
	"github.com/sagernet/ws/wsutil"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/gofrs/uuid/v5"
)

func connectionRouter(ctx context.Context, server *Server, router adapter.Router, trafficManager *trafficontrol.Manager) http.Handler { //karing
	r := chi.NewRouter()
	r.Get("/", getConnections(ctx, server, trafficManager)) //karing
	r.Delete("/", closeAllConnections(router, trafficManager))
	r.Delete("/{id}", closeConnection(trafficManager))
	return r
}

func getConnections(ctx context.Context, server *Server, trafficManager *trafficontrol.Manager) func(w http.ResponseWriter, r *http.Request) { //karing
	return func(w http.ResponseWriter, r *http.Request) {
		noConnections := r.URL.Query().Get("noConnections") //karing
		if r.Header.Get("Upgrade") != "websocket" {
			snapshot := trafficManager.Snapshot(noConnections != "true") //karing
			render.JSON(w, r, snapshot)
			return
		}

		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			return
		}
		defer conn.Close()

		intervalStr := r.URL.Query().Get("interval")
		interval := 1000
		if intervalStr != "" {
			t, err := strconv.Atoi(intervalStr)
			if err != nil {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, ErrBadRequest)
				return
			}

			interval = t
		}

		buf := &bytes.Buffer{}
		sendSnapshot := func() error {
			buf.Reset()
			snapshot := trafficManager.Snapshot(noConnections != "true") //karing
			if err := json.NewEncoder(buf).Encode(snapshot); err != nil {
				return err
			}
			return wsutil.WriteServerText(conn, buf.Bytes())
		}

		if err = sendSnapshot(); err != nil {
			return
		}

		tick := time.NewTicker(time.Millisecond * time.Duration(interval))
		closed := false               //karing
		server.AddTick(tick, func() { //karing
			closed = true
		})
		defer func() { //karing
			server.RemoveTick(tick)
			tick.Stop()
		}()

		for range tick.C { //karing
			if closed { //karing
				break
			}
			pauseManager := service.FromContext[pause.Manager](server.ctx) //karing
			if pauseManager == nil || pauseManager.IsDevicePaused() {      //karing
				break
			}
			if err = sendSnapshot(); err != nil {
				break
			}
		}
	}
}

func closeConnection(trafficManager *trafficontrol.Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := uuid.FromStringOrNil(chi.URLParam(r, "id"))
		snapshot := trafficManager.Snapshot(true) //karing
		for _, c := range snapshot.Connections {
			if id == c.Metadata().ID {
				c.Close()
				break
			}
		}
		render.NoContent(w, r)
	}
}

func closeAllConnections(router adapter.Router, trafficManager *trafficontrol.Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot := trafficManager.Snapshot(true) //karing
		for _, c := range snapshot.Connections {
			c.Close()
		}
		router.ResetNetwork()
		render.NoContent(w, r)
	}
}
