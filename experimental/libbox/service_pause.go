package libbox

import (
	"sync"
	"time"
)

type servicePauseFields struct {
	pauseAccess sync.Mutex
	pauseTimer  *time.Timer
}

func (s *BoxService) Pause() {
	if s.instance != nil && s.instance.Logger() != nil { //karing
		s.instance.Logger().Error("BoxService:DevicePause")
	}

	s.pauseAccess.Lock()
	defer s.pauseAccess.Unlock()
	if s.pauseTimer != nil {
		s.pauseTimer.Stop()
	}
	s.pauseTimer = time.AfterFunc(60*time.Second, s.ResetNetwork) //karing
}

func (s *BoxService) Wake() {
	if s.instance != nil && s.instance.Logger() != nil { //karing
		s.instance.Logger().Error("BoxService:DeviceWake")
	}

	s.pauseAccess.Lock()
	defer s.pauseAccess.Unlock()
	if s.pauseTimer != nil {
		s.pauseTimer.Stop()
	}
	//s.pauseTimer = time.AfterFunc(3*time.Minute, s.ResetNetwork)//karing
}

func (s *BoxService) ResetNetwork() {
	if s.instance != nil && s.instance.Logger() != nil { //karing
		s.instance.Logger().Error("BoxService:ResetNetwork")
	}
	s.pauseManager.DevicePause()  //karing
	s.pauseManager.NetworkPause() //karing
	s.instance.Router().ResetNetwork()
}
