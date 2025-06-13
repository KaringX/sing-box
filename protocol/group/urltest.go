package group

import (
	"context"
	"net"
	"slices"
	"sync"
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
	"github.com/sagernet/sing/common/atomic"
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
	group, err := NewURLTestGroup(s.ctx, s.outbound, s.logger, outbounds, s.link, s.interval, s.tolerance, s.idleTimeout, s.interruptExternalConnections, s.defaultTag, s.selectedHealthCheckInterval) //karing
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
	if s.group.selectedOutboundTCP != nil {
		return s.group.selectedOutboundTCP.Tag()
	} else if s.group.selectedOutboundUDP != nil {
		return s.group.selectedOutboundUDP.Tag()
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
	var outbound adapter.Outbound
	switch N.NetworkName(network) {
	case N.NetworkTCP:
		outbound = s.group.selectedOutboundTCP
	case N.NetworkUDP:
		outbound = s.group.selectedOutboundUDP
	default:
		return nil, E.Extend(N.ErrUnknownNetwork, network)
	}
	if outbound == nil {
		outbound, _ = s.group.Select(network)
	}
	if outbound == nil {
		return nil, E.New("missing supported outbound")
	}
	conn, err := outbound.DialContext(ctx, network, destination)
	realTag := RealTag(outbound) //karing
	if err == nil {
		s.updateHistory(outbound.Type(), realTag) //karing
		return s.group.interruptGroup.NewConn(conn, interrupt.IsExternalConnectionFromContext(ctx)), nil
	}

	s.logger.ErrorContext(ctx, "["+outbound.Tag()+"] when URLTest.DialContext", err) //karing
	//s.group.history.DeleteURLTestHistory(outbound.Tag()) //karing
	s.group.history.StoreURLTestHistory(realTag, &adapter.URLTestHistory{ //karing
		Time:  time.Now(),
		Delay: 0,
		Err:   err.Error(),
	})
	s.recheckSelectedOutboundUDP(outbound, "DialContext") //karing
	s.recheckSelectedOutboundTCP(outbound)                //karing

	return nil, err
}

func (s *URLTest) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if s.GetParseErr() != nil { //karing
		return nil, s.GetParseErr()
	}
	s.group.Touch()
	outbound := s.group.selectedOutboundUDP
	if outbound == nil {
		outbound, _ = s.group.Select(N.NetworkUDP)
	}
	if outbound == nil {
		return nil, E.New("missing supported outbound")
	}
	conn, err := outbound.ListenPacket(ctx, destination)
	realTag := RealTag(outbound) //karing
	if err == nil {
		s.updateHistory(outbound.Type(), realTag) //karing
		return s.group.interruptGroup.NewPacketConn(conn, interrupt.IsExternalConnectionFromContext(ctx)), nil
	}

	s.logger.ErrorContext(ctx, "["+outbound.Tag()+"] when URLTest.ListenPacket", err) //karing
	//s.group.history.DeleteURLTestHistory(outbound.Tag()) //karing
	s.group.history.StoreURLTestHistory(realTag, &adapter.URLTestHistory{ //karing
		Time:  time.Now(),
		Delay: 0,
		Err:   err.Error(),
	})
	s.recheckSelectedOutboundUDP(outbound, "ListenPacket") //karing

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

func (s *URLTest) UpdateCheck() { //karing
	s.group.performUpdateCheck(false)
}

func (s *URLTest) Checking() bool { //karing
	return s.group.Checking()
}

func (s *URLTest) recheckSelectedOutboundTCP(outbound adapter.Outbound) { //karing
	if outbound == s.group.selectedOutboundTCP {
		s.logger.Warn("URLTest TCP failed: ", s.Tag(), " (", s.outboundToString(s.group.selectedOutboundTCP), "), will performUpdateCheck")
		s.group.selectedOutboundTCP = nil
		s.group.performUpdateCheck((outbound == s.group.selectedOutboundTCP) && !s.Checking())
	}
}
func (s *URLTest) recheckSelectedOutboundUDP(outbound adapter.Outbound, from string) { //karing
	if outbound == s.group.selectedOutboundUDP {
		s.logger.Warn("URLTest UDP ", from, " failed: ", s.Tag(), " (", s.outboundToString(s.group.selectedOutboundUDP), "), will performUpdateCheck")
		s.group.selectedOutboundUDP = nil
		s.group.performUpdateCheck((outbound == s.group.selectedOutboundUDP) && !s.Checking())
	}
}

func (s *URLTest) updateHistory(outboundType string, realTag string) { //karing
	if !s.group.isProxyOutbound(outboundType) {
		return
	}
	if s.group.IsHealthChecking(realTag) {
		return
	}
	history := s.group.history.LoadURLTestHistory(realTag)
	if (history == nil) || len(history.Err) != 0 {
		s.group.history.DeleteURLTestHistory(realTag)
		s.group.HealthCheck(realTag)
	}
}

func (s *URLTest) InterfaceUpdated() { //karing
	pauseManager := service.FromContext[pause.Manager](s.ctx)
	if pauseManager == nil {
		return
	}
	if pauseManager.IsNetworkPaused() {
		return
	}
	if !s.reTestIfNetworkUpdate {
		return
	}
	go s.group.CheckOutbounds(true)
}

func (s *URLTest) outboundToString(outbound adapter.Outbound) string { //karing
	if outbound == nil {
		return "<nil>"
	}
	return outbound.Tag()
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
	access                      sync.Mutex
	ticker                      *time.Ticker
	close                       chan struct{}
	started                     bool
	lastActive                  atomic.TypedValue[time.Time]
}

func NewURLTestGroup(ctx context.Context, outboundManager adapter.OutboundManager, logger log.ContextLogger, outbounds []adapter.Outbound, link string, interval time.Duration, tolerance uint16, idleTimeout time.Duration, interruptExternalConnections bool, defaultTag string, selectedHealthCheckInterval time.Duration) (*URLTestGroup, error) { //karing
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
	return nil
}

func (g *URLTestGroup) Select(network string) (adapter.Outbound, bool) {
	var minDelay uint16
	var minOutbound adapter.Outbound
	switch network {
	case N.NetworkTCP:
		if g.selectedOutboundTCP != nil {
			if history := g.history.LoadURLTestHistory(RealTag(g.selectedOutboundTCP)); history != nil {
				minOutbound = g.selectedOutboundTCP
				minDelay = history.Delay
			}
		}
	case N.NetworkUDP:
		if g.selectedOutboundUDP != nil {
			if history := g.history.LoadURLTestHistory(RealTag(g.selectedOutboundUDP)); history != nil {
				minOutbound = g.selectedOutboundUDP
				minDelay = history.Delay
			}
		}
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
	if time.Now().Sub(g.lastActive.Load()) > g.interval {
		g.lastActive.Store(time.Now())
		g.CheckOutbounds(false)
	}
	for {
		select {
		case <-g.close:
			return
		case <-g.ticker.C:
		}
		if time.Now().Sub(g.lastActive.Load()) > g.idleTimeout {
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

func (g *URLTestGroup) Checking() bool { //karing
	return g.checking.Load()
}

func (g *URLTestGroup) CheckOutbounds(force bool) {
	_, _ = g.urlTest(g.ctx, force)
}

func (g *URLTestGroup) URLTest(ctx context.Context, force bool) (map[string]adapter.URLTestResult, error) { //karing
	return g.urlTest(ctx, force)
}

func (g *URLTestGroup) UpdateCheck() { //karing
	g.performUpdateCheck(false)
}

func (g *URLTestGroup) IsHealthChecking(realTag string) bool { //karing
	g.access.Lock()
	defer g.access.Unlock()
	_, ok := g.healthChecking[realTag]
	return ok
}

func (g *URLTestGroup) HealthCheck(realTag string) { //karing
	if g.Checking() {
		return
	}
	pauseManager := service.FromContext[pause.Manager](g.ctx)
	if pauseManager == nil {
		return
	}
	if pauseManager.IsNetworkPaused() || pauseManager.IsDevicePaused() {
		return
	}
	if outbound.OutboundHasConnections != nil {
		has, uploadLast, downloadLast := outbound.OutboundHasConnections(realTag)
		if !has || uploadLast.IsZero() {
			return
		}
		interval := time.Since(downloadLast).Seconds()
		if interval <= 5 {
			return
		}
	}

	g.access.Lock()
	defer g.access.Unlock()
	if _, ok := g.healthChecking[realTag]; !ok {
		g.healthChecking[realTag] = true
		go func() {
			ctx, cancel := context.WithTimeout(g.ctx, C.TCPTimeout)
			defer cancel()
			p, loaded := g.outbound.Outbound(realTag)
			if !loaded {
				g.access.Lock()
				delete(g.healthChecking, realTag)
				g.access.Unlock()
				return
			}
			t, _, err := urltest.URLTest(ctx, g.link, p)
			if err == nil {
				history := g.history.LoadURLTestHistory(realTag)
				if (history == nil) || len(history.Err) != 0 {
					g.history.StoreURLTestHistory(realTag, &adapter.URLTestHistory{
						Time:  time.Now(),
						Delay: t,
						Err:   "",
					})
				}
			} else {
				g.history.StoreURLTestHistory(realTag, &adapter.URLTestHistory{
					Time:  time.Now(),
					Delay: 0,
					Err:   err.Error(),
				})
				g.performUpdateCheck(true)
			}

			g.access.Lock()
			delete(g.healthChecking, realTag)
			g.access.Unlock()
		}()
	}
}

func (g *URLTestGroup) HealthCheckSelected() { //karing
	tags := make(map[string]bool)
	if g.selectedOutboundTCP != nil && g.isProxyOutbound(g.selectedOutboundTCP.Type()) {
		tags[RealTag(g.selectedOutboundTCP)] = true
	}
	if g.selectedOutboundUDP != nil && g.isProxyOutbound(g.selectedOutboundUDP.Type()) {
		tags[RealTag(g.selectedOutboundUDP)] = true
	}
	for tag := range tags {
		g.HealthCheck(tag)
	}
}

func (g *URLTestGroup) isProxyOutbound(outboundType string) bool { //karing
	switch outboundType {
	case C.TypeSOCKS:
		return true
	case C.TypeHTTP:
		return true
	case C.TypeShadowsocks:
		return true
	case C.TypeVMess:
		return true
	case C.TypeTrojan:
		return true
	case C.TypeWireGuard:
		return true
	case C.TypeHysteria:
		return true
	case C.TypeTor:
		return true
	case C.TypeSSH:
		return true
	case C.TypeShadowTLS:
		return true
	case C.TypeAnyTLS:
		return true
	case C.TypeShadowsocksR:
		return true
	case C.TypeVLESS:
		return true
	case C.TypeTUIC:
		return true
	case C.TypeHysteria2:
		return true
	case C.TypeTailscale:
		return true
	default:
		return false
	}
}

func (g *URLTestGroup) loopHealthCheckSelected() {
	for {
		select {
		case <-g.close:
			return
		case <-g.selectedHealthCheckTicker.C:
		}
		g.HealthCheckSelected()
	}
}

func (g *URLTestGroup) urlTest(ctx context.Context, force bool) (map[string]adapter.URLTestResult, error) { //karing
	result := make(map[string]adapter.URLTestResult) //karing
	if g.checking.Swap(true) {
		return result, nil
	}
	defer g.checking.Store(false)
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
		if !force && history != nil && time.Now().Sub(history.Time) < g.interval {
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
				g.tryInterfaceUpdated(detour, realTag)
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
		if g.selectedOutboundTCP != nil {
			updated = true
		}
		g.selectedOutboundTCP = outbound
		retest = !exists //karing
	}
	if outbound, exists := g.Select(N.NetworkUDP); outbound != nil && (g.selectedOutboundUDP == nil || (exists && outbound != g.selectedOutboundUDP)) {
		if g.selectedOutboundUDP != nil {
			updated = true
		}
		g.selectedOutboundUDP = outbound
		if !retest { //karing
			retest = !exists
		}
	}
	if updated {
		g.interruptGroup.Interrupt(g.interruptExternalConnections)
	}
	if (g.selectedOutboundTCP != nil || g.selectedOutboundUDP != nil) && g.selectedHealthCheckTicker == nil && g.selectedHealthCheckInterval != 0 { //karing
		g.selectedHealthCheckTicker = time.NewTicker(g.selectedHealthCheckInterval)
		go g.loopHealthCheckSelected()
	}
	if retestGroupIfAllFailed && retest { //karing
		pauseManager := service.FromContext[pause.Manager](g.ctx)
		if pauseManager != nil && !pauseManager.IsNetworkPaused() && !pauseManager.IsDevicePaused() {
			g.logger.WarnContext(g.ctx, "URLTest performUpdateCheck need retest")
			g.CheckOutbounds(true)
		}
	}
}

func (g *URLTestGroup) tryInterfaceUpdated(detour adapter.Outbound, realTag string) { //karing
	listener, isListener := detour.(adapter.InterfaceUpdateListener)
	needUpdate := detour.Type() == C.TypeHysteria || detour.Type() == C.TypeHysteria2 || detour.Type() == C.TypeTUIC
	if isListener && needUpdate {
		if outbound.OutboundHasConnections != nil {
			has, _, _ := outbound.OutboundHasConnections(realTag)
			if !has {
				listener.InterfaceUpdated()
			}
		}
	}
}
