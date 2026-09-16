package trafficcontrol

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	tun "github.com/sagernet/sing-tun"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/bufio"
	N "github.com/sagernet/sing/common/network"

	"github.com/gofrs/uuid/v5"
)

type TrackerMetadata struct {
	ID            uuid.UUID
	Metadata      adapter.InboundContext
	CreatedAt     time.Time
	ClosedAt      time.Time
	Upload        *atomic.Int64
	Download      *atomic.Int64
	UploadBlip    *atomic.Int64 //karing
	DownloadBlip  *atomic.Int64 //karing
	UploadSpeed   int64         //karing
	DownloadSpeed int64         //karing
	Chain         []string
	Rule          adapter.Rule
	Outbound      string
	OutboundType  string
	User          string       //karing
	Protocol      string       //karing
	UploadLast    *time.Time   //karing
	DownloadLast  *time.Time   //karing
	Dirty         *atomic.Bool //karing
}

type Tracker interface {
	Metadata() *TrackerMetadata
	Close() error
}

func (m *Manager) RoutedConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext, matchedRule adapter.Rule, matchOutbound adapter.Outbound) net.Conn {
	upload := new(atomic.Int64)
	download := new(atomic.Int64)
	uploadBlip := new(atomic.Int64)                                          //karing
	downloadBlip := new(atomic.Int64)                                        //karing
	uploadLast := new(time.Time)                                             //karing
	downloadLast := new(time.Time)                                           //karing
	dirty := new(atomic.Bool)                                                //karing
	_, _, outboundType := GetMatchRuleChain(m.outbound, matchOutbound.Tag()) //karing
	tracker := &connTracker{
		ExtendedConn: bufio.NewCounterConn(conn, []N.CountFunc{func(n int64) { //karing
			upload.Add(n)
			uploadBlip.Add(n)                               //karing
			dirty.Store(true)                               //karing
			*uploadLast = time.Now()                        //karing
			m.PushUploaded(n, outboundType == C.TypeDirect) //karing
		}}, []N.CountFunc{func(n int64) {
			download.Add(n)
			downloadBlip.Add(n)                               //karing
			dirty.Store(true)                                 //karing
			*downloadLast = time.Now()                        //karing
			m.PushDownloaded(n, outboundType == C.TypeDirect) //karing
		}}),
		metadata: m.newTrackerMetadata(metadata, matchedRule, matchOutbound, upload, download, uploadBlip, downloadBlip, uploadLast, downloadLast, dirty), //karing
		manager:  m,
	}
	m.join(tracker)
	return tracker
}

func (m *Manager) RoutedPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext, matchedRule adapter.Rule, matchOutbound adapter.Outbound) N.PacketConn {
	upload := new(atomic.Int64)
	download := new(atomic.Int64)
	uploadBlip := new(atomic.Int64)                                          //karing
	downloadBlip := new(atomic.Int64)                                        //karing
	uploadLast := new(time.Time)                                             //karing
	downloadLast := new(time.Time)                                           //karing
	dirty := new(atomic.Bool)                                                //karing
	_, _, outboundType := GetMatchRuleChain(m.outbound, matchOutbound.Tag()) //karing
	tracker := &packetConnTracker{
		PacketConn: bufio.NewCounterPacketConn(conn, []N.CountFunc{func(n int64) {
			upload.Add(n)
			dirty.Store(true)                               //karing
			uploadBlip.Add(n)                               //karing
			*uploadLast = time.Now()                        //karing
			m.PushUploaded(n, outboundType == C.TypeDirect) //karing
		}}, []N.CountFunc{func(n int64) {
			download.Add(n)
			downloadBlip.Add(n)                               //karing
			dirty.Store(true)                                 //karing
			*downloadLast = time.Now()                        //karing
			m.PushDownloaded(n, outboundType == C.TypeDirect) //karing
		}}),
		metadata: m.newTrackerMetadata(metadata, matchedRule, matchOutbound, upload, download, uploadBlip, downloadBlip, uploadLast, downloadLast, dirty), //karing
		manager:  m,
	}
	m.join(tracker)
	return tracker
}

func (m *Manager) RoutedFlow(ctx context.Context, metadata adapter.InboundContext, matchedRule adapter.Rule, matchOutbound adapter.Outbound) tun.FlowTracker {
	return &flowTracker{
		metadata: m.newTrackerMetadata(metadata, matchedRule, matchOutbound, new(atomic.Int64), new(atomic.Int64), new(atomic.Int64), new(atomic.Int64), new(time.Time), new(time.Time), new(atomic.Bool)), //karing
		manager:  m,
	}
}

