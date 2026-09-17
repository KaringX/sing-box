//karing

package libbox

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
	if s.StartedService == nil {
		return
	}
	instance := s.StartedService.Instance()
	if instance != nil && instance.Box() == nil && instance.Box().Logger() != nil {
		instance.Box().Logger().Info("BoxService:ScreenOn")
	}
}

func (s *BoxService) ScreenOff() {
	if s.StartedService == nil {
		return
	}
	instance := s.StartedService.Instance()
	if instance != nil && instance.Box() == nil && instance.Box().Logger() != nil {
		instance.Box().Logger().Info("BoxService:ScreenOff")
	}
	s.stopResetTimer()
}

func (s *BoxService) UserPresent() {
	if s.StartedService == nil {
		return
	}
	instance := s.StartedService.Instance()
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
	if s.StartedService == nil {
		return
	}
	instance := s.StartedService.Instance()
	tags := s.getOutboundIfHasIssue()
	if len(tags) > 0 {
		if instance != nil && instance.Box() == nil && instance.Box().Logger() != nil {
			instance.Box().Logger().Info("BoxService:tryResetNetwork")
		}

		s.ResetNetwork()
	}
}

func (s *BoxService) tryResetOutboundNetwork() {
	if s.StartedService == nil {
		return
	}
	instance := s.StartedService.Instance()
	tags := s.getOutboundIfHasIssue()
	if len(tags) > 0 {
		if instance != nil && instance.Box() == nil && instance.Box().Logger() != nil {
			instance.Box().Logger().Info("BoxService:tryResetOutboundNetwork")
		}
		if instance != nil {
			instance.Box().Router().ResetOutboundNetwork(s.ctx, tags)
		}
	}
}

func (s *BoxService) GetConnections(includeConnections bool) string {
	if s.StartedService == nil {
		return "{}"
	}
	instance := s.StartedService.Instance()
	if instance != nil {
		return instance.GetConnections(includeConnections)
	}
	return "{}"
}

func (s *BoxService) getOutboundIfHasIssue() []string {
	if s.StartedService == nil {
		return []string{}
	}
	instance := s.StartedService.Instance()
	if instance != nil {
		return instance.GetOutboundIfHasIssue()
	}

	return []string{}
}
