package box

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sagernet/sing-box/adapter"
	D "github.com/sagernet/sing-box/common/debug"
	"github.com/sagernet/sing-box/common/taskmonitor"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/experimental"
	"github.com/sagernet/sing-box/experimental/cachefile"
	"github.com/sagernet/sing-box/experimental/clashapi"
	"github.com/sagernet/sing-box/experimental/libbox/platform"
	"github.com/sagernet/sing-box/inbound"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/outbound"
	"github.com/sagernet/sing-box/route"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
)

var _ adapter.Service = (*Box)(nil)

type Box struct {
	createdAt    time.Time
	router       adapter.Router
	inbounds     []adapter.Inbound
	outbounds    []adapter.Outbound
	logFactory   log.Factory
	logger       log.ContextLogger
	preServices1 map[string]adapter.Service
	preServices2 map[string]adapter.Service
	postServices map[string]adapter.Service
	done         chan struct{}
	QuitSig      chan struct{} //karing
}

type Options struct {
	option.Options
	Context           context.Context
	PlatformInterface platform.Interface
	PlatformLogWriter log.PlatformWriter
}

func New(options Options) (*Box, error) {
	stacks := D.Stacks(false, false) //karing
	if len(stacks) > 0 {  //karing
		for key := range stacks {
			D.MainGoId = key
			break
		}
	}
	quitSig := make(chan struct{}) //karing
	createdAt := time.Now()
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = service.ContextWithDefaultRegistry(ctx)
	ctx = pause.WithDefaultManager(ctx)
	experimentalOptions := common.PtrValueOrDefault(options.Experimental)
	applyDebugOptions(common.PtrValueOrDefault(experimentalOptions.Debug))
	var needCacheFile bool
	var needClashAPI bool
	var needV2RayAPI bool
	if experimentalOptions.CacheFile != nil && experimentalOptions.CacheFile.Enabled || options.PlatformLogWriter != nil {
		needCacheFile = true
	}
	if experimentalOptions.ClashAPI != nil || options.PlatformLogWriter != nil {
		needClashAPI = true
	}
	if experimentalOptions.V2RayAPI != nil && experimentalOptions.V2RayAPI.Listen != "" {
		needV2RayAPI = true
	}
	var defaultLogWriter io.Writer
	if options.PlatformInterface != nil {
		defaultLogWriter = io.Discard
	}
	logFactory, err := log.New(log.Options{
		Context:        ctx,
		Options:        common.PtrValueOrDefault(options.Log),
		Observable:     needClashAPI,
		DefaultWriter:  defaultLogWriter,
		BaseTime:       createdAt,
		PlatformWriter: options.PlatformLogWriter,
	})
	if err != nil {
		return nil, E.Cause(err, "create log factory")
	}
	if needCacheFile { //karing
		cacheFile := service.FromContext[adapter.CacheFile](ctx)
		if cacheFile == nil {
			cacheFile = cachefile.New(ctx, common.PtrValueOrDefault(experimentalOptions.CacheFile))
			service.MustRegister[adapter.CacheFile](ctx, cacheFile)
		}
		err = cacheFile.BeforePreStart()
		if err != nil {
			return nil, E.Cause(err, "cacheFile load failed")
		}
	}
	router, err := route.NewRouter(
		ctx,
		logFactory,
		common.PtrValueOrDefault(options.Route),
		common.PtrValueOrDefault(options.DNS),
		common.PtrValueOrDefault(options.NTP),
		options.Inbounds,
		options.PlatformInterface,
		func (){ //karing
			quitSig <- struct{}{}
		},
	)
	if err != nil {
		return nil, E.Cause(err, "parse route options")
	}
	inbounds := make([]adapter.Inbound, 0, len(options.Inbounds))
	outbounds := make([]adapter.Outbound, 0, len(options.Outbounds))
	for i, inboundOptions := range options.Inbounds {
		var in adapter.Inbound
		var tag string
		if inboundOptions.Tag != "" {
			tag = inboundOptions.Tag
		} else {
			tag = F.ToString(i)
		}
		in, err = inbound.New(
			ctx,
			router,
			logFactory.NewLogger(F.ToString("inbound/", inboundOptions.Type, "[", tag, "]")),
			inboundOptions,
			options.PlatformInterface,
		)
		if err != nil {
			return nil, E.Cause(err, "parse inbound[", i, "][", tag, "]") //karing
		}
		inbounds = append(inbounds, in)
	}
	for i, outboundOptions := range options.Outbounds {
		var out adapter.Outbound
		var tag string
		if outboundOptions.Tag != "" {
			tag = outboundOptions.Tag
		} else {
			tag = F.ToString(i)
		}
		out, err = outbound.New(
			ctx,
			router,
			logFactory.NewLogger(F.ToString("outbound/", outboundOptions.Type, "[", tag, "]")),
			tag,
			outboundOptions)
		if err != nil {
			out.SetParseErr(E.Cause(err, "parse outbound")) //karing
			//return nil, E.Cause(err, "parse outbound[", i, "][", tag, "]") //karing
		}
		outbounds = append(outbounds, out)
	}
	err = router.Initialize(inbounds, outbounds, func() adapter.Outbound {
		out, oErr := outbound.New(ctx, router, logFactory.NewLogger("outbound/direct"), "direct", option.Outbound{Type: "direct", Tag: "default"})
		common.Must(oErr)
		outbounds = append(outbounds, out)
		return out
	})
	if err != nil {
		return nil, err
	}
	if options.PlatformInterface != nil {
		err = options.PlatformInterface.Initialize(ctx, router)
		if err != nil {
			return nil, E.Cause(err, "initialize platform interface")
		}
	}
	preServices1 := make(map[string]adapter.Service)
	preServices2 := make(map[string]adapter.Service)
	postServices := make(map[string]adapter.Service)
	if needCacheFile {
		cacheFile := service.FromContext[adapter.CacheFile](ctx)
		if cacheFile == nil {
			cacheFile = cachefile.New(ctx, common.PtrValueOrDefault(experimentalOptions.CacheFile))
			service.MustRegister[adapter.CacheFile](ctx, cacheFile)
		}
		preServices1["cache file"] = cacheFile
	}
	if needClashAPI {
		clashAPIOptions := common.PtrValueOrDefault(experimentalOptions.ClashAPI)
		clashAPIOptions.ModeList = experimental.CalculateClashModeList(options.Options)
		clashServer, err := experimental.NewClashServer(ctx, router, logFactory.(log.ObservableFactory), clashAPIOptions)
		if err != nil {
			return nil, E.Cause(err, "create clash api server")
		}

		outbound.OutboundHasConnections = func(tag string) bool { //karing
			trafficManager := clashServer.(*clashapi.Server).TrafficManager()
			hasConn := trafficManager.OutboundHasConnections(tag)
			return hasConn
		}
		router.SetClashServer(clashServer)
		preServices2["clash api"] = clashServer
	}
	if needV2RayAPI {
		v2rayServer, err := experimental.NewV2RayServer(logFactory.NewLogger("v2ray-api"), common.PtrValueOrDefault(experimentalOptions.V2RayAPI))
		if err != nil {
			return nil, E.Cause(err, "create v2ray api server")
		}
		router.SetV2RayServer(v2rayServer)
		preServices2["v2ray api"] = v2rayServer
	}
	return &Box{
		router:       router,
		inbounds:     inbounds,
		outbounds:    outbounds,
		createdAt:    createdAt,
		logFactory:   logFactory,
		logger:       logFactory.Logger(),
		preServices1: preServices1,
		preServices2: preServices2,
		postServices: postServices,
		done:         make(chan struct{}),
		QuitSig:      quitSig, //karing
	}, nil
}

