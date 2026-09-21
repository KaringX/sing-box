//karing

package libbox

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
	runtimeDebug "runtime/debug"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/daemon"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/service/oomkiller"
	"github.com/sagernet/sing-box/service/powerreport"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/service"
)

// CommandServer
type BoxService struct {
	instance          *daemon.StartedService
	ctx               context.Context
	managedService    *daemon.ManagedService
	handler           CommandServerHandler
	platformInterface PlatformInterface
	platformWrapper   *platformInterfaceWrapper
	powerManager      *powerreport.Manager
	oomRecorder       *oomkiller.Recorder
	endPauseTimer     *time.Timer
}

type BoxServiceHandler interface {
	ServiceStop() error
	ServiceReload() error
	GetSystemProxyStatus() (*SystemProxyStatus, error)
	SetSystemProxyEnabled(enabled bool) error
	TriggerNativeCrash() error
	WriteDebugMessage(message string)
	ConnectSSHAgent() (int32, error)
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
	powerManager := powerreport.NewManager()
	service.MustRegister[*powerreport.Manager](ctx, powerManager)
	var platformWrapper *platformInterfaceWrapper //karing
	if platformInterface != nil {                 //karing
		platformWrapper = &platformInterfaceWrapper{ //karing
			iif:          platformInterface,
			useProcFS:    platformInterface.UseProcFS(),
			powerManager: powerManager,
		}
		service.MustRegister[adapter.PlatformInterface](ctx, platformWrapper)
	}

	server := &BoxService{
		ctx:               ctx,
		handler:           handler,
		platformInterface: platformInterface,
		platformWrapper:   platformWrapper,
		powerManager:      powerManager,
	}
	server.instance = daemon.NewStartedService(daemon.ServiceOptions{
		Context: ctx,
		// Platform:         platformWrapper,
		Handler:           (*boxServicePlatformHandler)(server),
		Debug:             sDebug,
		LogMaxLines:       sLogMaxLines,
		OOMKillerEnabled:  sOOMKillerEnabled,
		OOMKillerDisabled: sOOMKillerDisabled,
		OOMMemoryLimit:    uint64(sOOMMemoryLimit),
		// WorkingDirectory: sWorkingPath,
		// TempDirectory:    sTempPath,
		// UserID:           sUserID,
		// GroupID:          sGroupID,
		// SystemProxyEnabled: false,
	})
	oomRecorder := oomkiller.NewRecorder(OOMRecorderOptions(server.instance))
	service.MustRegister[*oomkiller.Recorder](ctx, oomRecorder)
	oomRecorder.Start()
	server.oomRecorder = oomRecorder
	server.managedService = daemon.NewManagedService(daemon.ManagedServiceOptions{
		Handler:     (*boxServicePlatformHandler)(server),
		Debug:       sDebug,
		OOMRecorder: oomRecorder,
	})
	if sPowerReportEnabled {
		err := powerManager.Start(PowerReportOptions(server.instance))
		if err != nil {
			log.StdLogger().Error(E.Cause(err, "start power report recorder"))
		}
	}
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
	//daemon.RegisterStartedServiceServer(s.ctx, s.instance)
	err = s.instance.StartOrReloadService(s.ctx, configContent, nil)
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
	s.oomRecorder.Close()
	s.powerManager.Close()
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
	if instance == nil || instance.Box() == nil || instance.Box().Network() == nil { //karing
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
	if instance == nil || instance.Box() == nil || instance.Box().Network() == nil { //karing
		return
	}
	instance.Box().Network().UpdateWIFIState(context.Background())
}

type boxServicePlatformHandler BoxService

func (h *boxServicePlatformHandler) ServiceStop() error {
	return (*BoxService)(h).handler.ServiceStop()
}

func (h *boxServicePlatformHandler) ServiceReload(ctx context.Context) error {
	return (*BoxService)(h).handler.ServiceReload()
}

func (h *boxServicePlatformHandler) SystemProxyStatus() (*daemon.SystemProxyStatus, error) {
	status, err := (*BoxService)(h).handler.GetSystemProxyStatus()
	if err != nil {
		return nil, E.Cause(err, "get system proxy status")
	}
	return &daemon.SystemProxyStatus{
		Enabled:   status.Enabled,
		Available: status.Available,
	}, nil
}

func (h *boxServicePlatformHandler) SetSystemProxyEnabled(enabled bool) error {
	return (*BoxService)(h).handler.SetSystemProxyEnabled(enabled)
}

func (h *boxServicePlatformHandler) TriggerNativeCrash() error {
	return (*BoxService)(h).handler.TriggerNativeCrash()
}

func (h *boxServicePlatformHandler) WriteDebugMessage(message string) {
	(*BoxService)(h).handler.WriteDebugMessage(message)
}

func (h *boxServicePlatformHandler) ConnectSSHAgent() (int32, error) {
	return (*BoxService)(h).handler.ConnectSSHAgent()
}
