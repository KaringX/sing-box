package group

import (
	"context"
	"net"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alitto/pond"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/common/gofree"
	"github.com/sagernet/sing-box/common/interrupt"
	"github.com/sagernet/sing-box/common/urltest"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/x/list"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
)

func RegisterURLTest(registry *outbound.Registry) {
	outbound.Register[option.URLTestOutboundOptions](registry, C.TypeURLTest, NewURLTest)
}

var _ adapter.OutboundGroup = (*URLTest)(nil)

type URLTest struct {
	outbound.Adapter
	ctx                          context.Context
	router                       adapter.Router
	outbound                     adapter.OutboundManager
	connection                   adapter.ConnectionManager
	logger                       log.ContextLogger
	tags                         []string
	link                         string
	interval                     time.Duration
	tolerance                    uint16
	idleTimeout                  time.Duration
	group                        *URLTestGroup
	interruptExternalConnections bool
	defaultTag                   string        //karing
	selectedHealthCheckInterval  time.Duration //karing
	reTestIfNetworkUpdate        bool          //karing
	skipTest                     bool          //karing
}

func NewURLTest(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.URLTestOutboundOptions) (adapter.Outbound, error) {
	outbound := &URLTest{
		Adapter:                      outbound.NewAdapter(C.TypeURLTest, tag, []string{N.NetworkTCP, N.NetworkUDP}, options.Outbounds),
		ctx:                          ctx,
		router:                       router,
		outbound:                     service.FromContext[adapter.OutboundManager](ctx),
		connection:                   service.FromContext[adapter.ConnectionManager](ctx),
		logger:                       logger,
		tags:                         options.Outbounds,
		link:                         options.URL,
		interval:                     time.Duration(options.Interval),
		tolerance:                    options.Tolerance,
		idleTimeout:                  time.Duration(options.IdleTimeout),
		interruptExternalConnections: options.InterruptExistConnections,
		defaultTag:                   options.Default,                                    //karing
		selectedHealthCheckInterval:  time.Duration(options.SelectedHealthCheckInterval), //karing
		reTestIfNetworkUpdate:        options.ReTestIfNetworkUpdate,                      //karing
		skipTest:                     options.SkipTest,                                   //karing
	}
	if len(outbound.tags) == 0 {
		return outbound, E.New("missing tags") //karing
	}
	if len(options.Default) > 0 && !slices.Contains(outbound.tags, options.Default) { //karing
		return outbound, E.New("default tag not found")
	}
	return outbound, nil
}

func (s *URLTest) Start() error {
	if s.GetParseErr() != nil { //karing
		return s.GetParseErr()
	}
	outbounds := make([]adapter.Outbound, 0, len(s.tags))
	for i, tag := range s.tags {
		detour, loaded := s.outbound.Outbound(tag)
		if !loaded {
			return E.New("outbound ", i, " not found: ", tag)
		}
		outbounds = append(outbounds, detour)
	}
	group, err := NewURLTestGroup(s.ctx, s.outbound, s.logger, outbounds, s.link, s.interval, s.tolerance, s.idleTimeout, s.interruptExternalConnections, s.defaultTag, s.selectedHealthCheckInterval, s.skipTest) //karing
	if err != nil {
		return err
	}
	s.group = group
	return nil
}

func (s *URLTest) PostStart() error {
	if s.GetParseErr() != nil { //karing
		return s.GetParseErr()
	}
	if s.interval < 0 { //karing
		return nil
	}
	s.group.PostStart()
	return nil
}

func (s *URLTest) Close() error {
	if s.GetParseErr() != nil { //karing
		return nil
	}
	return common.Close(
		common.PtrOrNil(s.group),
	)
}

func (s *URLTest) Now() string {
	s.group.access.Lock()
	selectedTCP := s.group.selectedOutboundTCP
	selectedUDP := s.group.selectedOutboundUDP
	s.group.access.Unlock()

	if selectedTCP != nil { //karing
		return selectedTCP.Tag() //karing
	} else if selectedUDP != nil { //karing
		return selectedUDP.Tag() //karing
	}
	return ""
}

