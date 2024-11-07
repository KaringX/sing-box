package outbound

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/alitto/pond"
	"github.com/sagernet/sing-box/adapter"
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
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
)

const TimeoutDelay = 65535   //hiddify
const MinFailureToReset = 5 //hiddify

var (
	_ adapter.Outbound                = (*URLTest)(nil)
	_ adapter.OutboundGroup           = (*URLTest)(nil)
	_ adapter.InterfaceUpdateListener = (*URLTest)(nil)
)

type URLTest struct {
	myOutboundAdapter
	ctx                          context.Context
	tags                         []string
	link                         string
	interval                     time.Duration
	tolerance                    uint16
	idleTimeout                  time.Duration
	group                        *URLTestGroup
	interruptExternalConnections bool
	defaultTag                   string //karing
	reTestIfNetworkUpdate        bool   //karing
	parseErr                     error  //karing
}

func NewURLTest(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.URLTestOutboundOptions) (*URLTest, error) { //karing
	outbound := &URLTest{
		myOutboundAdapter: myOutboundAdapter{
			protocol:     C.TypeURLTest,
			network:      []string{N.NetworkTCP, N.NetworkUDP},
			router:       router,
			logger:       logger,
			tag:          tag,
			dependencies: options.Outbounds,
		},
		ctx:                          ctx,
		tags:                         options.Outbounds,
		link:                         options.URL,
		interval:                     time.Duration(options.Interval),
		tolerance:                    options.Tolerance,
		idleTimeout:                  time.Duration(options.IdleTimeout),
		interruptExternalConnections: options.InterruptExistConnections,
		defaultTag:                   options.Default,               //karing
		reTestIfNetworkUpdate:        options.ReTestIfNetworkUpdate, //karing
	}
	if len(outbound.tags) == 0 {
		return outbound, E.New("missing tags")  //karing
	}
	return outbound, nil
}

func (s *URLTest) Start() error {
	if(s.parseErr != nil){ //karing
		return s.parseErr
	}
	outbounds := make([]adapter.Outbound, 0, len(s.tags))
	for i, tag := range s.tags {
		detour, loaded := s.router.Outbound(tag)
		if !loaded {
			return E.New("outbound ", i, " not found: ", tag)
		}
		outbounds = append(outbounds, detour)
	}
	group, err := NewURLTestGroup(
		s.ctx,
		s.router,
		s.logger,
		outbounds,
		s.link,
		s.interval,
		s.tolerance,
		s.idleTimeout,
		s.interruptExternalConnections,
		s.defaultTag, //karing
	)
	if err != nil {
		return err
	}
	s.group = group
	return nil
}

func (s *URLTest) PostStart() error {
	if(s.parseErr != nil){ //karing
		return s.parseErr
	}
	if s.interval < 0 { //karing
		return nil
	}
	s.group.PostStart()
	return nil
}

func (s *URLTest) Close() error {
	if(s.parseErr != nil){ //karing
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

func (s *URLTest) URLTest(ctx context.Context) (map[string]urltest.URLTestResult, error) { //karing
	return s.group.URLTest(ctx)
}

func (s *URLTest) UpdateCheck() { //karing
	s.group.performUpdateCheck()
}

func (s *URLTest) CheckOutbounds() {
	s.group.CheckOutbounds(true)
}

func (s *URLTest) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	if(s.parseErr != nil){ //karing
		return nil, s.parseErr
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
		if outbound == s.group.selectedOutboundUDP { //karing
			s.group.udpConnectionFailureCount.Reset()
		} else if outbound == s.group.selectedOutboundTCP { //karing
			s.group.tcpConnectionFailureCount.Reset()
		}
		history := s.group.history.LoadURLTestHistory(realTag) //karing
		if history != nil && len(history.Err) != 0 {           //karing
			s.group.history.DeleteURLTestHistory(realTag) //karing
		}

		return s.group.interruptGroup.NewConn(conn, interrupt.IsExternalConnectionFromContext(ctx)), nil
	}
	//s.group.history.DeleteURLTestHistory(outbound.Tag())

	if !s.group.pauseManager.IsNetworkPaused() { //karing
		tag := "[" + outbound.Tag() + "] "
		s.logger.ErrorContext(ctx, tag, err)

		if outbound == s.group.selectedOutboundUDP {
			if s.group.udpConnectionFailureCount.IncrementConditionReset(MinFailureToReset) {
				s.logger.Warn("UDP URLTest Outbound ", s.tag, " (", outboundToString(s.group.selectedOutboundUDP), ") failed to connect for ", MinFailureToReset, " times==> test proxies again!")
				s.group.history.StoreURLTestHistory(realTag, &urltest.History{
					Time:  time.Now(),
					Delay: 0,
					Err:   err.Error(),
				})
				s.group.selectedOutboundUDP = nil
				s.group.performUpdateCheck()
				if outbound == s.group.selectedOutboundUDP {
					s.CheckOutbounds()
				}
			}
		} else if outbound == s.group.selectedOutboundTCP {
			if s.group.tcpConnectionFailureCount.IncrementConditionReset(MinFailureToReset) {
				s.logger.Warn("TCP URLTest Outbound ", s.tag, " (", outboundToString(s.group.selectedOutboundTCP), ") failed to connect for ", MinFailureToReset, " times==> test proxies again!")
				s.group.history.StoreURLTestHistory(realTag, &urltest.History{
					Time:  time.Now(),
					Delay: 0,
					Err:   err.Error(),
				})
				s.group.selectedOutboundTCP = nil
				s.group.performUpdateCheck()
				if outbound == s.group.selectedOutboundTCP {
					s.CheckOutbounds()
				}
			}
		}
	}

	return nil, err
}

func (s *URLTest) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if(s.parseErr != nil){ //karing
		return nil, s.parseErr
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
	if err == nil {
		s.group.udpConnectionFailureCount.Reset()
		return s.group.interruptGroup.NewPacketConn(conn, interrupt.IsExternalConnectionFromContext(ctx)), nil
	}
	tag := "[" + outbound.Tag() + "] "
	realTag := RealTag(outbound)                                   //karing
	s.logger.ErrorContext(ctx, tag, err)                           //karing
	s.group.history.StoreURLTestHistory(realTag, &urltest.History{ //karing
		Time:  time.Now(),
		Delay: 0,
		Err:   err.Error(),
	})
	//s.group.history.DeleteURLTestHistory(outbound.Tag())
	if !s.group.pauseManager.IsNetworkPaused() { //karing
		if outbound == s.group.selectedOutboundUDP {
			if s.group.udpConnectionFailureCount.IncrementConditionReset(MinFailureToReset) {
				s.group.selectedOutboundUDP = nil
				s.group.performUpdateCheck()
				if outbound == s.group.selectedOutboundUDP {
					s.CheckOutbounds()
				}
			}
		}
	}
	return nil, err
}

func (s *URLTest) NewConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	if(s.parseErr != nil){ //karing
		return s.parseErr
	}
	ctx = interrupt.ContextWithIsExternalConnection(ctx)
	return NewConnection(ctx, s, conn, metadata)
}

