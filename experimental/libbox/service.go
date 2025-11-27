package libbox

import (
	"context"
	"fmt"
	"net/netip"
	"runtime"
	"runtime/debug"
	runtimeDebug "runtime/debug"
	"strconv"
	"sync"
	"syscall"
	"time"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/adapter"
	D "github.com/sagernet/sing-box/common/debug"
	"github.com/sagernet/sing-box/common/process"
	"github.com/sagernet/sing-box/common/urltest"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/experimental/deprecated"
	"github.com/sagernet/sing-box/experimental/libbox/internal/procfs"
	"github.com/sagernet/sing-box/experimental/libbox/platform"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	tun "github.com/sagernet/sing-tun"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/control"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
)

type BoxService struct {
	ctx                   context.Context
	cancel                context.CancelFunc
	urlTestHistoryStorage adapter.URLTestHistoryStorage
	instance              *box.Box
	clashServer           adapter.ClashServer
	pauseManager          pause.Manager

	iOSPauseFields
}

func NewService(configContent string, platformInterface PlatformInterface) (boxService *BoxService, err error) { //karing
	SentryBoxServiceLaunch()
	defer func() { //karing
		if e := recover(); e != nil {
			panicErrMessage := fmt.Sprintf("%v", e)
			SentryCaptureErrorMessage(panicErrMessage, "panic: create service", SentryTrim(string(debug.Stack())))
		}
	}()
	D.MainGoroutineId = D.GetCurrentGoroutineId()                                                                 //karing
	ctx := context.WithValue(BaseContext(platformInterface), log.CtxKeyLogContextIdName, strconv.Itoa(contextId)) //karing
	contextId++                                                                                                   //karing
	service.MustRegister[deprecated.Manager](ctx, new(deprecatedManager))
	var options option.Options                     //karing
	options, err = parseConfig(ctx, configContent) //karing
	if err != nil {
		SentryCaptureError(err, "create service") //karing
		return nil, err
	}
	runtimeDebug.FreeOSMemory()
	ctx, cancel := context.WithCancel(ctx)
	urlTestHistoryStorage := urltest.NewHistoryStorage()
	service.MustRegister[adapter.URLTestHistoryStorage](ctx, urlTestHistoryStorage) //karing
	//ctx = service.ContextWithPtr(ctx, urlTestHistoryStorage)//karing
	var platformLogWriter log.PlatformWriter //karing
	if platformInterface != nil {            //karing
		var platformWrapper *platformInterfaceWrapper //karing
		platformWrapper = &platformInterfaceWrapper{  //karing
			iif:       platformInterface,
			useProcFS: platformInterface.UseProcFS(),
		}
		service.MustRegister[platform.Interface](ctx, platformWrapper)
		platformLogWriter = platformWrapper //karing
	}

	var instance *box.Box                //karing
	instance, err = box.New(box.Options{ //karing
		Context:           ctx,
		Options:           options,
		PlatformLogWriter: platformLogWriter, //karing
	})
	if err != nil {
		cancel()
		SentryCaptureError(err, "create service") //karing
		return nil, E.Cause(err, "create service")
	}
	runtimeDebug.FreeOSMemory()
	return &BoxService{
		ctx:                   ctx,
		cancel:                cancel,
		instance:              instance,
		urlTestHistoryStorage: urlTestHistoryStorage,
		pauseManager:          service.FromContext[pause.Manager](ctx),
		clashServer:           service.FromContext[adapter.ClashServer](ctx),
	}, nil
}

