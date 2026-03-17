package trafficontrol

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/compatible"
	"github.com/sagernet/sing-box/common/conntrack"
	safe "github.com/sagernet/sing-box/common/fix"
	"github.com/sagernet/sing-box/common/gofree"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/observable"
	"github.com/sagernet/sing/common/x/list"
	"github.com/sagernet/sing/service"
)

type ConnectionEventType int

const (
	ConnectionEventNew ConnectionEventType = iota
	ConnectionEventUpdate
	ConnectionEventClosed
)

type ConnectionEvent struct {
	Type          ConnectionEventType
	ID            uuid.UUID
	Metadata      *TrackerMetadata
	UplinkDelta   int64
	DownlinkDelta int64
	ClosedAt      time.Time
}

const closedConnectionsLimit = 1000

type Manager struct {
	ManagerExtension //karing

	uploadTotal   atomic.Int64
	downloadTotal atomic.Int64

	connections             compatible.Map[uuid.UUID, Tracker]
	closedConnectionsAccess sync.Mutex
	closedConnections       list.List[TrackerMetadata]
	memory                  uint64

	eventSubscriber *observable.Subscriber[ConnectionEvent]
}

func NewManager(ctx context.Context, logFactory log.ObservableFactory) *Manager { //karing
	///return &Manager{}//karing
	return newManagerWithExtension(ctx, logFactory) //karing
}

func (m *Manager) SetEventHook(subscriber *observable.Subscriber[ConnectionEvent]) {
	m.eventSubscriber = subscriber
}

func (m *Manager) Join(c Tracker) {
	metadata := c.Metadata()
	m.connections.Store(metadata.ID, c)
	if m.eventSubscriber != nil {
		m.eventSubscriber.Emit(ConnectionEvent{
			Type:     ConnectionEventNew,
			ID:       metadata.ID,
			Metadata: metadata,
		})
	}
}

func (m *Manager) Leave(c Tracker) {
	metadata := c.Metadata()
	_, loaded := m.connections.LoadAndDelete(metadata.ID)
	if loaded {
		closedAt := time.Now()
		metadata.ClosedAt = closedAt
		metadataCopy := *metadata
		m.closedConnectionsAccess.Lock()
		if m.closedConnections.Len() >= closedConnectionsLimit {
			m.closedConnections.PopFront()
		}
		m.closedConnections.PushBack(metadataCopy)
		statistics := service.FromContext[adapter.Statistics](m.ctx) //karing
		if statistics != nil {                                       //karing
			m.persistAccess.Lock()
			defer m.persistAccess.Unlock()
			m.closedConnectionsForPersist.PushBack(*metadata)
		}

		m.closedConnectionsAccess.Unlock()
		if m.eventSubscriber != nil {
			m.eventSubscriber.Emit(ConnectionEvent{
				Type:     ConnectionEventClosed,
				ID:       metadata.ID,
				Metadata: &metadataCopy,
				ClosedAt: closedAt,
			})
		}
	}
}

func (m *Manager) PushUploaded(size int64, direct bool) { //karing
	m.uploadTemp.Add(size) //karing
	m.uploadTotal.Add(size)
	if direct { //karing
		m.uploadTotalDirect.Add(size)
	}
}

func (m *Manager) PushDownloaded(size int64, direct bool) { //karing
	m.downloadTemp.Add(size) //karing
	m.downloadTotal.Add(size)
	if direct { //karing
		m.downloadTotalDirect.Add(size)
	}
}

func (m *Manager) Total() (up int64, down int64) {
	return m.uploadTotal.Load(), m.downloadTotal.Load()
}

func (m *Manager) ConnectionsLen() int {
	return m.connections.Len()
}

func (m *Manager) Connections() []*TrackerMetadata {
	var connections []*TrackerMetadata
	m.connections.Range(func(_ uuid.UUID, value Tracker) bool {
		connections = append(connections, value.Metadata())
		return true
	})
	return connections
}

func (m *Manager) ClosedConnections() []*TrackerMetadata {
	m.closedConnectionsAccess.Lock()
	values := m.closedConnections.Array()
	m.closedConnectionsAccess.Unlock()
	if len(values) == 0 {
		return nil
	}
	connections := make([]*TrackerMetadata, len(values))
	for i := range values {
		connections[i] = &values[i]
	}
	return connections
}

func (m *Manager) Connection(id uuid.UUID) Tracker {
	connection, loaded := m.connections.Load(id)
	if !loaded {
		return nil
	}
	return connection
}

func (m *Manager) Snapshot(includeConnections bool) *Snapshot { //karing
	var connections []Tracker
	var connectionsOut []TrackerMetadataOut //karing
	if includeConnections {                 //karing
		m.connections.Range(func(_ uuid.UUID, value Tracker) bool {
			//if value.Metadata().OutboundType != C.TypeDNS {//karing
			connections = append(connections, value)
			//}
			return true
		})
		connectionsOut = common.Map(conntrack.Connections(), func(t conntrack.OutboundConn) TrackerMetadataOut { //karing
			return TrackerMetadataOut{
				CreatedAt:   t.CreatedAt,
				Network:     t.Network,
				Source:      safe.SafeSocksaddrString(t.Source),      //karing
				Destination: safe.SafeSocksaddrString(t.Destination), //karing
				Fqdn:        t.Fqdn,
				Outbound:    t.Outbound,
			}
		})
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.memory = memStats.StackInuse + memStats.HeapInuse + memStats.HeapIdle - memStats.HeapReleased

	return &Snapshot{
		Upload:      m.uploadTotal.Load(),
		Download:    m.downloadTotal.Load(),
		Connections: connections,
		Memory:      m.memory,
		SnapshotExtension: SnapshotExtension{ //karing
			StartTime:           m.startTime,
			DownloadDirect:      m.downloadTotalDirect.Load(),
			UploadDirect:        m.uploadTotalDirect.Load(),
			DownloadSpeed:       m.downloadBlip.Load(),
			UploadSpeed:         m.uploadBlip.Load(),
			ConnectionsOut:      connectionsOut,
			ConnectionsOutCount: int32(conntrack.Count()),
			ConnectionsInCount:  int32(m.connections.Len()),
			Goroutines:          int32(runtime.NumGoroutine()),
			ThreadCount:         int32(gofree.ThreadNum()),
		},
	}
}

func (m *Manager) ResetStatistic() {
	m.uploadTotal.Store(0)
	m.downloadTotal.Store(0)

	m.resetStatistic() //karing
}

type Snapshot struct {
	SnapshotExtension //karing
	Download          int64
	Upload            int64
	Connections       []Tracker
	Memory            uint64
}

func (s *Snapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"downloadTotal":       s.Download,
		"uploadTotal":         s.Upload,
		"connections":         common.Map(s.Connections, func(t Tracker) TrackerMetadata { return *t.Metadata() }),
		"memory":              s.Memory,
		"startTime":           s.StartTime,           //karing
		"downloadTotalDirect": s.DownloadDirect,      //karing
		"uploadTotalDirect":   s.UploadDirect,        //karing
		"downloadSpeed":       s.DownloadSpeed,       //karing
		"uploadSpeed":         s.UploadSpeed,         //karing
		"connectionsOut":      s.ConnectionsOut,      //karing
		"connectionsOutCount": s.ConnectionsOutCount, //karing
		"connectionsInCount":  s.ConnectionsInCount,  //karing
		"goroutines":          s.Goroutines,          //karing
		"threadCount":         s.ThreadCount,         //karing
	})
}
