// karing
package group

import (
	"context"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/common/urltest"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
)

func (s *URLTest) UpdateCheck() {
	s.group.performUpdateCheck(false)
}

func (s *URLTest) Checking() bool {
	return s.group.Checking()
}

func (s *URLTest) recheckSelectedOutboundTCP(selectedOutbound adapter.Outbound) {
	if selectedOutbound == s.group.selectedOutboundTCP {
		s.logger.Warn("URLTest TCP failed: ", s.Tag(), " (", s.outboundToString(selectedOutbound), "), will performUpdateCheck")
		s.group.selectedOutboundTCP = nil
		s.group.performUpdateCheck(!s.Checking())
	}
}
func (s *URLTest) recheckSelectedOutboundUDP(selectedOutbound adapter.Outbound, from string) {
	if selectedOutbound == s.group.selectedOutboundUDP {
		s.logger.Warn("URLTest UDP ", from, " failed: ", s.Tag(), " (", s.outboundToString(selectedOutbound), "), will performUpdateCheck")
		s.group.selectedOutboundUDP = nil
		s.group.performUpdateCheck(!s.Checking())
	}
}

func (s *URLTest) updateHistory(outboundType string, realTag string) {
	if !s.group.isProxyOutbound(outboundType) {
		return
	}

	history := s.group.history.LoadURLTestHistory(realTag)
	if (history == nil) || len(history.Err) != 0 {
		s.group.history.DeleteURLTestHistory(realTag)
		s.group.HealthCheck(realTag, true)
	}
}

func (s *URLTest) InterfaceUpdated() {
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
	if s.group == nil {
		return
	}
	go s.group.CheckOutbounds(true)
}

func (g *URLTestGroup) Checking() bool {
	return g.checking.Load()
}

func (s *URLTest) outboundToString(selectedOutbound adapter.Outbound) string {
	if selectedOutbound == nil {
		return "<nil>"
	}
	return selectedOutbound.Tag()
}

func (g *URLTestGroup) UpdateCheck() {
	g.performUpdateCheck(false)
}

func (g *URLTestGroup) IsHealthChecking(realTag string) bool {
	g.access.Lock()
	defer g.access.Unlock()
	_, ok := g.healthChecking[realTag]
	return ok
}

func (g *URLTestGroup) HealthCheck(realTag string, skipActiveConnectionCheck bool) {
	if g.Checking() {
		return
	}
	if g.IsHealthChecking(realTag) {
		return
	}
	pauseManager := service.FromContext[pause.Manager](g.ctx)
	if pauseManager == nil {
		return
	}
	if pauseManager.IsNetworkPaused() || pauseManager.IsDevicePaused() {
		return
	}
	if outbound.OutboundGetLatestDownloadActiveConnection != nil && !skipActiveConnectionCheck {
		has, downloadLatest := outbound.OutboundGetLatestDownloadActiveConnection(realTag)
		if !has {
			return
		}
		interval := time.Since(downloadLatest).Seconds()
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
			if err != nil {
				g.outboundInterfaceUpdated(p)
				t, _, err = urltest.URLTest(ctx, g.link, p)
			}
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

func (g *URLTestGroup) HealthCheckSelected() {
	tags := make(map[string]bool)
	selectedOutboundTCP := g.selectedOutboundTCP
	selectedOutboundUDP := g.selectedOutboundUDP
	if selectedOutboundTCP != nil && g.isProxyOutbound(selectedOutboundTCP.Type()) {
		tags[RealTag(selectedOutboundTCP)] = true
	}
	if selectedOutboundUDP != nil && g.isProxyOutbound(selectedOutboundUDP.Type()) {
		tags[RealTag(selectedOutboundUDP)] = true
	}
	for tag := range tags {
		g.HealthCheck(tag, false)
	}
}

func (g *URLTestGroup) isProxyOutbound(outboundType string) bool {
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
	case C.TypeMieru:
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

func (g *URLTestGroup) quicOutboundInterfaceUpdated(detour adapter.Outbound, realTag string) {

	listener, isListener := detour.(adapter.InterfaceUpdateListener)
	needUpdate := detour.Type() == C.TypeHysteria || detour.Type() == C.TypeHysteria2 || detour.Type() == C.TypeTUIC
	if isListener && needUpdate {
		connections := detour.Connections()
		if connections > 0 {
			return
		}
		if outbound.OutboundGetLatestDownloadActiveConnection != nil {
			has, _ := outbound.OutboundGetLatestDownloadActiveConnection(realTag)
			if !has {
				listener.InterfaceUpdated()
			}
		}
	}
}

func (g *URLTestGroup) outboundInterfaceUpdated(detour adapter.Outbound) {
	listener, isListener := detour.(adapter.InterfaceUpdateListener)
	if isListener {
		listener.InterfaceUpdated()
	}
}
