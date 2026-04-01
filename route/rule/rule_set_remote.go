package rule

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/srs"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/ntp"
	"github.com/sagernet/sing/common/x/list"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"

	"go4.org/netipx"
)

var _ adapter.RuleSet = (*RemoteRuleSet)(nil)

type RemoteRuleSet struct {
	ctx            context.Context
	cancel         context.CancelFunc
	logger         logger.ContextLogger
	outbound       adapter.OutboundManager
	options        option.RuleSet
	updateInterval time.Duration
	dialer         N.Dialer
	access         sync.RWMutex
	rules          []adapter.HeadlessRule
	metadata       adapter.RuleSetMetadata
	lastUpdated    time.Time
	lastEtag       string
	updateTicker   *time.Ticker
	//cacheFile      adapter.CacheFile //karing
	pauseManager  pause.Manager
	callbacks     list.List[adapter.RuleSetUpdateCallback]
	refs          atomic.Int32
	downloadTimes int //karing
}

func NewRemoteRuleSet(ctx context.Context, logger logger.ContextLogger, options option.RuleSet) *RemoteRuleSet {
	ctx, cancel := context.WithCancel(ctx)
	var updateInterval time.Duration
	if options.RemoteOptions.UpdateInterval > 0 {
		updateInterval = time.Duration(options.RemoteOptions.UpdateInterval)
	} else {
		updateInterval = 24 * time.Hour
	}
	return &RemoteRuleSet{
		ctx:            ctx,
		cancel:         cancel,
		outbound:       service.FromContext[adapter.OutboundManager](ctx),
		logger:         logger,
		options:        options,
		updateInterval: updateInterval,
		pauseManager:   service.FromContext[pause.Manager](ctx),
	}
}

func (s *RemoteRuleSet) Name() string {
	return s.options.Tag
}

func (s *RemoteRuleSet) String() string {
	return strings.Join(F.MapToString(s.rules), " ")
}

func (s *RemoteRuleSet) StartContext(ctx context.Context, startContext *adapter.HTTPStartContext) error {
	cacheFile := service.FromContext[adapter.CacheFile](s.ctx) //karing
	var dialer N.Dialer
	if s.options.RemoteOptions.DownloadDetour != "" {
		outbound, loaded := s.outbound.Outbound(s.options.RemoteOptions.DownloadDetour)
		if !loaded {
			return E.New("download detour not found: ", s.options.RemoteOptions.DownloadDetour)
		}
		dialer = outbound
	} else {
		dialer = s.outbound.Default()
	}
	s.dialer = dialer
	if cacheFile != nil { //karing
		if savedSet := cacheFile.LoadRuleSet(s.options.RemoteOptions.URL); savedSet != nil { //karing
			err := s.loadBytes(savedSet.Content)
			if err != nil {
				cacheFile.DeleteRuleSet(s.options.RemoteOptions.URL) //karing
				//return E.Cause(err, "restore cached rule-set")
			} else { //karing
				s.lastUpdated = savedSet.LastUpdated
				s.lastEtag = savedSet.LastEtag
			}
		}
	}
	/*if s.lastUpdated.IsZero() {//karing
		err := s.fetch(ctx, startContext)
		if err != nil {
			return E.Cause(err, "initial rule-set: ", s.options.Tag)
		}
	}*/
	s.updateTicker = time.NewTicker(s.updateInterval)
	return nil
}

func (s *RemoteRuleSet) PostStart() error {
	go s.loopUpdate()
	return nil
}

func (s *RemoteRuleSet) Metadata() adapter.RuleSetMetadata {
	s.access.RLock()
	defer s.access.RUnlock()
	return s.metadata
}

func (s *RemoteRuleSet) ExtractIPSet() []*netipx.IPSet {
	s.access.RLock()
	defer s.access.RUnlock()
	return common.FlatMap(s.rules, extractIPSetFromRule)
}

func (s *RemoteRuleSet) IncRef() {
	s.refs.Add(1)
}

func (s *RemoteRuleSet) DecRef() {
	if s.refs.Add(-1) < 0 {
		panic("rule-set: negative refs")
	}
}

func (s *RemoteRuleSet) Cleanup() {
	if s.refs.Load() == 0 {
		s.rules = nil
	}
}

