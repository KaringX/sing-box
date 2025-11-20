//go:build with_karing

package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sagernet/sing-box/experimental/libbox"
	"github.com/sagernet/sing-box/log"
	E "github.com/sagernet/sing/common/exceptions"

	"github.com/spf13/cobra"
)

var commandRun2 = &cobra.Command{
	Use:   "run2",
	Short: "Run service",
	Run: func(cmd *cobra.Command, args []string) {
		err := runService()
		if err != nil {
			log.Error(err)
			return
		}

	},
}
var httpServer *http.Server
var boxService *libbox.BoxService
var quit = make(chan struct{})

func init() {
	mainCommand.AddCommand(commandRun2)
}

func createHttpServer() error {
	if serviceHttpPort != 0 {
		r := chi.NewMux()
		r.Route("/reload", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				extra, err := restartService()
				if err == nil {
					render.JSON(w, r, render.M{
						"err":   nil,
						"extra": nil,
					})
					return
				}
				render.JSON(w, r, render.M{
					"err":   err.Error(),
					"extra": extra,
				})

				go func() {
					time.Sleep(1 * time.Second)
					quit <- struct{}{}
				}()
			})
		})
		r.Route("/stop", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				libbox.SetRestart(false)
				quit <- struct{}{}
				render.JSON(w, r, render.M{
					"err": nil,
				})
			})
		})
		httpServer = &http.Server{
			Addr:    fmt.Sprintf("127.0.0.1:%d", serviceHttpPort),
			Handler: r,
		}
		go func() {
			err := httpServer.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) && !E.IsClosed(err) && boxService != nil {
				log.Fatal(E.Cause(err, "serve HTTP server"))
			}
		}()
	}

	if connectedPort != 0 {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", connectedPort))
		if err != nil {
			return err
		} else {
			conn.Close()
		}
	}
	return nil
}

func destoryServer() {
	server := httpServer
	httpServer = nil
	if server != nil {
		server.Close()
	}
}

func restartService() (map[string]string, error) {
	libbox.SetRestart(true)
	err := destoryService()
	if err != nil {
		var extra = make(map[string]string)
		extra["is_close_error"] = "true"
		return extra, err
	}
	libbox.StderrCheckAndCapture()
	err = createService()
	return nil, err
}

func destoryService() error {
	service := boxService
	boxService = nil
	if service != nil {
		return service.Close()
	}
	return nil
}

func createService() (err error) {
	defer func() {
		if e := recover(); e != nil {
			panicErrMessage := fmt.Sprintf("%v", e)
			libbox.SentryCaptureErrorMessage(panicErrMessage, "panic: createService", libbox.SentryTrim(string(debug.Stack())))
		}
	}()

	if len(configPaths) == 0 {
		return E.New("param [config] not found")
	}
	var configContent []byte
	for _, path := range configPaths {
		configContent, err = os.ReadFile(path)
		if err != nil {
			return err
		}
		if len(configContent) == 0 {
			return E.New("file content is empty: ", path)
		}
		break
	}

	boxService, err = libbox.NewService(string(configContent), nil)
	if err != nil {
		return err
	}

	err = boxService.Start()
	if err != nil {
		return err
	}

	return nil
}

func runService() (err error) {
	err = createService()
	if err != nil {
		return err
	}
	err = createHttpServer()
	if err != nil {
		return err
	}

	<-quit
	destoryAll()
	terminateCurrentProcess()
	return nil
}

func destoryAll() {
	defer func() {
		if e := recover(); e != nil {
			errMessage := fmt.Sprintf("%v", e)
			log.Error("panic: ", errMessage)
		}
		terminateCurrentProcess()
	}()
	destoryServer()
	destoryService()
}
