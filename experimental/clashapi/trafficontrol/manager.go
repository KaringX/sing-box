package trafficontrol

import (
	"runtime"
	"time"

	"github.com/sagernet/sing-box/common/conntrack"
	"github.com/sagernet/sing-box/common/gofree"
	"github.com/sagernet/sing-box/experimental/clashapi/compatible"
	"github.com/sagernet/sing/common/atomic"
)

type Manager struct {
	StartTime           time.Time //karing
	uploadTemp          atomic.Int64
	downloadTemp        atomic.Int64
	uploadBlip          atomic.Int64
	downloadBlip        atomic.Int64
	uploadTotal         atomic.Int64
	downloadTotal       atomic.Int64
	uploadTotalDirect   atomic.Int64 //karing
	downloadTotalDirect atomic.Int64 //karing

	connections compatible.Map[string, tracker]
	ticker      *time.Ticker
	done        chan struct{}
	// process     *process.Process
	memory uint64
}

func NewManager() *Manager {
	manager := &Manager{
		StartTime: time.Now(), //karing
		ticker:    time.NewTicker(time.Second),
		done:      make(chan struct{}),
		// process: &process.Process{Pid: int32(os.Getpid())},
	}
	go manager.handle()
	return manager
}

func (m *Manager) Join(c tracker) {
	m.connections.Store(c.ID(), c)
}

func (m *Manager) Leave(c tracker) {
	m.connections.Delete(c.ID())
}

func (m *Manager) PushUploaded(size int64, protocol string, outbound string) { //karing
	m.uploadTemp.Add(size)
	m.uploadTotal.Add(size)
	if protocol == "direct" { //karing
		m.uploadTotalDirect.Add(size) //karing
	}
}

func (m *Manager) PushDownloaded(size int64, protocol string, outbound string) { //karing
	m.downloadTemp.Add(size)
	m.downloadTotal.Add(size)
	if protocol == "direct" { //karing
		m.downloadTotalDirect.Add(size) //karing
	}
}

func (m *Manager) Now() (up int64, down int64) {
	return m.uploadBlip.Load(), m.downloadBlip.Load()
}

func (m *Manager) Total() (up int64, down int64) {
	return m.uploadTotal.Load(), m.downloadTotal.Load()
}

func (m *Manager) Connections() int {
	return m.connections.Len()
}

func (m *Manager) OutboundHasConnections(tag string) bool {
	hasConn := false;
	m.connections.Range(func(_ string, value tracker) bool {
		if info, istrack := value.(*tcpTracker); istrack {
			for _, data := range info.Chain{
				if(data == tag){
					hasConn = true
					return false;
				}
			}
			return true;
		}
		if info, istrack := value.(*udpTracker); istrack {
			for _, data := range info.Chain{
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
func (m *Manager) Snapshot(includeConnections bool) *Snapshot { //karing
	var connections []tracker
	if includeConnections { //karing
		m.connections.Range(func(_ string, value tracker) bool {
			connections = append(connections, value)
			return true
		})
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.memory = memStats.StackInuse + memStats.HeapInuse + memStats.HeapIdle - memStats.HeapReleased

	return &Snapshot{
		StartTime:           m.StartTime, //karing
		UploadTotal:         m.uploadTotal.Load(),
		DownloadTotal:       m.downloadTotal.Load(),
		UploadTotalDirect:   m.uploadTotalDirect.Load(),   //karing
		DownloadSpeed:       m.downloadBlip.Load(),        //karing
		UploadSpeed:         m.uploadBlip.Load(),          //karing
		DownloadTotalDirect: m.downloadTotalDirect.Load(), //karing
		ConnectionsOut:      int32(conntrack.Count()),     //karing
		ConnectionsIn:       int32(m.connections.Len()),   //karing
		Goroutines:          int32(runtime.NumGoroutine()),//karing
		Connections:         connections,
		ThreadCount:         int32(gofree.ThreadNum()),    //karing
		Memory:              m.memory,
	}
}

func (m *Manager) ResetStatistic() {
	m.uploadTemp.Store(0)
	m.uploadBlip.Store(0)
	m.uploadTotal.Store(0)
	m.downloadTemp.Store(0)
	m.downloadBlip.Store(0)
	m.downloadTotal.Store(0)
	m.uploadTotalDirect.Store(0)   //karing
	m.downloadTotalDirect.Store(0) //karing
}

func (m *Manager) handle() {
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

func (m *Manager) Close() error {
	m.ticker.Stop()
	close(m.done)
	m.ResetStatistic()  //karing
	m.connections.Clear()  //karing
	return nil
}

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
	Connections         []tracker `json:"connections"`
	ThreadCount         int32     `json:"threadCount"`         //karing
	Memory              uint64    `json:"memory"`
}