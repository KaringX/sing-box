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
	"github.com/sagernet/sing-box/common/gofree"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing/common"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/x/list"
	"github.com/sagernet/sing/service"
)

var coreStartTime = time.Now() //karing
type Manager struct {
	ctx           context.Context
	logger        log.ContextLogger //karing
	uploadTotal   atomic.Int64
	downloadTotal atomic.Int64

	connections             compatible.Map[uuid.UUID, Tracker]
	closedConnectionsAccess sync.Mutex
	closedConnections       list.List[TrackerMetadata]
	// process     *process.Process
	memory   uint64
	ticker   *time.Ticker  //karing
	dbTicker *time.Ticker  //karing
	done     chan struct{} //karing

	startTime           time.Time    //karing
	uploadTemp          atomic.Int64 //karing
	downloadTemp        atomic.Int64 //karing
	uploadBlip          atomic.Int64 //karing
	downloadBlip        atomic.Int64 //karing
	uploadTotalDirect   atomic.Int64 //karing
	downloadTotalDirect atomic.Int64 //karing
}

func NewManager(ctx context.Context, logFactory log.ObservableFactory) *Manager { //karing
	///return &Manager{}//karing
	manager := &Manager{ //karing
		ctx:       ctx,
		logger:    logFactory.NewLogger("traficontrol"),
		startTime: time.Now(),
		ticker:    time.NewTicker(time.Second),
		dbTicker:  time.NewTicker(time.Second),
		done:      make(chan struct{}),
		// process: &process.Process{Pid: int32(os.Getpid())},
	}
	dbFile := service.FromContext[adapter.DBFile](ctx) //karing
	if dbFile != nil {                                 //karing
		err := dbFile.Exec(createTableSQL())
		if err != nil {
			manager.logger.WarnContext(manager.ctx, "create table connection_track: ", err)
		} else {
			go func() {
				dbFile.Exec(deleteOldSQL())
			}()
			go manager.handleDB()
		}
	}
	go manager.handle() //karing

	return manager //karing
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

func (m *Manager) OutboundGetLatestDownloadActiveConnection(tag string) (bool, time.Time) { //karing
	hasConn := false
	var downloadLatest time.Time
	m.connections.Range(func(_ uuid.UUID, value Tracker) bool {
		if info, istrack := value.(*TCPConn); istrack {
			for _, data := range info.metadata.Chain {
				if data == tag {
					hasConn = true
					if downloadLatest.Before(*info.metadata.DownloadLast) {
						downloadLatest = *info.metadata.DownloadLast
					}
				}
			}
			return true
		}
		if info, istrack := value.(*UDPConn); istrack {
			for _, data := range info.metadata.Chain {
				if data == tag {
					hasConn = true
					if downloadLatest.Before(*info.metadata.DownloadLast) {
						downloadLatest = *info.metadata.DownloadLast
					}
				}
			}
			return true
		}

		return true
	})
	return hasConn, downloadLatest
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

func (m *Manager) handleDB() { //karing
	for {
		select {
		case <-m.done:
			return
		case <-m.dbTicker.C:
			m.persistToDB()
		}
	}
}

func (m *Manager) persistToDB() { //karing
	dbFile := service.FromContext[adapter.DBFile](m.ctx)
	if dbFile != nil {
		tx, err := dbFile.BeginTx()
		if err != nil {
			m.logger.WarnContext(m.ctx, "db begin transaction: ", err)
			return
		}
		stmt, err := tx.Prepare(prepareSQL())
		if err != nil {
			m.logger.WarnContext(m.ctx, "db transaction prepare: ", err)
			return
		}
		defer stmt.Close()

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		m.memory = memStats.StackInuse + memStats.HeapInuse + memStats.HeapIdle - memStats.HeapReleased

		m.connections.Range(func(_ uuid.UUID, value Tracker) bool {
			t := value.Metadata()
			if !t.Dirty.Load() {
				return true
			}
			t.Dirty.Store(false)
			var inbound string
			if t.Metadata.Inbound != "" {
				inbound = t.Metadata.InboundType + "/" + t.Metadata.Inbound
			} else {
				inbound = t.Metadata.InboundType
			}
			var domain string
			if t.Metadata.Domain != "" {
				domain = t.Metadata.Domain
			} else {
				domain = t.Metadata.Destination.Fqdn
			}
			var processPath string
			var packageName string
			if t.Metadata.ProcessInfo != nil {
				if t.Metadata.ProcessInfo.ProcessPath != "" {
					processPath = t.Metadata.ProcessInfo.ProcessPath
				} else if t.Metadata.ProcessInfo.PackageName != "" {
					packageName = t.Metadata.ProcessInfo.PackageName
				}
				if processPath == "" {
					if t.Metadata.ProcessInfo.UserId != -1 {
						processPath = F.ToString(t.Metadata.ProcessInfo.UserId)
					}
				} else if t.Metadata.ProcessInfo.User != "" {
					processPath = F.ToString(processPath, " (", t.Metadata.ProcessInfo.User, ")")
				} else if t.Metadata.ProcessInfo.UserId != -1 {
					processPath = F.ToString(processPath, " (", t.Metadata.ProcessInfo.UserId, ")")
				}
			}
			var rule0 string
			var rule1 string
			if t.Rule != nil {
				rule0 = t.Rule.Name()
				rule1 = t.Rule.Action().Target()
			} else {
				rule0 = "final"
			}
			var chain0 string
			var chain1 string
			if len(t.Chain) == 1 {
				chain0 = t.Chain[0]
			} else if len(t.Chain) >= 2 {
				chain1 = t.Chain[0]
				chain0 = t.Chain[len(t.Chain)-1]
			}
			var source_ip string
			var destination_ip string
			if t.Metadata.Source.Addr.IsValid() {
				source_ip = t.Metadata.Source.Addr.String()
			}
			if t.Metadata.Destination.Addr.IsValid() {
				destination_ip = t.Metadata.Destination.Addr.String()
			}

			/*
					core_start ,
				    total_upload ,
					total_download ,
					total_upload_speed ,
					total_download_speed ,
					total_upload_direct ,
					total_download_direct ,
					connections_in ,
					connections_out ,
					goroutines ,
					thread ,
					memory ,
				    connection_id ,
					create_at ,
				    inbound ,
					network ,
					protocol ,
					process ,
					package ,
					source_ip ,
					source_port ,
					host ,
					destination_ip ,
					destination_port ,
					upload ,
					download ,
					upload_speed ,
					download_speed ,
					rule0 ,
					rule1 ,
					chain0 ,
					chain1 ,
					outbound_type */
			_, err = stmt.Exec(
				coreStartTime,
				m.uploadTotal.Load(),
				m.downloadTotal.Load(),
				m.uploadBlip.Load(),
				m.downloadBlip.Load(),
				m.uploadTotalDirect.Load(),
				m.downloadTotalDirect.Load(),
				int32(m.connections.Len()),
				int32(conntrack.Count()),
				int32(runtime.NumGoroutine()),
				int32(gofree.ThreadNum()),
				m.memory,
				t.ID.String(),
				t.CreatedAt,
				inbound,
				t.Metadata.Network,
				t.Protocol,
				processPath,
				packageName,
				source_ip,
				t.Metadata.Source.Port,
				domain,
				destination_ip,
				t.Metadata.Destination.Port,
				t.Upload.Load(),
				t.Download.Load(),
				0,
				0,
				rule0,
				rule1,
				chain0,
				chain1,
				t.OutboundType)
			if err != nil {
				m.logger.WarnContext(m.ctx, "db stmt exec: ", err)
				return false
			}
			return true
		})

		err = tx.Commit()
		if err != nil {
			m.logger.WarnContext(m.ctx, "db transaction commit: ", err)
			return
		}
	}
}
func (m *Manager) Close() error { //karing
	m.ticker.Stop()
	m.dbTicker.Stop()
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

func createTableSQL() string { //karing
	return `
CREATE TABLE IF NOT EXISTS connections_track (
    core_start DATETIME,
    total_upload INTEGER,
	total_download INTEGER,
	total_upload_speed INTEGER,
	total_download_speed INTEGER,
	total_upload_direct INTEGER,
	total_download_direct INTEGER,
	connections_in INTEGER,
	connections_out INTEGER,
	goroutines INTEGER,
	thread INTEGER,
	memory INTEGER,
    connection_id TEXT,
	create_at DATETIME,
    inbound TEXT,
	network TEXT,
	protocol TEXT,
	process TEXT,
	package TEXT,
	source_ip TEXT,
	source_port INTEGER,
	host TEXT,
	destination_ip TEXT,
	destination_port INTEGER,
	upload INTEGER,
	download INTEGER,
	upload_speed INTEGER,
	download_speed INTEGER,
	rule0 TEXT, 
	rule1 TEXT, 
	chain0 TEXT,
	chain1 TEXT,
	outbound_type TEXT 
);`
}

func deleteOldSQL() string { //karing
	return `DELETE FROM connections_track WHERE create_at < datetime('now', '-7 day');`
}

func prepareSQL() string { //karing
	return `
INSERT INTO connections_track(
    core_start ,
    total_upload ,
	total_download ,
	total_upload_speed ,
	total_download_speed ,
	total_upload_direct ,
	total_download_direct ,
	connections_in ,
	connections_out ,
	goroutines ,
	thread ,
	memory ,
    connection_id ,
	create_at ,
    inbound ,
	network ,
	protocol ,
	process ,
	package ,
	source_ip ,
	source_port ,
	host ,
	destination_ip ,
	destination_port ,
	upload ,
	download ,
	upload_speed ,
	download_speed ,
	rule0 , 
	rule1 , 
	chain0 ,
	chain1 ,
	outbound_type ) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
}