func (s *URLTest) All() []string {
	return s.tags
}

func (s *URLTest) URLTest(ctx context.Context, force bool) (map[string]adapter.URLTestResult, error) { //karing
	return s.group.URLTest(ctx, force) //karing
}

func (s *URLTest) CheckOutbounds() {
	s.group.CheckOutbounds(true)
}

func (s *URLTest) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	if s.GetParseErr() != nil { //karing
		return nil, s.GetParseErr()
	}
	s.group.Touch()
	var selectedOutbound adapter.Outbound //karing
	switch N.NetworkName(network) {
	case N.NetworkTCP:
		s.group.access.Lock()
		selectedOutbound = s.group.selectedOutboundTCP //karing
		s.group.access.Unlock()
	case N.NetworkUDP:
		s.group.access.Lock()
		selectedOutbound = s.group.selectedOutboundUDP //karing
		s.group.access.Unlock()
	default:
		return nil, E.Extend(N.ErrUnknownNetwork, network)
	}
	if selectedOutbound == nil { //karing
		selectedOutbound, _ = s.group.Select(network) //karing
	}
	if selectedOutbound == nil { //karing
		return nil, E.New("missing supported outbound")
	}
	conn, err := selectedOutbound.DialContext(ctx, network, destination) //karing
	realTag := RealTag(selectedOutbound)                                 //karing
	if err == nil {
		s.updateHistory(selectedOutbound.Type(), realTag) //karing
		return s.group.interruptGroup.NewConn(conn, interrupt.IsExternalConnectionFromContext(ctx)), nil
	}

	s.logger.ErrorContext(ctx, "DialContext ["+selectedOutbound.Tag()+"] :", err) //karing
	//s.group.history.DeleteURLTestHistory(selectedOutbound.Tag()) //karing
	s.group.history.StoreURLTestHistory(realTag, &adapter.URLTestHistory{ //karing
		Time:  time.Now(),
		Delay: 0,
		Err:   err.Error(),
	})
	s.recheckSelectedOutboundUDP(selectedOutbound, "DialContext") //karing
	s.recheckSelectedOutboundTCP(selectedOutbound)                //karing

	return nil, err
}

func (s *URLTest) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if s.GetParseErr() != nil { //karing
		return nil, s.GetParseErr()
	}
	s.group.Touch()
	s.group.access.Lock()
	selectedOutbound := s.group.selectedOutboundUDP //karing
	s.group.access.Unlock()
	if selectedOutbound == nil { //karing
		selectedOutbound, _ = s.group.Select(N.NetworkUDP) //karing
	}
	if selectedOutbound == nil { //karing
		return nil, E.New("missing supported outbound")
	}
	conn, err := selectedOutbound.ListenPacket(ctx, destination) //karing
	realTag := RealTag(selectedOutbound)                         //karing
	if err == nil {
		s.updateHistory(selectedOutbound.Type(), realTag) //karing
		return s.group.interruptGroup.NewPacketConn(conn, interrupt.IsExternalConnectionFromContext(ctx)), nil
	}

	s.logger.ErrorContext(ctx, "ListenPacket ["+selectedOutbound.Tag()+"] :", err) //karing
	//s.group.history.DeleteURLTestHistory(selectedOutbound.Tag()) //karing
	s.group.history.StoreURLTestHistory(realTag, &adapter.URLTestHistory{ //karing
		Time:  time.Now(),
		Delay: 0,
		Err:   err.Error(),
	})
	s.recheckSelectedOutboundUDP(selectedOutbound, "ListenPacket") //karing

	return nil, err
}

func (s *URLTest) NewConnectionEx(ctx context.Context, conn net.Conn, metadata adapter.InboundContext, onClose N.CloseHandlerFunc) {
	if s.GetParseErr() != nil { //karing
		return
	}
	ctx = interrupt.ContextWithIsExternalConnection(ctx)
	s.connection.NewConnection(ctx, s, conn, metadata, onClose)
}

func (s *URLTest) NewPacketConnectionEx(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext, onClose N.CloseHandlerFunc) {
	if s.GetParseErr() != nil { //karing
		return
	}
	ctx = interrupt.ContextWithIsExternalConnection(ctx)
	s.connection.NewPacketConnection(ctx, s, conn, metadata, onClose)
}