func (s *Box) PreStart() error {
	err := s.preStart()
	if err != nil {
		sentry.CaptureException(err) //karing
		// TODO: remove catch error
		defer func() {
			v := recover()
			if v != nil {
				println(err.Error())
				debug.PrintStack()
				panic("panic on early close: " + fmt.Sprint(v))
			}
		}()
		s.Close()
		return err
	}
	s.logger.Info("sing-box pre-started (", F.Seconds(time.Since(s.createdAt).Seconds()), "s)")
	return nil
}

func (s *Box) Start() error {
	err := s.start()
	if err != nil {
		sentry.CaptureException(err) //karing

		// TODO: remove catch error
		defer func() {
			v := recover()
			if v != nil {
				println(err.Error())
				debug.PrintStack()
				println("panic on early start: " + fmt.Sprint(v))
			}
		}()
		s.Close()
		return err
	}
	for i, in := range s.inbounds {  //karing
		var tag string
		if in.Tag() == "" {
			tag = F.ToString(i)
		} else {
			tag = in.Tag()
		}
		err = in.Start()
		if err != nil {
			return E.Cause(err, "initialize inbound/", in.Type(), "[", tag, "]")
		}
	}
	s.logger.Info("started (", F.Seconds(time.Since(s.createdAt).Seconds()), "s) since ", s.createdAt) //karing
	return nil
}

