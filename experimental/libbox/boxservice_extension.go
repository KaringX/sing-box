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

/*
func (s *BoxService) Start() (err error) { //karing

		defer func() { //karing
			if e := recover(); e != nil {
				panicErrMessage := fmt.Sprintf("%v", e)
				stack := SentryTrim(string(debug.Stack()))
				err = E.New(panicErrMessage, "\n", "panic: start service", "\n", stack)
				SentryCaptureErrorMessage(panicErrMessage, "panic: start service", stack)
			}
		}()
		D.MainGoroutineId = D.GetCurrentGoroutineId() //karing

			err = s.instance.Start() //karing

		if err != nil { //karing
			SentryCaptureError(err, "start service")
		} else { //karing
			go func() {
				runtime.GC()
				runtimeDebug.FreeOSMemory()
			}()
		}
		return err
	}

*/

func (s *BoxService) ScreenOn() {
	instance := s.StartedService.Instance()
	if instance != nil && instance.Box() == nil && instance.Box().Logger() != nil {
		instance.Box().Logger().Info("BoxService:ScreenOn")
	}
}

func (s *BoxService) ScreenOff() {
	instance := s.StartedService.Instance()
	if instance != nil && instance.Box() == nil && instance.Box().Logger() != nil {
		instance.Box().Logger().Info("BoxService:ScreenOff")
	}
	s.stopResetTimer()
}

func (s *BoxService) UserPresent() {
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
	tags := s.getOutboundIfHasIssue()
	if len(tags) > 0 {
		instance := s.StartedService.Instance()
		if instance != nil && instance.Box() == nil && instance.Box().Logger() != nil {
			instance.Box().Logger().Info("BoxService:tryResetNetwork")
		}

		s.ResetNetwork()
	}
}

func (s *BoxService) tryResetOutboundNetwork() {
	tags := s.getOutboundIfHasIssue()
	if len(tags) > 0 {
		instance := s.StartedService.Instance()
		if instance != nil && instance.Box() == nil && instance.Box().Logger() != nil {
			instance.Box().Logger().Info("BoxService:tryResetOutboundNetwork")
		}
		if instance != nil {
			instance.Box().Router().ResetOutboundNetwork(tags)
		}
	}
}

func (s *BoxService) GetConnections(includeConnections bool) string {
	instance := s.StartedService.Instance()
	if instance != nil {
		return instance.GetConnections(includeConnections)
	}
	return "{}"
}

func (s *BoxService) getOutboundIfHasIssue() []string {
	instance := s.StartedService.Instance()
	if instance != nil {
		return instance.GetOutboundIfHasIssue()
	}

	return []string{}
}
