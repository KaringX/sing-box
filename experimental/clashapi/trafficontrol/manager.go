package trafficontrol

import (
	"context"
	"fmt"
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
	"github.com/sagernet/sing/service/pause"
)

var coreStartTime time.Time //karing
var coreUuid string         //karing

type DeviceEventTracker struct { //karing
	CreatedAt time.Time
	Name      string
}

type Manager struct {
	ctx                         context.Context               //karing
	logger                      log.ContextLogger             //karing
	pause                       pause.Manager                 //karing
	pauseCallback               *list.Element[pause.Callback] //karing
	uploadTotal                 atomic.Int64
	downloadTotal               atomic.Int64
	persistAccess               sync.Mutex                    //karing
	eventsForPersist            list.List[DeviceEventTracker] //karing
	closedConnectionsForPersist list.List[TrackerMetadata]    //karing
	connections                 compatible.Map[uuid.UUID, Tracker]
	closedConnectionsAccess     sync.Mutex
	closedConnections           list.List[TrackerMetadata]
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

func NewManager(ctx context.Context, logFactory log.ObservableFactory) *Manager { //karing
	///return &Manager{}//karing
	manager := &Manager{ //karing
		ctx:       ctx,
		logger:    logFactory.NewLogger("trafficontrolmanager"),
		startTime: time.Now(),
		ticker:    time.NewTicker(time.Second),
		done:      make(chan struct{}),
		// process: &process.Process{Pid: int32(os.Getpid())},
	}
	restart := true
	if coreStartTime.IsZero() {
		restart = false
		coreStartTime = manager.startTime
		id, _ := uuid.NewV4()
		coreUuid = id.String()
	}
	dbFile := service.FromContext[adapter.DBFile](ctx) //karing
	if dbFile != nil {                                 //karing
		err := dbFile.Exec(createTableSQL())
		if err != nil {
			manager.logger.WarnContext(manager.ctx, "create table connection_track: ", err)
		} else {
			go func() {
				dbFile.Exec(deleteOldSQL(dbFile.CacheDays()))
			}()
			manager.pause = service.FromContext[pause.Manager](ctx)
			if manager.pause != nil {
				manager.pauseCallback = manager.pause.RegisterCallback(manager.onPauseUpdated)
			}

			if restart {
				manager.addNewEvent("core:restart")
			} else {
				manager.addNewEvent("core:start")
			}
		}
	}
	go manager.handle() //karing
	return manager      //karing
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

		dbFile := service.FromContext[adapter.DBFile](m.ctx) //karing
		if dbFile != nil {                                   //karing
			m.persistAccess.Lock()
			defer m.persistAccess.Unlock()
			m.closedConnectionsForPersist.PushBack(metadata)
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

func (m *Manager) ConnectionsForPersist() []TrackerMetadata { //karing
	var connections []TrackerMetadata
	m.connections.Range(func(_ uuid.UUID, value Tracker) bool {
		md := value.Metadata()
		if !md.Dirty.Load() {
			return true
		}
		md.Dirty.Store(false)
		md.UploadSpeed = md.UploadBlip.Load()
		md.DownloadSpeed = md.DownloadBlip.Load()
		connections = append(connections, md)
		md.UploadBlip.Store(0)
		md.DownloadBlip.Store(0)
		return true
	})
	return connections
}

func (m *Manager) ClosedConnectionsForPersist() []TrackerMetadata { //karing
	m.persistAccess.Lock()
	defer m.persistAccess.Unlock()
	data := m.closedConnectionsForPersist.Array()
	for i := range data {
		data[i].UploadSpeed = data[i].UploadBlip.Load()
		data[i].DownloadSpeed = data[i].DownloadBlip.Load()
	}
	for !m.closedConnectionsForPersist.IsEmpty() {
		m.closedConnectionsForPersist.PopFront()
	}
	return data
}

func (m *Manager) EventsForPersist() []DeviceEventTracker { //karing
	m.persistAccess.Lock()
	defer m.persistAccess.Unlock()
	data := m.eventsForPersist.Array()
	for !m.eventsForPersist.IsEmpty() {
		m.eventsForPersist.PopFront()
	}
	return data
}

func (m *Manager) Connection(id uuid.UUID) Tracker {
	connection, loaded := m.connections.Load(id)
	if !loaded {
		return nil
	}
	return connection
}

func (m *Manager) onPauseUpdated(event int) {
	var name string
	switch event {
	case pause.EventDevicePaused:
		name = "device:pause"
	case pause.EventNetworkPause:
		name = "network:pause"
	case pause.EventDeviceWake:
		name = "device:wake"
	case pause.EventNetworkWake:
		name = "network:wake"
	default:
		return
	}
	m.addNewEvent(name)
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

		dbFile := service.FromContext[adapter.DBFile](m.ctx) //karing
		if dbFile != nil {                                   //karing
			m.persistDeviceEventsToDB(m.EventsForPersist())
			m.persistConnectionsToDB(m.ClosedConnectionsForPersist(), nil)
			m.persistConnectionsToDB(m.ConnectionsForPersist(), nil)
		}
	}
}

func (m *Manager) addNewEvent(name string) {
	m.persistAccess.Lock()
	defer m.persistAccess.Unlock()
	m.eventsForPersist.PushBack(DeviceEventTracker{CreatedAt: time.Now(), Name: name})
}

func (m *Manager) persistConnectionsToDB(connections []TrackerMetadata, closeAt *time.Time) { //karing
	if len(connections) == 0 {
		return
	}
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
		now := time.Now()
		for _, t := range connections {
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
			var connectionCloseAt *time.Time
			if closeAt != nil {
				connectionCloseAt = closeAt
			} else {
				if !t.ClosedAt.IsZero() {
					connectionCloseAt = &t.ClosedAt
				}
			}
			_, err = stmt.Exec(
				coreStartTime,
				m.startTime,
				now,
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
				connectionCloseAt,
				"",
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
				t.UploadSpeed,
				t.DownloadSpeed,
				rule0,
				rule1,
				chain0,
				chain1,
				t.OutboundType)
			if err != nil {
				m.logger.WarnContext(m.ctx, "db stmt exec: ", err)

			}
		}

		err = tx.Commit()
		if err != nil {
			m.logger.WarnContext(m.ctx, "db transaction commit: ", err)
			return
		}
	}
}

func (m *Manager) persistDeviceEventsToDB(events []DeviceEventTracker) { //karing
	if len(events) == 0 {
		return
	}
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

		now := time.Now()
		for _, t := range events {
			_, err = stmt.Exec(
				coreStartTime,
				m.startTime,
				now,
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
				coreUuid,
				t.CreatedAt,
				nil,
				t.Name,
				"",
				"",
				"",
				"",
				"",
				"",
				0,
				"",
				"",
				0,
				0,
				0,
				0,
				0,
				"",
				"",
				"",
				"",
				"")
			if err != nil {
				m.logger.WarnContext(m.ctx, "db stmt exec: ", err)
			}
		}

		err = tx.Commit()
		if err != nil {
			m.logger.WarnContext(m.ctx, "db transaction commit: ", err)
			return
		}
	}
}

