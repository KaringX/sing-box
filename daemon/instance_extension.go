package daemon

import (
	"encoding/json"
	"fmt"
	"runtime"
	runtimeDebug "runtime/debug"
	"time"

	"github.com/sagernet/sing-box/adapter"
	D "github.com/sagernet/sing-box/common/debug"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/experimental/clashapi"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/service"
)

type outboundUploadDownload struct {
	upload   int64
	download int64
}

func (i *Instance) close() error {
	i.cancel()
	if i.urlTestHistoryStorage != nil { //karing
		i.urlTestHistoryStorage.Close()
	}

	var goroutineId int         //karing
	var err error               //karing
	done := make(chan struct{}) //karing
	go func() {                 //karing
		goroutineId = D.GetCurrentGoroutineId()
		err = i.instance.Close()
		close(done)
		i.urlTestHistoryStorage = nil //karing
		i.clashServer = nil           //karing
		i.pauseManager = nil          //karing
		i.instance = nil              //karing
		runtime.GC()                  //karing
		runtimeDebug.FreeOSMemory()   //karing
	}()
	select {
	case <-done:
		return err
	case <-time.After(C.FatalStopTimeout):
		stack := D.GetGoroutineStack(goroutineId)      //karing
		return E.New("close service timeout:" + stack) //karing
	}
}

func (s *Instance) GetOutboundIfHasIssue() []string {
	outboundTranffics := make(map[string]outboundUploadDownload)
	clashServer := service.FromContext[adapter.ClashServer](s.ctx)
	if clashServer != nil {
		trafficManager := clashServer.(*clashapi.Server).TrafficManager()
		if trafficManager != nil {
			connections := trafficManager.Connections()
			for _, connection := range connections {
				if isProxyOutbound(connection.OutboundType) {
					conn, exist := outboundTranffics[connection.Outbound]
					if exist {
						outboundTranffics[connection.Outbound] = outboundUploadDownload{
							upload:   conn.upload + connection.Upload.Load(),
							download: conn.download + connection.Download.Load()}
					} else {
						outboundTranffics[connection.Outbound] = outboundUploadDownload{
							upload:   connection.Upload.Load(),
							download: connection.Download.Load()}
					}
				}
			}
		}
	}
	tags := make([]string, 0)
	for tag, outbound := range outboundTranffics {
		if outbound.upload > 0 && outbound.download == 0 {
			tags = append(tags, tag)
		}
	}
	return tags
}

func (s *Instance) GetConnections(includeConnections bool) string {
	if s.clashServer != nil {
		trafficManager := s.clashServer.(*clashapi.Server).TrafficManager()
		if trafficManager != nil {
			snapshot := trafficManager.Snapshot(includeConnections)
			data, err := json.Marshal(snapshot)
			if err != nil {
				return fmt.Sprintf("{err:%s}", err.Error())
			}
			if len(data) == 0 {
				return fmt.Sprintf("{}")
			}
			return string(data)
		}
	}
	return "{}"
}

func isProxyOutbound(outboundType string) bool {
	switch outboundType {
	case C.TypeSOCKS:
		return true
	case C.TypeHTTP:
		return true
	case C.TypeShadowsocks:
		return true
	case C.TypeVMess:
		return true
	case C.TypeTrojan:
		return true
	case C.TypeWireGuard:
		return true
	case C.TypeHysteria:
		return true
	case C.TypeTor:
		return true
	case C.TypeSSH:
		return true
	case C.TypeShadowTLS:
		return true
	case C.TypeAnyTLS:
		return true
	case C.TypeMieru:
		return true
	case C.TypeShadowsocksR:
		return true
	case C.TypeVLESS:
		return true
	case C.TypeTUIC:
		return true
	case C.TypeHysteria2:
		return true
	case C.TypeTailscale:
		return true
	default:
		return false
	}
}
