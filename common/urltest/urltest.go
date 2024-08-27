package urltest

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sagernet/sing/common"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

type URLTestResult struct { // karing
	Delay uint16 `json:"delay,omitempty"`
	Err   string `json:"err,omitempty"`
}

type History struct {
	Time  time.Time `json:"time"`
	Delay uint16    `json:"delay"`
	Err   string    `json:"err"` //karing
}

type HistoryStorage struct {
	access       sync.RWMutex
	delayHistory map[string]*History
	updateHook   chan<- struct{}
}

func NewHistoryStorage() *HistoryStorage {
	return &HistoryStorage{
		delayHistory: make(map[string]*History),
	}
}

func (s *HistoryStorage) SetHook(hook chan<- struct{}) {
	s.updateHook = hook
}

func (s *HistoryStorage) LoadURLTestHistory(tag string) *History {
	if s == nil {
		return nil
	}
	s.access.RLock()
	defer s.access.RUnlock()
	return s.delayHistory[tag]
}

func (s *HistoryStorage) DeleteURLTestHistory(tag string) {
	s.access.Lock()
	delete(s.delayHistory, tag)
	s.access.Unlock()
	s.notifyUpdated()
}

func (s *HistoryStorage) StoreURLTestHistory(tag string, history *History) {
	s.access.Lock()
	s.delayHistory[tag] = history
	s.access.Unlock()
	s.notifyUpdated()
}

func (s *HistoryStorage) GetURLTestHistory() map[string]*History { // karing
	history := make(map[string]*History)
	s.access.Lock()
	for k, v := range s.delayHistory {
		history[k] = v
	}
	s.access.Unlock()
	return history
}

func (s *HistoryStorage) notifyUpdated() {
	updateHook := s.updateHook
	if updateHook != nil {
		select {
		case updateHook <- struct{}{}:
		default:
		}
	}
}

func (s *HistoryStorage) Close() error {
	s.updateHook = nil
	return nil
}

func URLTest(ctx context.Context, link string, detour N.Dialer) (t uint16, err error) {
	if link == "" {
		link = "https://www.gstatic.com/generate_204"
	}
	linkURL, err := url.Parse(link)
	if err != nil {
		return
	}
	hostname := linkURL.Hostname()
	port := linkURL.Port()
	if port == "" {
		switch linkURL.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		}
	}

	start := time.Now()
	instance, err := detour.DialContext(ctx, "tcp", M.ParseSocksaddrHostPortStr(hostname, port))
	if err != nil {
		return
	}
	defer instance.Close()
	if earlyConn, isEarlyConn := common.Cast[N.EarlyConn](instance); isEarlyConn && earlyConn.NeedHandshake() {
		start = time.Now()
	}
	req, err := http.NewRequest(http.MethodHead, link, nil)
	if err != nil {
		return
	}
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return instance, nil
			},
			//DisableKeepAlives:   true,// karing
			//TLSHandshakeTimeout: C.TCPTimeout,// karing
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	defer client.CloseIdleConnections()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return
	}
	resp.Body.Close()
	{ //karing
		start2 := time.Now()
		resp, err = client.Do(req.WithContext(ctx))
		if err != nil {
			t = uint16(time.Since(start).Milliseconds())
			return t, nil
		}
		start = start2
		resp.Body.Close()
	}

	t = uint16(time.Since(start).Milliseconds())
	return
}