func (s *Box) preStart() error {
	monitor := taskmonitor.New(s.logger, C.StartTimeout)
	monitor.Start("start logger")
	err := s.logFactory.Start()
	monitor.Finish()
	if err != nil {
		return E.Cause(err, "start logger")
	}
	for serviceName, service := range s.preServices1 {
		if preService, isPreService := service.(adapter.PreStarter); isPreService {
			monitor.Start("pre-start ", serviceName)
			err := preService.PreStart()
			monitor.Finish()
			if err != nil {
				return E.Cause(err, "pre-start ", serviceName)
			}
		}
	}
	for serviceName, service := range s.preServices2 {
		if preService, isPreService := service.(adapter.PreStarter); isPreService {
			monitor.Start("pre-start ", serviceName)
			err := preService.PreStart()
			monitor.Finish()
			if err != nil {
				return E.Cause(err, "pre-start ", serviceName)
			}
		}
	}
	err = s.router.PreStart()
	if err != nil {
		return E.Cause(err, "pre-start router")
	}
	err = s.startOutbounds()
	if err != nil {
		return err
	}
	return s.router.Start()
}

func (s *Box) start() error {
	err := s.preStart()
	if err != nil {
		return err
	}
	for serviceName, service := range s.preServices1 {
		err = service.Start()
		if err != nil {
			return E.Cause(err, "start ", serviceName)
		}
	}
	for serviceName, service := range s.preServices2 {
		err = service.Start()
		if err != nil {
			return E.Cause(err, "start ", serviceName)
		}
	}
	/*for i, in := range s.inbounds { //karing
		var tag string
		if in.Tag() == "" {
			tag = F.ToString(i)
		} else {
			tag = in.Tag()
		}
		err = in.Start()
		if err != nil {
			return E.Cause(err, "initialize inbound/", in.Type(), "[", tag, "]")
		}
	}*/
	return s.postStart()
}

func (s *Box) postStart() error {
	for serviceName, service := range s.postServices {
		err := service.Start()
		if err != nil {
			return E.Cause(err, "start ", serviceName)
		}
	}
	for _, outbound := range s.outbounds {
		if lateOutbound, isLateOutbound := outbound.(adapter.PostStarter); isLateOutbound {
			err := lateOutbound.PostStart()
			if err != nil {
				return E.Cause(err, "post-start outbound/", outbound.Tag())
			}
		}
	}

	return s.router.PostStart()
}

func (s *Box) Close() error {
	select {
	case <-s.done:
		return os.ErrClosed
	default:
		close(s.done)
	}
	monitor := taskmonitor.New(s.logger, C.StopTimeout)
	var errors error
	for serviceName, service := range s.postServices {
		monitor.Start("close ", serviceName)
		errors = E.Append(errors, service.Close(), func(err error) error {
			return E.Cause(err, "close ", serviceName)
		})
		monitor.Finish()
	}
	for i, in := range s.inbounds {
		monitor.Start("close inbound/", in.Type(), "[", in.Tag(), "]")  //karing
		errors = E.Append(errors, in.Close(), func(err error) error {
			return E.Cause(err, "close inbound/", in.Type(), "[", i, "]")
		})
		monitor.Finish()
	}
	//karing
	monitor.Start("close router")
	if err := common.Close(s.router); err != nil {
		errors = E.Append(errors, err, func(err error) error {
			return E.Cause(err, "close router")
		})
	}
	monitor.Finish()
	for i, out := range s.outbounds {
		monitor.Start("close outbound/", out.Type(), "[", out.Tag(), "]")  //karing
		errors = E.Append(errors, common.Close(out), func(err error) error {
			return E.Cause(err, "close outbound/", out.Type(), "[", i, "]")
		})
		monitor.Finish()
	}
	//karing
	for serviceName, service := range s.preServices1 {
		monitor.Start("close ", serviceName)
		errors = E.Append(errors, service.Close(), func(err error) error {
			return E.Cause(err, "close ", serviceName)
		})
		monitor.Finish()
	}
	for serviceName, service := range s.preServices2 {
		monitor.Start("close ", serviceName)
		errors = E.Append(errors, service.Close(), func(err error) error {
			return E.Cause(err, "close ", serviceName)
		})
		monitor.Finish()
	}
	if err := common.Close(s.logFactory); err != nil {
		errors = E.Append(errors, err, func(err error) error {
			return E.Cause(err, "close logger")
		})
	}
	close(s.QuitSig) //karing
	return errors
}

func (s *Box) Router() adapter.Router {
	return s.router
}