func (s *URLTest) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	if(s.parseErr != nil){ //karing
		return s.parseErr
	}
	ctx = interrupt.ContextWithIsExternalConnection(ctx)
	return NewPacketConnection(ctx, s, conn, metadata)
}

func (s *URLTest) InterfaceUpdated() {
	if s.group.pauseManager.IsNetworkPaused() { //karing
		return
	}
	if !s.reTestIfNetworkUpdate { //karing
		return
	}
	go s.group.CheckOutbounds(true)
}
func (s *URLTest) SetParseErr(err error){ //karing
	s.parseErr = err
}
type URLTestGroup struct {
	ctx                          context.Context
	router                       adapter.Router
	logger                       log.Logger
	outbounds                    []adapter.Outbound
	link                         string
	interval                     time.Duration
	tolerance                    uint16
	idleTimeout                  time.Duration
	history                      *urltest.HistoryStorage
	checking                     atomic.Bool
	pauseManager                 pause.Manager
	selectedOutboundTCP          adapter.Outbound
	selectedOutboundUDP          adapter.Outbound
	interruptGroup               *interrupt.Group
	interruptExternalConnections bool
	defaultTag                   string //karing

	access     sync.Mutex
	ticker     *time.Ticker
	close      chan struct{}
	started    bool
	lastActive atomic.TypedValue[time.Time]

	tcpConnectionFailureCount MinZeroAtomicInt64 //hiddify
	udpConnectionFailureCount MinZeroAtomicInt64 //hiddify
}

func NewURLTestGroup(
	ctx context.Context,
	router adapter.Router,
	logger log.Logger,
	outbounds []adapter.Outbound,
	link string,
	interval time.Duration,
	tolerance uint16,
	idleTimeout time.Duration,
	interruptExternalConnections bool,
	defaultTag string, //karing
) (*URLTestGroup, error) {
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
	var history *urltest.HistoryStorage
	if history = service.PtrFromContext[urltest.HistoryStorage](ctx); history != nil {
	} else if clashServer := router.ClashServer(); clashServer != nil {
		history = clashServer.HistoryStorage()
	} else {
		history = urltest.NewHistoryStorage()
	}
	return &URLTestGroup{
		ctx:                          ctx,
		router:                       router,
		logger:                       logger,
		outbounds:                    outbounds,
		link:                         link,
		interval:                     interval,
		tolerance:                    tolerance,
		idleTimeout:                  idleTimeout,
		history:                      history,
		close:                        make(chan struct{}),
		pauseManager:                 service.FromContext[pause.Manager](ctx),
		interruptGroup:               interrupt.NewGroup(),
		interruptExternalConnections: interruptExternalConnections,
		defaultTag:                   defaultTag, //karing
	}, nil
}

func (g *URLTestGroup) PostStart() {
	g.started = true
	g.lastActive.Store(time.Now())
	g.performUpdateCheck() //karing
	go g.CheckOutbounds(false)
}

