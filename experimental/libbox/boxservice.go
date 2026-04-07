package libbox

//karing
import (
	"fmt"
	"runtime"
	"runtime/debug"
	runtimeDebug "runtime/debug"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/daemon"
	"github.com/sagernet/sing-box/log"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/service"
)

// CommandServer
type BoxService struct {
	instance          *daemon.StartedService
	handler           CommandServerHandler
	platformInterface PlatformInterface
	platformWrapper   *platformInterfaceWrapper
	endPauseTimer     *time.Timer
}

type BoxServiceHandler interface {
	ServiceStop() error
	ServiceReload() error
	GetSystemProxyStatus() (*SystemProxyStatus, error)
	SetSystemProxyEnabled(enabled bool) error
	WriteDebugMessage(message string)
}

func NewService(handler BoxServiceHandler, platformInterface PlatformInterface) (boxService *BoxService, err error) {
	SentryBoxServiceLaunch()
	defer func() {
		if e := recover(); e != nil {
			panicErrMessage := fmt.Sprintf("%v", e)
			stack := SentryTrim(string(debug.Stack()))
			err = E.New(panicErrMessage, "\n", "panic: create service", "\n", stack)
			SentryCapturePanicMessage(panicErrMessage, "panic: create service", stack)
		}
	}()
	ctx := baseContext(platformInterface)

	var platformWrapper *platformInterfaceWrapper //karing
	if platformInterface != nil {                 //karing
		platformWrapper = &platformInterfaceWrapper{ //karing
			iif:       platformInterface,
			useProcFS: platformInterface.UseProcFS(),
		}
		service.MustRegister[adapter.PlatformInterface](ctx, platformWrapper)
	}

	server := &BoxService{
		handler:           handler,
		platformInterface: platformInterface,
		platformWrapper:   platformWrapper,
	}
	server.instance = daemon.NewStartedService(daemon.ServiceOptions{
		Context: ctx,
		// Platform:         platformWrapper,
		Handler:     (*boxServiceplatformHandler)(server),
		Debug:       sDebug,
		LogMaxLines: sLogMaxLines,
		OOMKiller:   memoryLimitEnabled,
		// WorkingDirectory: sWorkingPath,
		// TempDirectory:    sTempPath,
		// UserID:           sUserID,
		// GroupID:          sGroupID,
		// SystemProxyEnabled: false,
	})
	return server, nil
}

func (s *BoxService) Start(configContent string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			panicErrMessage := fmt.Sprintf("%v", e)
			stack := SentryTrim(string(debug.Stack()))
			err = E.New(panicErrMessage, "\n", "panic: start service", "\n", stack)
			SentryCapturePanicMessage(panicErrMessage, "panic: start service", stack)
		}
	}()
	//daemon.RegisterStartedServiceServer(nil, s.StartedService)
	err = s.instance.StartOrReloadService(configContent, nil)
	if err != nil {
		SentryCaptureMessage(err, "start service")
		return err
	}
	go func() {
		runtime.GC()
		runtimeDebug.FreeOSMemory()
	}()
	return nil
}

func (s *BoxService) Close() error {
	if s.instance == nil {
		return nil
	}
	s.instance.Close()
	return s.instance.CloseService()
}

func (s *BoxService) WriteMessage(level int32, message string) {
	if s.instance == nil {
		return
	}
	s.instance.WriteMessage(log.Level(level), message)
}

func (s *BoxService) SetError(message string) {
	if s.instance == nil {
		return
	}
	s.instance.SetError(E.New(message))
}

func (s *BoxService) NeedWIFIState() bool {
	if s.instance == nil {
		return false
	}
	instance := s.instance.Instance()
	if instance == nil || instance.Box() == nil {
		return false
	}
	return instance.Box().Network().NeedWIFIState()
}

func (s *BoxService) NeedFindProcess() bool {
	if s.instance == nil {
		return false
	}
	instance := s.instance.Instance()
	if instance == nil || instance.Box() == nil {
		return false
	}
	return instance.Box().Router().NeedFindProcess()
}

func (s *BoxService) Pause() {
	if s.instance == nil {
		return
	}
	instance := s.instance.Instance()
	if instance == nil || instance.PauseManager() == nil {
		return
	}
	if instance.Box() != nil && instance.Box().Logger() != nil {
		instance.Box().Logger().Info("BoxService:Pause")
	}
	instance.PauseManager().DevicePause()
	s.stopResetTimer()
}

func (s *BoxService) Wake() {
	if s.instance == nil {
		return
	}
	instance := s.instance.Instance()
	if instance == nil || instance.PauseManager() == nil {
		return
	}
	if instance.Box() != nil && instance.Box().Logger() != nil {
		instance.Box().Logger().Info("BoxService:Wake")
	}

	s.ResetNetwork()
	instance.PauseManager().DeviceWake()
	s.startResetTimer(15*time.Second, s.tryResetNetwork)
}

func (s *BoxService) ResetNetwork() {
	if s.instance == nil {
		return
	}
	instance := s.instance.Instance()
	if instance == nil || instance.Box() == nil {
		return
	}
	instance.Box().Router().ResetNetwork()
}

func (s *BoxService) UpdateWIFIState() {
	if s.instance == nil {
		return
	}
	instance := s.instance.Instance()
	if instance == nil || instance.Box() == nil {
		return
	}
	instance.Box().Network().UpdateWIFIState()
}

type boxServiceplatformHandler BoxService

func (h *boxServiceplatformHandler) ServiceStop() error {
	handler := (*BoxService)(h).handler
	if handler == nil {
		return nil
	}
	return handler.ServiceStop()
}

func (h *boxServiceplatformHandler) ServiceReload() error {
	handler := (*BoxService)(h).handler
	if handler == nil {
		return nil
	}
	return handler.ServiceReload()
}

func (h *boxServiceplatformHandler) SystemProxyStatus() (*daemon.SystemProxyStatus, error) {
	handler := (*BoxService)(h).handler
	if handler == nil {
		return nil, nil
	}
	status, err := handler.GetSystemProxyStatus()
	if err != nil {
		return nil, err
	}
	return &daemon.SystemProxyStatus{
		Enabled:   status.Enabled,
		Available: status.Available,
	}, nil
}

func (h *boxServiceplatformHandler) SetSystemProxyEnabled(enabled bool) error {
	handler := (*BoxService)(h).handler
	if handler == nil {
		return nil
	}
	return handler.SetSystemProxyEnabled(enabled)
}

func (h *boxServiceplatformHandler) WriteDebugMessage(message string) {
	handler := (*BoxService)(h).handler
	if handler == nil {
		return
	}
	handler.WriteDebugMessage(message)
}