type URLTestGroup struct {
	ctx                          context.Context
	router                       adapter.Router
	outbound                     adapter.OutboundManager
	pause                        pause.Manager
	pauseCallback                *list.Element[pause.Callback]
	logger                       log.ContextLogger //karing
	outbounds                    []adapter.Outbound
	link                         string
	interval                     time.Duration
	tolerance                    uint16
	idleTimeout                  time.Duration
	history                      adapter.URLTestHistoryStorage
	checking                     atomic.Bool
	selectedOutboundTCP          adapter.Outbound
	selectedOutboundUDP          adapter.Outbound
	interruptGroup               *interrupt.Group
	interruptExternalConnections bool

	defaultTag                  string          //karing
	selectedHealthCheckInterval time.Duration   //karing
	healthChecking              map[string]bool //karing
	selectedHealthCheckTicker   *time.Ticker    //karing
	skipTest                    bool            //karing
	testTimes                   int             //karing
	access                      sync.Mutex
	ticker                      *time.Ticker
	close                       chan struct{}
	started                     bool
	lastActive                  common.TypedValue[time.Time]
}

func NewURLTestGroup(ctx context.Context, outboundManager adapter.OutboundManager, logger log.ContextLogger, outbounds []adapter.Outbound, link string, interval time.Duration, tolerance uint16, idleTimeout time.Duration, interruptExternalConnections bool, defaultTag string, selectedHealthCheckInterval time.Duration, skipTest bool) (*URLTestGroup, error) { //karing
	if interval == 0 {
		interval = C.DefaultURLTestInterval
	}
	//if tolerance == 0 {//karing
	//tolerance = 50  //karing
	//}//karing
	if idleTimeout == 0 {
		idleTimeout = C.DefaultURLTestIdleTimeout
	}
	if interval > idleTimeout {
		return nil, E.New("interval must be less or equal than idle_timeout")
	}
	var history adapter.URLTestHistoryStorage
	if historyFromCtx := service.PtrFromContext[urltest.HistoryStorage](ctx); historyFromCtx != nil {
		history = historyFromCtx
	} else if clashServer := service.FromContext[adapter.ClashServer](ctx); clashServer != nil {
		history = clashServer.HistoryStorage()
	} else {
		history = urltest.NewHistoryStorage()
	}
	return &URLTestGroup{
		ctx:                          ctx,
		outbound:                     outboundManager,
		logger:                       logger,
		outbounds:                    outbounds,
		link:                         link,
		interval:                     interval,
		tolerance:                    tolerance,
		idleTimeout:                  idleTimeout,
		history:                      history,
		close:                        make(chan struct{}),
		pause:                        service.FromContext[pause.Manager](ctx),
		interruptGroup:               interrupt.NewGroup(),
		interruptExternalConnections: interruptExternalConnections,
		defaultTag:                   defaultTag,                  //karing
		selectedHealthCheckInterval:  selectedHealthCheckInterval, //karing
		healthChecking:               make(map[string]bool),       //karing
		skipTest:                     skipTest,                    //karing
	}, nil
}

func (g *URLTestGroup) PostStart() {
	g.access.Lock()
	defer g.access.Unlock()
	g.started = true
	g.lastActive.Store(time.Now())
	g.performUpdateCheck(false) //karing
	go g.CheckOutbounds(false)
}

func (g *URLTestGroup) Touch() {
	if !g.started {
		return
	}
	g.access.Lock()
	defer g.access.Unlock()
	if g.interval < 0 { //karing
		return
	}
	if g.ticker != nil {
		g.lastActive.Store(time.Now())
		return
	}
	g.ticker = time.NewTicker(g.interval)
	go g.loopCheck()
	g.pauseCallback = pause.RegisterTicker(g.pause, g.ticker, g.interval, nil)
}