func (g *URLTestGroup) Touch() {
	if !g.started {
		return
	}
	if g.ticker != nil {
		g.lastActive.Store(time.Now())
		return
	}
	g.access.Lock()
	defer g.access.Unlock()
	if g.ticker != nil {
		return
	}
	if g.interval < 0 { //karing
		return
	}
	g.ticker = time.NewTicker(g.interval)
	go g.loopCheck()
}

func (g *URLTestGroup) Close() error {
	if g.ticker == nil {
		return nil
	}
	g.ticker.Stop()
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
		for _, detour := range g.outbounds { //karing
			if !common.Contains(detour.Network(), network) {
				continue
			}
			if g.defaultTag != "" && detour.Tag() == g.defaultTag {
				return detour, true
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
			g.access.Unlock()
			return
		}
		g.pauseManager.WaitActive()
		g.CheckOutbounds(false)
	}
}

func (g *URLTestGroup) CheckOutbounds(force bool) {
	_, _ = g.urlTest(g.ctx, force)
}

func (g *URLTestGroup) URLTest(ctx context.Context) (map[string]urltest.URLTestResult, error) { //karing
	return g.urlTest(ctx, false)
}

func (g *URLTestGroup) UpdateCheck() { //karing
	g.performUpdateCheck()
}

func (g *URLTestGroup) urlTest(ctx context.Context, force bool) (map[string]urltest.URLTestResult, error) { //karing
	result := make(map[string]urltest.URLTestResult) //karing
	if g.checking.Swap(true) {
		return result, nil
	}
	defer g.checking.Store(false)
	//b, _ := batch.New(ctx, batch.WithConcurrencyNum[any](10))
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
		p, loaded := g.router.Outbound(realTag)
		if !loaded {
			continue
		}
		group.Submit(func() { //karing
		//b.Go(realTag, func() (any, error) {		
			testCtx, cancel := context.WithTimeout(g.ctx, C.TCPTimeout)
			defer cancel()
			t, _, err := urltest.URLTest(testCtx, g.link, p)
			if err != nil {
				g.logger.Debug("outbound ", tag, " unavailable: ", err)
				//g.history.DeleteURLTestHistory(realTag)
				g.history.StoreURLTestHistory(realTag, &urltest.History{ //karing
					Time:  time.Now(),
					Delay: 0,
					Err:   err.Error(),
				})
			} else {
				g.logger.Debug("outbound ", tag, " available: ", t, "ms")
				g.history.StoreURLTestHistory(realTag, &urltest.History{
					Time:  time.Now(),
					Delay: t,
					Err:   "",
				})
			}
			resultAccess.Lock()
			if err == nil { //karing
				result[tag] = urltest.URLTestResult{Delay: t, Err: ""}
			} else {
				result[tag] = urltest.URLTestResult{Delay: t, Err: err.Error()}
			}
			resultAccess.Unlock()
			//return nil, nil//karing
		})
		count++            //karing
		if count%20 == 0 { //karing
			group.Wait()           //karing
			g.performUpdateCheck() //karing
		} //karing
	}
	pool.StopAndWait() //karing
	//b.Wait()
	gofree.FreeIdleThread()

	g.performUpdateCheck()
	return result, nil
}

func (g *URLTestGroup) performUpdateCheck() {
	var updated bool
	if outbound, exists := g.Select(N.NetworkTCP); outbound != nil && (g.selectedOutboundTCP == nil || (exists && outbound != g.selectedOutboundTCP)) {
		g.selectedOutboundTCP = outbound
		g.tcpConnectionFailureCount.Reset() //hiddify
		updated = true
	}
	if outbound, exists := g.Select(N.NetworkUDP); outbound != nil && (g.selectedOutboundUDP == nil || (exists && outbound != g.selectedOutboundUDP)) {
		g.selectedOutboundUDP = outbound
		g.udpConnectionFailureCount.Reset() //hiddify
		updated = true
	}
	if updated {
		g.interruptGroup.Interrupt(g.interruptExternalConnections)
	}
}

// hiddify
type MinZeroAtomicInt64 struct {
	access sync.Mutex
	count  int64
}

func (m *MinZeroAtomicInt64) Increment() int64 {
	m.access.Lock()
	defer m.access.Unlock()
	if m.count < 0 {
		m.count = 0
	}
	m.count++
	return m.count
}

func (m *MinZeroAtomicInt64) Decrement(useMutex bool) int64 {
	if useMutex {
		m.access.Lock()
		defer m.access.Unlock()
	}
	if m.count > 0 {
		m.count--
	}
	return m.count
}

func (m *MinZeroAtomicInt64) Get(useMutex bool) int64 {
	if useMutex {
		m.access.Lock()
		defer m.access.Unlock()
	}
	return m.count
}

func (m *MinZeroAtomicInt64) Reset() int64 {
	m.access.Lock()
	defer m.access.Unlock()
	m.count = 0
	return m.count
}
func (m *MinZeroAtomicInt64) IncrementConditionReset(condition int64) bool {
	m.access.Lock()
	defer m.access.Unlock()
	m.count++
	if m.count >= condition {
		m.count = 0
		return true
	}
	return false
}

func outboundToString(outbound adapter.Outbound) string {
	if outbound == nil {
		return "<nil>"
	}
	return outbound.Tag()
}

//hiddify