func (s *RemoteRuleSet) RegisterCallback(callback adapter.RuleSetUpdateCallback) *list.Element[adapter.RuleSetUpdateCallback] {
	s.access.Lock()
	defer s.access.Unlock()
	return s.callbacks.PushBack(callback)
}

func (s *RemoteRuleSet) UnregisterCallback(element *list.Element[adapter.RuleSetUpdateCallback]) {
	s.access.Lock()
	defer s.access.Unlock()
	s.callbacks.Remove(element)
}

func (s *RemoteRuleSet) loadBytes(content []byte) error {
	var (
		ruleSet option.PlainRuleSetCompat
		err     error
	)
	switch s.options.Format {
	case C.RuleSetFormatSource:
		ruleSet, err = json.UnmarshalExtended[option.PlainRuleSetCompat](content)
		if err != nil {
			return err
		}
	case C.RuleSetFormatBinary:
		ruleSet, err = srs.Read(bytes.NewReader(content), false)
		if err != nil {
			return err
		}
	default:
		return E.New("unknown rule-set format: ", s.options.Format)
	}
	plainRuleSet, err := ruleSet.Upgrade()
	if err != nil {
		return err
	}
	rules := make([]adapter.HeadlessRule, len(plainRuleSet.Rules))
	for i, ruleOptions := range plainRuleSet.Rules {
		rules[i], err = NewHeadlessRule(s.ctx, ruleOptions)
		if err != nil {
			return E.Cause(err, "parse rule_set.rules.[", i, "]")
		}
	}
	s.access.Lock()
	s.metadata.ContainsProcessRule = HasHeadlessRule(plainRuleSet.Rules, isProcessHeadlessRule)
	s.metadata.ContainsWIFIRule = HasHeadlessRule(plainRuleSet.Rules, isWIFIHeadlessRule)
	s.metadata.ContainsIPCIDRRule = HasHeadlessRule(plainRuleSet.Rules, isIPCIDRHeadlessRule)
	s.rules = rules
	callbacks := s.callbacks.Array()
	s.access.Unlock()
	for _, callback := range callbacks {
		callback(s)
	}
	return nil
}

func (s *RemoteRuleSet) loopUpdate() {
	if s.lastUpdated.IsZero() || time.Since(s.lastUpdated) > s.updateInterval { //karing
		err := s.fetch(s.ctx, nil)
		if err != nil {
			s.logger.Error("fetch rule-set ", s.options.Tag, ": ", err)
		} else if s.refs.Load() == 0 {
			s.rules = nil
		}
	}
	for {
		runtime.GC()
		select {
		case <-s.ctx.Done():
			return
		case <-s.updateTicker.C:
			s.pauseManager.WaitActive() //karing
			s.updateOnce()
		}
	}
}

func (s *RemoteRuleSet) updateOnce() {
	err := s.fetch(s.ctx, nil)
	if err != nil {
		s.logger.Error("fetch rule-set ", s.options.Tag, ": ", err)
	} else if s.refs.Load() == 0 {
		s.rules = nil
	}
}