func (g *URLTestGroup) Close() error {
	g.access.Lock()
	defer g.access.Unlock()
	if g.selectedHealthCheckTicker != nil { //karing
		g.selectedHealthCheckTicker.Stop()
	}
	if g.ticker == nil {
		return nil
	}
	g.ticker.Stop()
	g.pause.UnregisterCallback(g.pauseCallback)
	close(g.close)
	g.outbounds = make([]adapter.Outbound, 0) //karing
	g.selectedOutboundTCP = nil               //karing
	g.selectedOutboundUDP = nil               //karing
	return nil
}

func (g *URLTestGroup) Select(network string) (adapter.Outbound, bool) {
	var minDelay uint16
	var minOutbound adapter.Outbound
	switch network {
	case N.NetworkTCP:
		selectOutbound := g.selectedOutboundTCP //karing
		if selectOutbound != nil {              //karing
			if history := g.history.LoadURLTestHistory(RealTag(selectOutbound)); history != nil { //karing
				minOutbound = selectOutbound //karing
				minDelay = history.Delay
			}
		}
	case N.NetworkUDP:
		selectOutbound := g.selectedOutboundUDP //karing
		if selectOutbound != nil {
			if history := g.history.LoadURLTestHistory(RealTag(selectOutbound)); history != nil { //karing
				minOutbound = selectOutbound //karing
				minDelay = history.Delay
			}
		}
	}
	if len(g.outbounds) == 0 { //karing
		return nil, false
	}
	for _, detour := range g.outbounds {
		if !common.Contains(detour.Network(), network) {
			continue
		}
		history := g.history.LoadURLTestHistory(RealTag(detour))
		if history == nil {
			continue
		}

		if len(history.Err) != 0 { //karing
			continue
		}
		if minDelay == 0 || minDelay >= history.Delay+g.tolerance { //karing
			minDelay = history.Delay
			minOutbound = detour
		}
	}
	if minOutbound == nil {
		if g.defaultTag != "" { //karing
			for _, detour := range g.outbounds {
				if !common.Contains(detour.Network(), network) {
					continue
				}
				if detour.Tag() == g.defaultTag {
					return detour, false
				}
			}
		}

		for _, detour := range g.outbounds {
			if !common.Contains(detour.Network(), network) {
				continue
			}
			return detour, false
		}
		return nil, false
	}
	return minOutbound, true
}

func (g *URLTestGroup) loopCheck() {
	if time.Since(g.lastActive.Load()) > g.interval {
		g.lastActive.Store(time.Now())
		g.CheckOutbounds(false)
	}
	for {
		select {
		case <-g.close:
			return
		case <-g.ticker.C:
		}
		if time.Since(g.lastActive.Load()) > g.idleTimeout {
			g.access.Lock()
			g.ticker.Stop()
			g.ticker = nil
			g.pause.UnregisterCallback(g.pauseCallback)
			g.pauseCallback = nil
			g.access.Unlock()
			return
		}
		g.CheckOutbounds(false)
	}
}

func (g *URLTestGroup) CheckOutbounds(force bool) {
	_, _ = g.urlTest(g.ctx, force)
}

func (g *URLTestGroup) URLTest(ctx context.Context, force bool) (map[string]adapter.URLTestResult, error) { //karing
	return g.urlTest(ctx, force)
}

