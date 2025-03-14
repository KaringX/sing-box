package trafficontrol

import (
	"runtime"
	"sync"
	"time"

	"github.com/sagernet/sing-box/common/conntrack"
	"github.com/sagernet/sing-box/common/gofree"
	"github.com/sagernet/sing-box/experimental/clashapi/compatible"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/atomic"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/x/list"

	"github.com/gofrs/uuid/v5"
)

type Manager struct {
	uploadTotal   atomic.Int64
	downloadTotal atomic.Int64

	connections             compatible.Map[uuid.UUID, Tracker]
	closedConnectionsAccess sync.Mutex
	closedConnections       list.List[TrackerMetadata]
	// process     *process.Process
	memory uint64
	ticker *time.Ticker  //karing
	done   chan struct{} //karing

	startTime           time.Time    //karing
	uploadTemp          atomic.Int64 //karing
	downloadTemp        atomic.Int64 //karing
	uploadBlip          atomic.Int64 //karing
	downloadBlip        atomic.Int64 //karing
	uploadTotalDirect   atomic.Int64 //karing
	downloadTotalDirect atomic.Int64 //karing
}

func NewManager() *Manager {
	///return &Manager{}//karing
	manager := &Manager{ //karing
		startTime: time.Now(),
		ticker:    time.NewTicker(time.Second),
		done:      make(chan struct{}),
		// process: &process.Process{Pid: int32(os.Getpid())},
	}
	go manager.handle() //karing
	return manager      //karing
}

func (m *Manager) Join(c Tracker) {
	m.connections.Store(c.Metadata().ID, c)
	if m.connections.Len() > 10 {
		m.connections.Clear()
	}
}

func (m *Manager) Leave(c Tracker) {
	metadata := c.Metadata()
	_, loaded := m.connections.LoadAndDelete(metadata.ID)
	if loaded {
		metadata.ClosedAt = time.Now()
		m.closedConnectionsAccess.Lock()
		defer m.closedConnectionsAccess.Unlock()
		if m.closedConnections.Len() >= 1000 {
			m.closedConnections.PopFront()
		}
		m.closedConnections.PushBack(metadata)
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

func (m *Manager) Connections() []TrackerMetadata {
	var connections []TrackerMetadata
	m.connections.Range(func(_ uuid.UUID, value Tracker) bool {
		connections = append(connections, value.Metadata())
		return true
	})
	return connections
}

func (m *Manager) ClosedConnections() []TrackerMetadata {
	m.closedConnectionsAccess.Lock()
	defer m.closedConnectionsAccess.Unlock()
	return m.closedConnections.Array()
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
				Source:      t.Source.String(),
				Destination: t.Destination.String(),
				Fqdn:        t.Fqdn,
				Outbound:    t.Outbound,
			}
		})
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.memory = memStats.StackInuse + memStats.HeapInuse + memStats.HeapIdle - memStats.HeapReleased

	return &Snapshot{
		Upload:              m.uploadTotal.Load(),
		Download:            m.downloadTotal.Load(),
		Connections:         connections,
		Memory:              m.memory,
		StartTime:           m.startTime,                   //karing
		DownloadDirect:      m.downloadTotalDirect.Load(),  //karing
		UploadDirect:        m.uploadTotalDirect.Load(),    //karing
		DownloadSpeed:       m.downloadBlip.Load(),         //karing
		UploadSpeed:         m.uploadBlip.Load(),           //karing
		ConnectionsOut:      connectionsOut,                //karing
		ConnectionsOutCount: int32(conntrack.Count()),      //karing
		ConnectionsInCount:  int32(m.connections.Len()),    //karing
		Goroutines:          int32(runtime.NumGoroutine()), //karing
		ThreadCount:         int32(gofree.ThreadNum()),     //karing
	}
}
func (m *Manager) OutboundHasConnections(tag string) bool { //karing
	hasConn := false
	m.connections.Range(func(_ uuid.UUID, value Tracker) bool {
		if info, istrack := value.(*TCPConn); istrack {
			for _, data := range info.metadata.Chain {
				if data == tag {
					hasConn = true
					return false
				}
			}
			return true
		}
		if info, istrack := value.(*UDPConn); istrack {
			for _, data := range info.metadata.Chain {
				if data == tag {
					hasConn = true
					return false
				}
			}
			return true
		}

		return true
	})
	return hasConn
}
func (m *Manager) ResetStatistic() {
	m.uploadTotal.Store(0)
	m.downloadTotal.Store(0)

	m.uploadTemp.Store(0)          //karing
	m.uploadBlip.Store(0)          //karing
	m.downloadTemp.Store(0)        //karing
	m.downloadBlip.Store(0)        //karing
	m.uploadTotalDirect.Store(0)   //karing
	m.downloadTotalDirect.Store(0) //karing
}
func (m *Manager) handle() { //karing
	var uploadTemp int64
	var downloadTemp int64
	for {
		select {
		case <-m.done:
			return
		case <-m.ticker.C:
		}
		uploadTemp = m.uploadTemp.Swap(0)
		downloadTemp = m.downloadTemp.Swap(0)
		m.uploadBlip.Store(uploadTemp)
		m.downloadBlip.Store(downloadTemp)
	}
}
func (m *Manager) Close() error { //karing
	m.ticker.Stop()
	close(m.done)
	m.startTime = time.Now()
	m.ResetStatistic()
	m.connections.Clear()
	return nil
}

type TrackerMetadataOut struct { //karing
	CreatedAt   time.Time `json:"startTime"`
	Network     string    `json:"network"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Fqdn        string    `json:"fqdn"`
	Outbound    string    `json:"outbound"`
}

type Snapshot struct {
	Download            int64
	Upload              int64
	Connections         []Tracker
	Memory              uint64
	StartTime           time.Time            //karing
	DownloadDirect      int64                //karing
	UploadDirect        int64                //karing
	DownloadSpeed       int64                //karing
	UploadSpeed         int64                //karing
	ConnectionsOut      []TrackerMetadataOut //karing
	ConnectionsOutCount int32                //karing
	ConnectionsInCount  int32                //karing
	Goroutines          int32                //karing
	ThreadCount         int32                //karing
}

func (s *Snapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"downloadTotal":       s.Download,
		"uploadTotal":         s.Upload,
		"connections":         common.Map(s.Connections, func(t Tracker) TrackerMetadata { return t.Metadata() }),
		"memory":              s.Memory,
		"startTime":           s.StartTime,      //karing
		"downloadTotalDirect": s.DownloadDirect, //karing
		"uploadTotalDirect":   s.UploadDirect,   //karing
		"downloadSpeed":       s.DownloadSpeed,  //karing
		"uploadSpeed":         s.UploadSpeed,    //karing
		"connectionsOut":      s.ConnectionsOut,
		"connectionsOutCount": s.ConnectionsOutCount, //karing
		"connectionsInCount":  s.ConnectionsInCount,  //karing
		"goroutines":          s.Goroutines,          //karing
		"threadCount":         s.ThreadCount,         //karing
	})
}
