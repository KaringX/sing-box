package trafficontrol

import (
	"runtime"
	"sync"
	"time"

	"github.com/sagernet/sing-box/common/conntrack"
	"github.com/sagernet/sing-box/common/gofree"
	"github.com/sagernet/sing-box/experimental/clashapi/compatible"
	"github.com/sagernet/sing/common/atomic"
	"github.com/sagernet/sing/common/x/list"

	"github.com/gofrs/uuid/v5"
)

type Manager struct {
	startTime           time.Time //karing
	uploadTemp          atomic.Int64 //karing
	downloadTemp        atomic.Int64 //karing
	uploadBlip          atomic.Int64 //karing
	downloadBlip        atomic.Int64 //karing
	uploadTotal         atomic.Int64
	downloadTotal       atomic.Int64
	uploadTotalDirect   atomic.Int64 //karing
	downloadTotalDirect atomic.Int64 //karing

	connections             compatible.Map[uuid.UUID, Tracker]
	closedConnectionsAccess sync.Mutex
	closedConnections       list.List[TrackerMetadata]
	// process     *process.Process
	memory uint64
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Join(c Tracker) {
	m.connections.Store(c.Metadata().ID, c)
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
	m.uploadTemp.Add(size)
	m.uploadTotal.Add(size)
	if direct { //karing
		m.uploadTotalDirect.Add(size) //karing
	}
}

func (m *Manager) PushDownloaded(size int64, direct bool) { //karing
	m.downloadTemp.Add(size)
	m.downloadTotal.Add(size)
	if direct { //karing
		m.downloadTotalDirect.Add(size) //karing
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
	if includeConnections { //karing
		m.connections.Range(func(_ uuid.UUID, value Tracker) bool {
			//if value.Metadata().OutboundType != C.TypeDNS {//karing
			connections = append(connections, value)
			//}	
			return true
		})
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.memory = memStats.StackInuse + memStats.HeapInuse + memStats.HeapIdle - memStats.HeapReleased

	return &Snapshot{//karing
		StartTime:           m.startTime,  
		UploadTotal:         m.uploadTotal.Load(),
		DownloadTotal:       m.downloadTotal.Load(),
		UploadTotalDirect:   m.uploadTotalDirect.Load(),    
		DownloadSpeed:       m.downloadBlip.Load(),         
		UploadSpeed:         m.uploadBlip.Load(),           
		DownloadTotalDirect: m.downloadTotalDirect.Load(),  
		ConnectionsOut:      int32(conntrack.Count()),      
		ConnectionsIn:       int32(m.connections.Len()),    
		Goroutines:          int32(runtime.NumGoroutine()), 
		Connections:         connections,
		ThreadCount:         int32(gofree.ThreadNum()),
		Memory:              m.memory,
	}
}
func (m *Manager) OutboundHasConnections(tag string) bool {  //karing
	hasConn := false;
	m.connections.Range(func(_ uuid.UUID, value Tracker) bool {
		if info, istrack := value.(*TCPConn); istrack {
			for _, data := range info.metadata.Chain{
				if(data == tag){
					hasConn = true
					return false;
				}
			}
			return true;
		}
		if info, istrack := value.(*UDPConn); istrack {
			for _, data := range info.metadata.Chain{
				if(data == tag){
					hasConn = true
					return false;
				}
			}
			return true;
		}

		return true
	})
	return hasConn
}
func (m *Manager) ResetStatistic() {
	m.uploadTotal.Store(0)
	m.downloadTotal.Store(0)
	m.uploadTotalDirect.Store(0)   //karing
	m.downloadTotalDirect.Store(0) //karing
}
func (m *Manager) Close() error { //karing
	m.startTime = time.Now()
	m.ResetStatistic()
	m.connections.Clear()
	return nil
}

/* //karing
type Snapshot struct {
	Download    int64
	Upload      int64
	Connections []Tracker
	Memory      uint64
}

func (s *Snapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"downloadTotal": s.Download,
		"uploadTotal":   s.Upload,
		"connections":   common.Map(s.Connections, func(t Tracker) TrackerMetadata { return t.Metadata() }),
		"memory":        s.Memory,
	})
}
*/
type Snapshot struct {
	StartTime           time.Time `json:"startTime"` //karing
	DownloadTotal       int64     `json:"downloadTotal"`
	UploadTotal         int64     `json:"uploadTotal"`
	DownloadTotalDirect int64     `json:"downloadTotalDirect"` //karing
	UploadTotalDirect   int64     `json:"uploadTotalDirect"`   //karing
	DownloadSpeed       int64     `json:"downloadSpeed"`       //karing
	UploadSpeed         int64     `json:"uploadSpeed"`         //karing
	ConnectionsOut      int32     `json:"connectionsOut"`      //karing
	ConnectionsIn       int32     `json:"connectionsIn"`       //karing
	Goroutines          int32     `json:"goroutines"`          //karing
	ConnectionsCount    int32     `json:"connectionsCount"`    //karing
	Connections         []Tracker `json:"connections"`
	ThreadCount         int32     `json:"threadCount"`         //karing
	Memory              uint64    `json:"memory"`
}





