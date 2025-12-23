package libbox

import (
	"time"
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
	if C.IsIos {
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