func (s *BoxService) Start() (err error) { //karing
	defer func() { //karing
		if e := recover(); e != nil {
			panicErrMessage := fmt.Sprintf("%v", e)
			SentryCaptureErrorMessage(panicErrMessage, "panic: start service", SentryTrim(string(debug.Stack())))
		}
	}()
	D.MainGoroutineId = D.GetCurrentGoroutineId() //karing
	if sFixAndroidStack {
		//var err error //karing
		done := make(chan struct{})
		go func() {
			err = s.instance.Start()
			close(done)
		}()
		<-done
		//return err //karing
	} else {
		err = s.instance.Start() //karing
	}
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

func (s *BoxService) Close() error {
	s.cancel()
	if s.urlTestHistoryStorage != nil { //karing
		s.urlTestHistoryStorage.Close()
	}

	var goroutineId int //karing
	var err error
	done := make(chan struct{})
	go func() {
		goroutineId = D.GetCurrentGoroutineId()
		err = s.instance.Close()
		close(done)
		s.urlTestHistoryStorage = nil //karing
		s.clashServer = nil           //karing
		s.pauseManager = nil          //karing
		s.instance = nil              //karing

		runtime.GC()                //karing
		runtimeDebug.FreeOSMemory() //karing
	}()
	select {
	case <-done:
		return err
	case <-time.After(C.FatalStopTimeout):
		stack := D.GetGoroutineStack(goroutineId)      //karing
		return E.New("close service timeout:" + stack) //karing
		//os.Exit(1) //karing
	}
}

func (s *BoxService) NeedWIFIState() bool {
	return s.instance.Router().NeedWIFIState()
}

var (
	_ platform.Interface = (*platformInterfaceWrapper)(nil)
	_ log.PlatformWriter = (*platformInterfaceWrapper)(nil)
)

type platformInterfaceWrapper struct {
	iif                    PlatformInterface
	useProcFS              bool
	networkManager         adapter.NetworkManager
	myTunName              string
	defaultInterfaceAccess sync.Mutex
	defaultInterface       *control.Interface
	isExpensive            bool
	isConstrained          bool
}

func (w *platformInterfaceWrapper) Initialize(networkManager adapter.NetworkManager) error {
	w.networkManager = networkManager
	return nil
}

func (w *platformInterfaceWrapper) UsePlatformAutoDetectInterfaceControl() bool {
	return w.iif.UsePlatformAutoDetectInterfaceControl()
}

func (w *platformInterfaceWrapper) AutoDetectInterfaceControl(fd int) error {
	return w.iif.AutoDetectInterfaceControl(int32(fd))
}

func (w *platformInterfaceWrapper) OpenTun(options *tun.Options, platformOptions option.TunPlatformOptions) (tun.Tun, error) {
	if len(options.IncludeUID) > 0 || len(options.ExcludeUID) > 0 {
		return nil, E.New("platform: unsupported uid options")
	}
	if len(options.IncludeAndroidUser) > 0 {
		return nil, E.New("platform: unsupported android_user option")
	}
	routeRanges, err := options.BuildAutoRouteRanges(true)
	if err != nil {
		return nil, E.New("build auto_route_ranges") //karing
	}
	tunFd, err := w.iif.OpenTun(&tunOptions{options, routeRanges, platformOptions})
	if err != nil {
		return nil, E.New("opentun") //karing
	}
	options.Name, err = getTunnelName(tunFd)
	if err != nil {
		return nil, E.Cause(err, "query tun name")
	}
	options.InterfaceMonitor.RegisterMyInterface(options.Name)
	dupFd, err := dup(int(tunFd))
	if err != nil {
		return nil, E.Cause(err, "dup tun file descriptor")
	}
	options.FileDescriptor = dupFd
	w.myTunName = options.Name
	tun, err := tun.New(*options) //karing
	if err != nil {               //karing
		return nil, E.Cause(err, "tun.new")
	}
	return tun, err //karing
}

func (w *platformInterfaceWrapper) CreateDefaultInterfaceMonitor(logger logger.Logger) tun.DefaultInterfaceMonitor {
	return &platformDefaultInterfaceMonitor{
		platformInterfaceWrapper: w,
		logger:                   logger,
	}
}

func (w *platformInterfaceWrapper) Interfaces() ([]adapter.NetworkInterface, error) {
	interfaceIterator, err := w.iif.GetInterfaces()
	if err != nil {
		return nil, err
	}
	var interfaces []adapter.NetworkInterface
	for _, netInterface := range iteratorToArray[*NetworkInterface](interfaceIterator) {
		if netInterface.Name == w.myTunName {
			continue
		}
		w.defaultInterfaceAccess.Lock()
		// (GOOS=windows) SA4006: this value of `isDefault` is never used
		// Why not used?
		//nolint:staticcheck
		isDefault := w.defaultInterface != nil && int(netInterface.Index) == w.defaultInterface.Index
		w.defaultInterfaceAccess.Unlock()
		interfaces = append(interfaces, adapter.NetworkInterface{
			Interface: control.Interface{
				Index:     int(netInterface.Index),
				MTU:       int(netInterface.MTU),
				Name:      netInterface.Name,
				Addresses: common.Map(iteratorToArray[string](netInterface.Addresses), netip.MustParsePrefix),
				Flags:     linkFlags(uint32(netInterface.Flags)),
			},
			Type:        C.InterfaceType(netInterface.Type),
			DNSServers:  iteratorToArray[string](netInterface.DNSServer),
			Expensive:   netInterface.Metered || isDefault && w.isExpensive,
			Constrained: isDefault && w.isConstrained,
		})
	}
	return interfaces, nil
}

func (w *platformInterfaceWrapper) UnderNetworkExtension() bool {
	return w.iif.UnderNetworkExtension()
}

func (w *platformInterfaceWrapper) IncludeAllNetworks() bool {
	return w.iif.IncludeAllNetworks()
}

func (w *platformInterfaceWrapper) ClearDNSCache() {
	w.iif.ClearDNSCache()
}

func (w *platformInterfaceWrapper) ReadWIFIState() adapter.WIFIState {
	wifiState := w.iif.ReadWIFIState()
	if wifiState == nil {
		return adapter.WIFIState{}
	}
	return (adapter.WIFIState)(*wifiState)
}

func (w *platformInterfaceWrapper) SystemCertificates() []string {
	return iteratorToArray[string](w.iif.SystemCertificates())
}

func (w *platformInterfaceWrapper) FindProcessInfo(ctx context.Context, network string, source netip.AddrPort, destination netip.AddrPort) (*process.Info, error) {
	var uid int32
	if w.useProcFS {
		uid = procfs.ResolveSocketByProcSearch(network, source, destination)
		if uid == -1 {
			return nil, E.New("procfs: not found")
		}
	} else {
		var ipProtocol int32
		switch N.NetworkName(network) {
		case N.NetworkTCP:
			ipProtocol = syscall.IPPROTO_TCP
		case N.NetworkUDP:
			ipProtocol = syscall.IPPROTO_UDP
		default:
			return nil, E.New("unknown network: ", network)
		}
		var err error
		uid, err = w.iif.FindConnectionOwner(ipProtocol, source.Addr().String(), int32(source.Port()), destination.Addr().String(), int32(destination.Port()))
		if err != nil {
			return nil, err
		}
	}
	packageName, _ := w.iif.PackageNameByUid(uid)
	return &process.Info{UserId: uid, PackageName: packageName}, nil
}

func (w *platformInterfaceWrapper) DisableColors() bool {
	return runtime.GOOS != "android"
}

func (w *platformInterfaceWrapper) WriteMessage(level log.Level, message string) {
	w.iif.WriteLog(message)
}

func (w *platformInterfaceWrapper) SendNotification(notification *platform.Notification) error {
	return w.iif.SendNotification((*Notification)(notification))
}
