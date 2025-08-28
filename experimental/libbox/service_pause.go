package libbox

import (
	"time"

	C "github.com/sagernet/sing-box/constant"

	"github.com/sagernet/sing-box/adapter"

	"github.com/sagernet/sing-box/experimental/clashapi"
	"github.com/sagernet/sing/service"
)

type iOSPauseFields struct {
	endPauseTimer *time.Timer
}

func (s *BoxService) Pause() {
	if s.instance != nil && s.instance.Logger() != nil { //karing
		s.instance.Logger().Info("BoxService:Pause")
	}
	s.pauseManager.DevicePause()
	/*//karing
	if !C.IsIos {
		s.instance.Router().ResetNetwork()
	} else {
		if s.endPauseTimer == nil {
			s.endPauseTimer = time.AfterFunc(time.Minute, s.pauseManager.DeviceWake)
		} else {
			s.endPauseTimer.Reset(time.Minute)
		}
	}
	*/
	s.stopResetTimer() //karing
}

func (s *BoxService) Wake() {
	if s.instance != nil && s.instance.Logger() != nil { //karing
		s.instance.Logger().Info("BoxService:Wake")
	}
	/*//karing
	if !C.IsIos {
		s.pauseManager.DeviceWake()
		s.instance.Router().ResetNetwork()
	}
	*/

	s.ResetNetwork()            //karing
	s.pauseManager.DeviceWake() //karing

	s.startResetTimer(15*time.Second, s.tryResetNetwork) //karing
}

func (s *BoxService) ResetNetwork() {
	s.instance.Router().ResetNetwork()
}

func (s *BoxService) UpdateWIFIState() {
	s.instance.Network().UpdateWIFIState()
}

func (s *BoxService) ScreenOn() { //karing
	if s.instance != nil && s.instance.Logger() != nil {
		s.instance.Logger().Info("BoxService:ScreenOn")
	}
}

func (s *BoxService) ScreenOff() { //karing
	if s.instance != nil && s.instance.Logger() != nil {
		s.instance.Logger().Info("BoxService:ScreenOff")
	}
	s.stopResetTimer()
}

func (s *BoxService) UserPresent() { //karing
	if s.instance != nil && s.instance.Logger() != nil {
		s.instance.Logger().Info("BoxService:UserPresent")
	}
	s.startResetTimer(5, s.tryResetOutboundNetwork)
}

func (s *BoxService) startResetTimer(d time.Duration, f func()) { //karing
	s.stopResetTimer()
	s.endPauseTimer = time.AfterFunc(d, f)
}

func (s *BoxService) stopResetTimer() { //karing
	if s.endPauseTimer != nil {
		s.endPauseTimer.Stop()
	}
}

func (s *BoxService) tryResetNetwork() { //karing
	tags := s.getOutboundIfHasIssue()
	if len(tags) > 0 {
		if s.instance != nil && s.instance.Logger() != nil {
			s.instance.Logger().Info("BoxService:tryResetNetwork")
		}
		s.ResetNetwork()
	}
}

func (s *BoxService) tryResetOutboundNetwork() { //karing
	tags := s.getOutboundIfHasIssue()
	if len(tags) > 0 {
		if s.instance != nil && s.instance.Logger() != nil {
			s.instance.Logger().Info("BoxService:tryResetOutboundNetwork:", tags)
		}
		s.instance.Router().ResetOutboundNetwork(tags)
		//conntrack.Close()
	}
}

type outboundTranffic struct { //karing
	upload   int64
	download int64
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

func (s *BoxService) getOutboundIfHasIssue() []string { //karing
	outboundTranffics := make(map[string]outboundTranffic)
	clashServer := service.FromContext[adapter.ClashServer](s.ctx)
	if clashServer != nil {
		trafficManager := clashServer.(*clashapi.Server).TrafficManager()
		if trafficManager != nil {
			connections := trafficManager.Connections()
			for _, connection := range connections {
				if isProxyOutbound(connection.OutboundType) {
					conn, exist := outboundTranffics[connection.Outbound]
					if exist {
						outboundTranffics[connection.Outbound] = outboundTranffic{
							upload:   conn.upload + connection.Upload.Load(),
							download: conn.download + connection.Download.Load()}
					} else {
						outboundTranffics[connection.Outbound] = outboundTranffic{
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