func (m *Manager) Close() error { //karing
	m.ticker.Stop()
	close(m.done)

	if m.pauseCallback != nil {
		m.pause.UnregisterCallback(m.pauseCallback)
		m.pauseCallback = nil
	}

	dbFile := service.FromContext[adapter.DBFile](m.ctx)
	if dbFile != nil {
		/*var uploadTemp int64
		var downloadTemp int64

		uploadTemp = m.uploadTemp.Swap(0)
		downloadTemp = m.downloadTemp.Swap(0)
		m.uploadBlip.Store(uploadTemp)
		m.downloadBlip.Store(downloadTemp)

		closeAt := time.Now()
		m.persistConnectionsToDB(m.ClosedConnectionsForPersist(), &closeAt)
		m.persistConnectionsToDB(m.ConnectionsForPersist(), nil)*/

		m.addNewEvent("core:stop")
		m.persistDeviceEventsToDB(m.EventsForPersist())
	}
	m.connections.Clear()

	m.closedConnectionsAccess.Lock()
	defer m.closedConnectionsAccess.Unlock()
	for !m.closedConnectionsForPersist.IsEmpty() {
		m.closedConnectionsForPersist.PopFront()
	}

	m.persistAccess.Lock()
	defer m.persistAccess.Unlock()
	for !m.eventsForPersist.IsEmpty() {
		m.eventsForPersist.PopFront()
	}
	m.startTime = time.Now() //karing
	m.ResetStatistic()

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
CREATE TABLE IF NOT EXISTS records (
    core_start DATETIME,
	last_start DATETIME,
	persist DATETIME,
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
    session_id TEXT,
	begin DATETIME,
	end DATETIME,
	method TEXT,
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

func deleteOldSQL(cacheDays int) string { //karing
	if cacheDays <= 0 {
		cacheDays = 7
	}

	return fmt.Sprintf(`DELETE FROM records WHERE create_at < datetime('now', '-%d day');`, cacheDays)
}

func prepareSQL() string { //karing
	return `
INSERT INTO records(
    core_start ,
	last_start ,
	persist ,
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
    session_id ,
	begin ,
	end ,
	method ,
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
	outbound_type ) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
}