func (m *Manager) newTrackerMetadata(metadata adapter.InboundContext, matchedRule adapter.Rule, matchOutbound adapter.Outbound, upload *atomic.Int64, download *atomic.Int64, uploadBlip *atomic.Int64, downloadBlip *atomic.Int64, uploadLast *time.Time, downloadLast *time.Time, dirty *atomic.Bool) TrackerMetadata { //karing
	id, _ := uuid.NewV4()
	chain, outbound, outboundType := GetMatchRuleChain(m.outbound, matchOutbound.Tag()) //karing
	/* //karing
	var (
		chain        []string
		next         string
		outbound     string
		outboundType string
	)
	if matchOutbound != nil {
		next = matchOutbound.Tag()
	} else {
		next = m.outbound.Default().Tag()
	}
	for {
		detour, loaded := m.outbound.Outbound(next)
		if !loaded {
			break
		}
		chain = append(chain, next)
		outbound = detour.Tag()
		outboundType = detour.Type()
		outboundGroup, isGroup := detour.(adapter.OutboundGroup)
		if !isGroup {
			break
		}
		next = outboundGroup.Now()
	}
	*/
	return TrackerMetadata{
		ID:            id,
		Metadata:      metadata,
		CreatedAt:     time.Now(),
		Upload:        upload,
		Download:      download,
		UploadBlip:    uploadBlip,   //karing
		DownloadBlip:  downloadBlip, //karing
		UploadSpeed:   0,            //karing
		DownloadSpeed: 0,            //karing
		Chain:         common.Reverse(chain),
		Rule:          matchedRule,
		Outbound:      outbound,
		OutboundType:  outboundType,
		User:          metadata.User,     //karing
		Protocol:      metadata.Protocol, //karing
		UploadLast:    uploadLast,        //karing
		DownloadLast:  downloadLast,      //karing
		Dirty:         dirty,             //karing
	}
}

type connTracker struct {
	N.ExtendedConn
	metadata TrackerMetadata
	manager  *Manager
}

func (t *connTracker) Metadata() *TrackerMetadata {
	return &t.metadata
}

func (t *connTracker) Close() error {
	t.metadata.ClosedAt = time.Now() //karing
	t.metadata.Dirty.Store(true)     //karing
	t.manager.leave(t)
	return t.ExtendedConn.Close()
}

func (t *connTracker) Upstream() any {
	return t.ExtendedConn
}

func (t *connTracker) ReaderReplaceable() bool {
	return true
}

func (t *connTracker) WriterReplaceable() bool {
	return true
}

var (
	_ Tracker         = (*flowTracker)(nil)
	_ tun.FlowTracker = (*flowTracker)(nil)
)

type flowTracker struct {
	metadata TrackerMetadata
	manager  *Manager
	handle   tun.FlowHandle
}

func (t *flowTracker) Metadata() *TrackerMetadata {
	return &t.metadata
}

func (t *flowTracker) AttachFlow(handle tun.FlowHandle) {
	t.handle = handle
	t.manager.join(t)
}

func (t *flowTracker) CountForward(n int) {
	t.metadata.Upload.Add(int64(n))
}

func (t *flowTracker) CountReverse(n int) {
	t.metadata.Download.Add(int64(n))
}

func (t *flowTracker) FlowEstablished() {
}

func (t *flowTracker) CloseFlow(reason tun.FlowCloseReason) {
	t.manager.leave(t)
}

func (t *flowTracker) Close() error {
	t.metadata.ClosedAt = time.Now() //karing
	t.metadata.Dirty.Store(true)     //karing
	handle := t.handle
	if handle != nil {
		handle.CloseFlow()
	} else {
		t.manager.leave(t)
	}
	return nil
}

type packetConnTracker struct {
	N.PacketConn
	metadata TrackerMetadata
	manager  *Manager
}

func (t *packetConnTracker) Metadata() *TrackerMetadata {
	return &t.metadata
}

func (t *packetConnTracker) Close() error {
	t.metadata.ClosedAt = time.Now() //karing
	t.metadata.Dirty.Store(true)     //karing
	t.manager.leave(t)
	return t.PacketConn.Close()
}

func (t *packetConnTracker) Upstream() any {
	return t.PacketConn
}

func (t *packetConnTracker) ReaderReplaceable() bool {
	return true
}

func (t *packetConnTracker) WriterReplaceable() bool {
	return true
}
