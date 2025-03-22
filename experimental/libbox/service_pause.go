package libbox

import (
	"sync"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/conntrack"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/experimental/clashapi"
	"github.com/sagernet/sing/service"
)

type servicePauseFields struct {
	pauseAccess sync.Mutex
	pauseTimer  *time.Timer
}

type proxyOutboundUploadAndDownload struct {
	upload   int64
	download int64
}

func (s *BoxService) Pause() {
	in, out := s.getConnectionInAndOutCount()
	if s.instance != nil && s.instance.Logger() != nil { //karing
		s.instance.Logger().Info("BoxService:DevicePause connectionsIn:", in, " connectionsOut:", out)
	}
	s.pauseManager.DevicePause() //karing

	s.pauseAccess.Lock()
	defer s.pauseAccess.Unlock()
	if s.pauseTimer != nil {
		s.pauseTimer.Stop()
	}
	//s.pauseTimer = time.AfterFunc(3*time.Second, s.ResetNetwork) //karing
}

func (s *BoxService) Wake() {
	in, out := s.getConnectionInAndOutCount()
	if s.instance != nil && s.instance.Logger() != nil { //karing
		s.instance.Logger().Info("BoxService:DeviceWake connectionsIn:", in, " connectionsOut:", out)
	}
	//s.ResetNetwork()            //karing
	s.pauseManager.DeviceWake() //karing

	s.pauseAccess.Lock()
	defer s.pauseAccess.Unlock()
	if s.pauseTimer != nil {
		s.pauseTimer.Stop()
	}
	//s.pauseTimer = time.AfterFunc(15*time.Second, s.TryResetNetwork) //karing
	//s.pauseTimer = time.AfterFunc(3*time.Minute, s.ResetNetwork) //karing
}

func (s *BoxService) ResetNetwork() {
	s.instance.Router().ResetNetwork()
}

func (s *BoxService) UpdateWIFIState() {
	s.instance.Network().UpdateWIFIState()
}

func (s *BoxService) getConnectionInAndOutCount() (int, int) { //karing
	var connectionsIn int
	var connectionsOut int
	clashServer := service.FromContext[adapter.ClashServer](s.ctx)
	if clashServer != nil {
		trafficManager := clashServer.(*clashapi.Server).TrafficManager()
		if trafficManager != nil {
			connectionsIn = trafficManager.ConnectionsLen()
		}
	}
	connectionsOut = conntrack.Count()
	return connectionsIn, connectionsOut
}

func isProxyOutbound(outboundType string) bool { //karing
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
	case C.TypeShadowsocksR:
		return true
	case C.TypeVLESS:
		return true
	case C.TypeTUIC:
		return true
	case C.TypeHysteria2:
		return true
	default:
		return false
	}
}

func (s *BoxService) TryResetNetwork() { //karing
	outboundTranffic := make(map[string]proxyOutboundUploadAndDownload)
	clashServer := service.FromContext[adapter.ClashServer](s.ctx)
	if clashServer != nil {
		trafficManager := clashServer.(*clashapi.Server).TrafficManager()
		if trafficManager != nil {
			connections := trafficManager.Connections()
			for _, connection := range connections {
				if isProxyOutbound(connection.OutboundType) {
					conn, exist := outboundTranffic[connection.Outbound]
					if exist {
						outboundTranffic[connection.Outbound] = proxyOutboundUploadAndDownload{
							upload:   conn.upload + connection.Upload.Load(),
							download: conn.download + connection.Download.Load()}
					} else {
						outboundTranffic[connection.Outbound] = proxyOutboundUploadAndDownload{
							upload:   connection.Upload.Load(),
							download: connection.Download.Load()}
					}
				}
			}
		}
	}
	reset := false
	for _, outbound := range outboundTranffic {
		if outbound.upload > 0 && outbound.download == 0 {
			reset = true
			break
		}
	}
	if reset {
		if s.instance != nil && s.instance.Logger() != nil { //karing
			s.instance.Logger().Error("BoxService:TryResetNetwork")
		}
		s.ResetNetwork()
	}
}
