package libbox

//karing
import (
	"time"
)

var restartFlag bool

func SetRestart(restart bool) {
	restartFlag = restart
}

func GetRestart() bool {
	return restartFlag
}

func (s *BoxService) ScreenOn() {
	if s.instance == nil {
		return
	}
	instance := s.instance.Instance()
	if instance != nil && instance.Box() == nil && instance.Box().Logger() != nil {
		instance.Box().Logger().Info("BoxService:ScreenOn")
	}
}

func (s *BoxService) ScreenOff() {
	if s.instance == nil {
		return
	}
	instance := s.instance.Instance()
	if instance != nil && instance.Box() == nil && instance.Box().Logger() != nil {
		instance.Box().Logger().Info("BoxService:ScreenOff")
	}
	s.stopResetTimer()
}

func (s *BoxService) UserPresent() {
	if s.instance == nil {
		return
	}
	instance := s.instance.Instance()
	if instance != nil && instance.Box() == nil && instance.Box().Logger() != nil {
		instance.Box().Logger().Info("BoxService:UserPresent")
	}
	s.startResetTimer(5, s.tryResetOutboundNetwork)
}

func (s *BoxService) startResetTimer(d time.Duration, f func()) {
	s.stopResetTimer()
	s.endPauseTimer = time.AfterFunc(d, f)
}

func (s *BoxService) stopResetTimer() {
	if s.endPauseTimer != nil {
		s.endPauseTimer.Stop()
	}
}

func (s *BoxService) tryResetNetwork() {
	if s.instance == nil {
		return
	}
	instance := s.instance.Instance()
	tags := s.getOutboundIfHasIssue()
	if len(tags) > 0 {
		if instance != nil && instance.Box() == nil && instance.Box().Logger() != nil {
			instance.Box().Logger().Info("BoxService:tryResetNetwork")
		}

		s.ResetNetwork()
	}
}

func (s *BoxService) tryResetOutboundNetwork() {
	if s.instance == nil {
		return
	}
	instance := s.instance.Instance()
	tags := s.getOutboundIfHasIssue()
	if len(tags) > 0 {
		if instance != nil && instance.Box() == nil && instance.Box().Logger() != nil {
			instance.Box().Logger().Info("BoxService:tryResetOutboundNetwork")
		}
		if instance != nil {
			instance.Box().Router().ResetOutboundNetwork(tags)
		}
	}
}

func (s *BoxService) GetConnections(includeConnections bool) string {
	if s.instance == nil {
		return "{}"
	}
	instance := s.instance.Instance()
	if instance != nil {
		return instance.GetConnections(includeConnections)
	}
	return "{}"
}

func (s *BoxService) getOutboundIfHasIssue() []string {
	if s.instance == nil {
		return []string{}
	}
	instance := s.instance.Instance()
	if instance != nil {
		return instance.GetOutboundIfHasIssue()
	}

	return []string{}
}