func (g *URLTestGroup) urlTest(ctx context.Context, force bool) (map[string]adapter.URLTestResult, error) { //karing
	result := make(map[string]adapter.URLTestResult) //karing
	if g.checking.Swap(true) {
		return result, nil
	}
	defer g.checking.Store(false)
	if g.skipTest { //karing
		if g.testTimes != 0 {
			g.performUpdateCheck(false)
			return result, nil
		}
	}
	g.testTimes++
	//b, _ := batch.New(ctx, batch.WithConcurrencyNum[any](10)) //karing
	pool := pond.New(10, 20) //karing
	group := pool.Group()    //karing
	count := 0               //karing

	checked := make(map[string]bool)
	var resultAccess sync.Mutex
	for _, detour := range g.outbounds {
		tag := detour.Tag()
		realTag := RealTag(detour)
		if checked[realTag] {
			continue
		}
		history := g.history.LoadURLTestHistory(realTag)
		if !force && history != nil && time.Since(history.Time) < g.interval {
			continue
		}
		checked[realTag] = true
		p, loaded := g.outbound.Outbound(realTag)
		if !loaded {
			continue
		}
		pauseManager := service.FromContext[pause.Manager](g.ctx) //karing
		if pauseManager == nil {
			return result, nil
		}
		if pauseManager.IsNetworkPaused() { //karing
			pauseManager.WaitActive()
		}
		group.Submit(func() { //karing
			//b.Go(realTag, func() (any, error) {
			testCtx, cancel := context.WithTimeout(g.ctx, C.TCPTimeout)
			defer cancel()
			t, _, err := urltest.URLTest(testCtx, g.link, p)
			pauseManager := service.FromContext[pause.Manager](g.ctx) //karing
			if pauseManager == nil {
				return
			}
			if err != nil {
				g.logger.DebugContext(g.ctx, "outbound ", tag, " unavailable: ", err) //karing
				//g.history.DeleteURLTestHistory(realTag)
				g.history.StoreURLTestHistory(realTag, &adapter.URLTestHistory{ //karing
					Time:  time.Now(),
					Delay: 0,
					Err:   err.Error(),
				})
			} else {
				g.logger.DebugContext(g.ctx, "outbound ", tag, " available: ", t, "ms") //karing
				g.history.StoreURLTestHistory(realTag, &adapter.URLTestHistory{
					Time:  time.Now(),
					Delay: t,
					Err:   "",
				})
			}
			resultAccess.Lock()
			if err == nil { //karing
				result[tag] = adapter.URLTestResult{Delay: t, Err: ""}
			} else {
				result[tag] = adapter.URLTestResult{Delay: 0, Err: err.Error()}
			}
			resultAccess.Unlock()
			//return nil, nil//karing
		})
		count++
		pauseManager = service.FromContext[pause.Manager](g.ctx) //karing
		if pauseManager == nil {
			return result, nil
		} //karing
		if count%10 == 0 || count == len(g.outbounds) { //karing
			group.Wait()                //karing
			g.performUpdateCheck(false) //karing
		} //karing
	}

	//b.Wait() //karing
	pool.StopAndWait()          //karing
	gofree.FreeIdleThread()     //karing
	g.performUpdateCheck(false) //karing
	return result, nil
}

func (g *URLTestGroup) performUpdateCheck(retestGroupIfAllFailed bool) {
	var updated bool
	var retest bool

	if outbound, exists := g.Select(N.NetworkTCP); outbound != nil && (g.selectedOutboundTCP == nil || (exists && outbound != g.selectedOutboundTCP)) {
		g.access.Lock()
		if g.selectedOutboundTCP != nil {
			updated = true
		}
		g.selectedOutboundTCP = outbound
		g.access.Unlock()
		retest = !exists //karing
	}
	if outbound, exists := g.Select(N.NetworkUDP); outbound != nil && (g.selectedOutboundUDP == nil || (exists && outbound != g.selectedOutboundUDP)) {
		g.access.Lock()
		if g.selectedOutboundUDP != nil {
			updated = true
		}
		g.selectedOutboundUDP = outbound
		g.access.Unlock()
		if !retest { //karing
			retest = !exists
		}
	}
	g.access.Lock()
	selectedTCP := g.selectedOutboundTCP
	selectedUDP := g.selectedOutboundUDP
	g.access.Unlock()

	if updated {
		g.interruptGroup.Interrupt(g.interruptExternalConnections)
	}
	if (selectedTCP != nil || selectedUDP != nil) && g.selectedHealthCheckTicker == nil && g.selectedHealthCheckInterval != 0 { //karing
		g.selectedHealthCheckTicker = time.NewTicker(g.selectedHealthCheckInterval)
		go g.loopHealthCheckSelected()
	}
	if retestGroupIfAllFailed && retest { //karing
		pauseManager := service.FromContext[pause.Manager](g.ctx)
		if pauseManager != nil && !pauseManager.IsNetworkPaused() && !pauseManager.IsDevicePaused() {
			g.logger.WarnContext(g.ctx, "URLTest performUpdateCheck need retest")
			go g.CheckOutbounds(true)
		}
	}
}
