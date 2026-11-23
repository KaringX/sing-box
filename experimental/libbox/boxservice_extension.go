//karing

package libbox

import (
	"fmt"
	"runtime"
	"runtime/debug"
	runtimeDebug "runtime/debug"
	"time"

	E "github.com/sagernet/sing/common/exceptions"
)

func (s *BoxService) Reload(configContent string) (closeErr bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			panicErrMessage := fmt.Sprintf("%v", e)
			stack := SentryTrim(string(debug.Stack()))
			err = E.New(panicErrMessage, "\n", "panic: reload service", "\n", stack)
			SentryCapturePanicMessage(panicErrMessage, "panic: reload service", stack)
		}
	}()
	if s.instance == nil || s.instance.Instance() == nil {
		return false, E.New("instance is not started")
	}
	preTunInbound, preTunOptions := s.instance.Instance().InboundTunSuspend()
	err = s.instance.CloseService()
	if err != nil {
		return true, err
	}
	go func() {
		runtime.GC()
		runtimeDebug.FreeOSMemory()
	}()
	//daemon.RegisterStartedServiceServer(s.ctx, s.instance)
	err = s.instance.StartOrReloadService(s.ctx, configContent, nil, preTunInbound, preTunOptions)
	if err != nil {
		SentryCaptureMessage(err, "reload service")
		if preTunInbound != nil {
			preTunInbound.Close()
		}
		return false, err
	}

	go func() {
		runtime.GC()
		runtimeDebug.FreeOSMemory()
	}()
	return false, nil
}

func (s *BoxService) ScreenOn() {
	if s.instance == nil {
		return
	}
	instance := s.instance.Instance()
	if instance != nil && instance.Box() != nil && instance.Box().Logger() != nil {
		instance.Box().Logger().Info("BoxService:ScreenOn")
	}
}

func (s *BoxService) ScreenOff() {
	if s.instance == nil {
		return
	}
	instance := s.instance.Instance()
	if instance != nil && instance.Box() != nil && instance.Box().Logger() != nil {
		instance.Box().Logger().Info("BoxService:ScreenOff")
	}
	s.stopResetTimer()
}

func (s *BoxService) UserPresent() {
	if s.instance == nil {
		return
	}
	instance := s.instance.Instance()
	if instance != nil && instance.Box() != nil && instance.Box().Logger() != nil {
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
		if instance != nil && instance.Box() != nil && instance.Box().Logger() != nil {
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
		if instance != nil && instance.Box() != nil && instance.Box().Logger() != nil {
			instance.Box().Logger().Info("BoxService:tryResetOutboundNetwork")
		}
		if instance != nil {
			instance.Box().Router().ResetOutboundNetwork(s.ctx, tags)
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
