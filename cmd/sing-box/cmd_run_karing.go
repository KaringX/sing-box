//go:build with_karing

package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	runtimeDebug "runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	D "github.com/sagernet/sing-box/common/debug"
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
				destoryService()
				err := createService()
				if err != nil {
					render.JSON(w, r, render.M{
						"err": err.Error(),
					})
					if httpServer != nil {
						httpServer.Close()
						httpServer = nil
					}
					go func() {
						time.Sleep(1 * time.Second)
						terminateCurrentProcess()
					}()
				} else {
					go runtimeDebug.FreeOSMemory()
					render.JSON(w, r, render.M{
						"err": nil,
					})
				}
			})
		})
		r.Route("/stop", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				if httpServer != nil {
					httpServer.Close()
					httpServer = nil
				}
				if boxService != nil {
					boxService.Close()
					boxService = nil
				}
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
			if err != nil && !errors.Is(err, http.ErrServerClosed) && !E.IsClosed(err) {
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

func destoryService() {
	if boxService != nil {
		boxService.Close()
		boxService = nil
	}
}
func createService() (err error) {
	defer func() {
		if e := recover(); e != nil {
			recoverMessage := fmt.Sprintf("%v", e)
			libbox.SentryCaptureException(recoverMessage, "panic: create service", libbox.SentryTrim(string(debug.Stack())))
		}
	}()
	stacks := D.Stacks(false, false)
	if len(stacks) > 0 {
		for key := range stacks {
			D.MainGoId = key
			break
		}
	}
	if len(configPaths) == 0 {
		return E.Cause(err, "param [config] not found")
	}
	var configContent []byte
	for _, path := range configPaths {
		configContent, err = os.ReadFile(path)
		if err != nil {
			return err
		}
		if len(configContent) == 0 {
			return E.Cause(err, "file content is empty: ", path)
		}
		break
	}

	boxService, err = libbox.NewService(string(configContent), nil)
	if err != nil {
		libbox.SentryCaptureMessage(E.Cause(err, "create service"))
		return E.Cause(err, "create service")
	}

	err = boxService.Start()
	if err != nil {
		libbox.SentryCaptureMessage(E.Cause(err, "start service"))
		return E.Cause(err, "start service")
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
	go runtimeDebug.FreeOSMemory()
	<-quit
	terminateCurrentProcess()
	return nil
}