func (s *RemoteRuleSet) fetch(ctx context.Context, startContext *adapter.HTTPStartContext) error {
	cacheFile := service.FromContext[adapter.CacheFile](s.ctx) //karing
	if s.downloadTimes >= 5 {                                  //karing
		s.Close()
		if cacheFile != nil {
			cacheFile.SetRulesetFetchError(s.options.Tag, "failed to download rule-set after multiple attempts")
		}
		return E.New("cancel updating rule-set ", s.options.Tag, " download_detour ", s.options.RemoteOptions.DownloadDetour, " from URL:", s.options.RemoteOptions.URL, " after ", s.downloadTimes, " times")
	}
	s.downloadTimes++                                                                                                                                                                                //karing
	s.logger.Debug("updating rule-set ", s.options.Tag, " download_detour ", s.options.RemoteOptions.DownloadDetour, " from URL: ", s.options.RemoteOptions.URL, " try ", s.downloadTimes, " times") //karing

	var httpClient *http.Client
	if startContext != nil {
		httpClient = startContext.HTTPClient(s.options.RemoteOptions.DownloadDetour, s.dialer)
	} else {
		httpClient = &http.Client{
			Transport: &http.Transport{
				ForceAttemptHTTP2:   true,
				TLSHandshakeTimeout: C.TCPTimeout,
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return s.dialer.DialContext(ctx, network, M.ParseSocksaddr(addr))
				},
				TLSClientConfig: &tls.Config{
					Time:    ntp.TimeFuncFromContext(s.ctx),
					RootCAs: adapter.RootPoolFromContext(s.ctx),
				},
			},
		}
	}
	request, err := http.NewRequest("GET", s.options.RemoteOptions.URL, nil)
	if err != nil {
		if cacheFile != nil { //karing
			cacheFile.SetRulesetFetchError(s.options.Tag, err.Error())
		}
		return err
	}
	if s.lastEtag != "" {
		request.Header.Set("If-None-Match", s.lastEtag)
	}
	response, err := httpClient.Do(request.WithContext(ctx))
	if err != nil {
		if cacheFile != nil { //karing
			cacheFile.SetRulesetFetchError(s.options.Tag, err.Error())
		}
		return err
	}
	switch response.StatusCode {
	case http.StatusOK:
	case http.StatusNotModified:
		s.lastUpdated = time.Now()
		cacheFile := service.FromContext[adapter.CacheFile](s.ctx) //karing
		if cacheFile != nil {                                      //karing
			savedRuleSet := cacheFile.LoadRuleSet(s.options.RemoteOptions.URL) //karing
			if savedRuleSet != nil {
				savedRuleSet.LastUpdated = s.lastUpdated
				err = cacheFile.SaveRuleSet(s.options.RemoteOptions.URL, savedRuleSet) //karing
				if err != nil {
					cacheFile.SetRulesetFetchError(s.options.Tag, err.Error()) //karing
					s.logger.Error("save rule-set updated time: ", err)
					return nil
				}
			}
			cacheFile.SetRulesetFetchError(s.options.Tag, "") //karing
		}

		s.logger.Info("update rule-set ", s.options.Tag, ": not modified")
		return nil
	default:
		if cacheFile != nil { //karing
			cacheFile.SetRulesetFetchError(s.options.Tag, fmt.Sprintf("unexpected status: %s", response.Status))
		}
		return E.New("unexpected status: ", response.Status)
	}
	content, err := io.ReadAll(response.Body)
	if err != nil {
		response.Body.Close()
		if cacheFile != nil { //karing
			cacheFile.SetRulesetFetchError(s.options.Tag, err.Error())
		}
		return err
	}
	err = s.loadBytes(content)
	if err != nil {
		response.Body.Close()
		if cacheFile != nil { //karing
			cacheFile.SetRulesetFetchError(s.options.Tag, err.Error())
		}
		return err
	}
	response.Body.Close()
	eTagHeader := response.Header.Get("Etag")
	if eTagHeader != "" {
		s.lastEtag = eTagHeader
	}
	s.lastUpdated = time.Now()
	if cacheFile != nil { //karing
		err = cacheFile.SaveRuleSet(s.options.RemoteOptions.URL, &adapter.SavedBinary{ //karing
			LastUpdated: s.lastUpdated,
			Content:     content,
			LastEtag:    s.lastEtag,
		})
		if err != nil {
			cacheFile.SetRulesetFetchError(s.options.Tag, err.Error()) //karing
			s.logger.Error("save rule-set cache: ", err)
		} else {
			cacheFile.SetRulesetFetchError(s.options.Tag, "") //karing
		}
	}

	s.downloadTimes = 0 //karing
	s.logger.Info("updated rule-set ", s.options.Tag)
	return nil
}

func (s *RemoteRuleSet) Close() error {
	s.rules = nil
	s.cancel()
	if s.updateTicker != nil {
		s.updateTicker.Stop()
	}
	return nil
}

func (s *RemoteRuleSet) Match(metadata *adapter.InboundContext) bool {
	return !s.matchStates(metadata).isEmpty()
}

func (s *RemoteRuleSet) matchStates(metadata *adapter.InboundContext) ruleMatchStateSet {
	return s.matchStatesWithBase(metadata, 0)
}

func (s *RemoteRuleSet) matchStatesWithBase(metadata *adapter.InboundContext, base ruleMatchState) ruleMatchStateSet {
	var stateSet ruleMatchStateSet
	for _, rule := range s.rules {
		nestedMetadata := *metadata
		nestedMetadata.ResetRuleMatchCache()
		stateSet = stateSet.merge(matchHeadlessRuleStatesWithBase(rule, &nestedMetadata, base))
	}
	return stateSet
}
