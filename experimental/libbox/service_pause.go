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
		s.instance.Logger().Info("BoxService:DevicePause")
	}
	s.ResetNetwork()             //karing
	s.pauseManager.DevicePause() //karing

	s.pauseAccess.Lock()
	defer s.pauseAccess.Unlock()
	if s.pauseTimer != nil {
		s.pauseTimer.Stop()
	}
	//s.pauseTimer = time.AfterFunc(3*time.Second, s.ResetNetwork) //karing
}

func (s *BoxService) Wake() {
	if s.instance != nil && s.instance.Logger() != nil { //karing
		s.instance.Logger().Info("BoxService:DeviceWake")
	}
	s.ResetNetwork()            //karing
	s.pauseManager.DeviceWake() //karing

	s.pauseAccess.Lock()
	defer s.pauseAccess.Unlock()
	if s.pauseTimer != nil {
		s.pauseTimer.Stop()
	}
	//s.pauseTimer = time.AfterFunc(3*time.Minute, s.ResetNetwork)//karing
}

func (s *BoxService) ResetNetwork() {
	s.instance.Router().ResetNetwork()
}

func (s *BoxService) UpdateWIFIState() {
	s.instance.Network().UpdateWIFIState()
}
