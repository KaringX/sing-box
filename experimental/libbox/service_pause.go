package libbox

import (
	"sync"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/conntrack"
	"github.com/sagernet/sing-box/experimental/clashapi"
	"github.com/sagernet/sing/service"
)

type servicePauseFields struct {
	pauseAccess sync.Mutex
	pauseTimer  *time.Timer
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
	s.pauseTimer = time.AfterFunc(3*time.Second, s.ResetNetwork)
}

func (s *BoxService) Wake() {
	in, out := s.getConnectionInAndOutCount()
	if s.instance != nil && s.instance.Logger() != nil { //karing
		s.instance.Logger().Info("BoxService:DeviceWake connectionsIn:", in, " connectionsOut:", out)
	}

	s.pauseManager.DeviceWake() //karing

	s.pauseAccess.Lock()
	defer s.pauseAccess.Unlock()
	if s.pauseTimer != nil {
		s.pauseTimer.Stop()
	}
	s.pauseTimer = time.AfterFunc(3*time.Minute, s.ResetNetwork) //karing
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

func (s *BoxService) ResetNetwork() {
	s.instance.Router().ResetNetwork()
}

func (s *BoxService) UpdateWIFIState() {
	s.instance.Network().UpdateWIFIState()
}
