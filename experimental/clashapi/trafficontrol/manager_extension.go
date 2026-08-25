//go:build with_karing

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
	safe "github.com/sagernet/sing-box/common/fix"
	"github.com/sagernet/sing-box/common/gofree"
	"github.com/sagernet/sing-box/log"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/memory"
	"github.com/sagernet/sing/common/x/list"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
)

var coreStartTime time.Time
var coreRestart = false

type TrackerMetadataOut struct {
	CreatedAt   time.Time `json:"startTime"`
	Network     string    `json:"network"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Fqdn        string    `json:"fqdn"`
	Outbound    string    `json:"outbound"`
}

type DeviceEventTracker struct {
	CreatedAt time.Time
	Name      string
	ID        string
}

type ManagerExtension struct {
	ctx                         context.Context
	logger                      log.ContextLogger
	pause                       pause.Manager
	pauseCallback               *list.Element[pause.Callback]
	persistAccess               sync.Mutex
	eventsForPersist            list.List[DeviceEventTracker]
	closedConnectionsForPersist list.List[TrackerMetadata]
	ticker                      *time.Ticker
	dbSizeTicker                *time.Ticker
	dbCacheSizeLimited          atomic.Bool
	done                        chan struct{}
	memoryTotal                 uint64
	startTime                   time.Time
	uploadTemp                  atomic.Int64
	downloadTemp                atomic.Int64
	uploadBlip                  atomic.Int64
	downloadBlip                atomic.Int64
	uploadTotalDirect           atomic.Int64
	downloadTotalDirect         atomic.Int64
}

type SnapshotExtension struct {
	StartTime           time.Time
	DownloadDirect      int64
	UploadDirect        int64
	DownloadSpeed       int64
	UploadSpeed         int64
	ConnectionsOut      []TrackerMetadataOut
	ConnectionsOutCount int32
	ConnectionsInCount  int32
	Goroutines          int32
	ThreadCount         int32
}

func IsCoreRestart() bool {
	return coreRestart
}

func newManagerWithExtension(ctx context.Context, logFactory log.ObservableFactory) *Manager {
	manager := &Manager{
		ManagerExtension: ManagerExtension{ctx: ctx,
			logger:       logFactory.NewLogger("trafficontrolmanager"),
			startTime:    time.Now(),
			ticker:       time.NewTicker(time.Second),
			dbSizeTicker: time.NewTicker(time.Minute * 1),
			done:         make(chan struct{}),
		},
	}

	if coreStartTime.IsZero() {
		coreStartTime = manager.startTime
	} else {
		coreRestart = true
	}
	statistics := service.FromContext[adapter.Statistics](ctx)
	if statistics != nil {
		_, err := statistics.Exec(createTableSQL())
		if err != nil {
			manager.logger.WarnContext(manager.ctx, "create table connection_track: ", err)
		} else {
			go func() {
				statistics.Exec(deleteOldSQL(statistics.CacheDays()))
			}()
			manager.pause = service.FromContext[pause.Manager](ctx)
			if manager.pause != nil {
				manager.pauseCallback = manager.pause.RegisterCallback(manager.onPauseUpdated)
			}

			if coreRestart {
				manager.addNewEvent("core:restart")
			} else {
				manager.addNewEvent("core:start")
			}
		}
	}
	go manager.handle()
	go manager.handleDBSizeCheck()

	return manager
}

func (m *Manager) GetLatestDownloadTime(tag string) (bool, time.Time) {
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

func (m *Manager) getConnectionsForPersist() []TrackerMetadata {
	var connections []TrackerMetadata
	if m.dbCacheSizeLimited.Load() {
		return connections
	}
	m.connections.Range(func(_ uuid.UUID, value Tracker) bool {
		md := value.Metadata()
		if !md.Dirty.Load() {
			return true
		}
		md.Dirty.Store(false)
		md.UploadSpeed = md.UploadBlip.Load()
		md.DownloadSpeed = md.DownloadBlip.Load()
		connections = append(connections, *md)
		md.UploadBlip.Store(0)
		md.DownloadBlip.Store(0)
		return true
	})
	return connections
}

func (m *Manager) getClosedConnectionsForPersist() []TrackerMetadata {
	m.persistAccess.Lock()
	defer m.persistAccess.Unlock()
	data := m.closedConnectionsForPersist.Array()
	for !m.closedConnectionsForPersist.IsEmpty() {
		m.closedConnectionsForPersist.PopFront()
	}
	for i := range data {
		data[i].UploadSpeed = data[i].UploadBlip.Load()
		data[i].DownloadSpeed = data[i].DownloadBlip.Load()
	}
	return data
}

func (m *Manager) getEventsForPersist() []DeviceEventTracker {
	m.persistAccess.Lock()
	defer m.persistAccess.Unlock()
	data := m.eventsForPersist.Array()
	for !m.eventsForPersist.IsEmpty() {
		m.eventsForPersist.PopFront()
	}
	return data
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

		statistics := service.FromContext[adapter.Statistics](m.ctx)
		if statistics != nil {
			persistTime := time.Now()
			writeCount := 0
			writeCount += m.persistDeviceEventsToDB(m.getEventsForPersist(), persistTime)
			writeCount += m.persistConnectionsToDB(m.getClosedConnectionsForPersist(), persistTime, nil)
			writeCount += m.persistConnectionsToDB(m.getConnectionsForPersist(), persistTime, nil)
			if writeCount == 0 {
				m.addNewEvent("track:idle")
				m.persistDeviceEventsToDB(m.getEventsForPersist(), persistTime)
			}
		}
	}
}

func (m *Manager) handleDBSizeCheck() {
	m.dBSizeCheck()
	for {
		select {
		case <-m.done:
			return
		case <-m.dbSizeTicker.C:
			m.dBSizeCheck()
		}
	}
}

func (m *Manager) dBSizeCheck() {
	statistics := service.FromContext[adapter.Statistics](m.ctx)
	if statistics != nil {
		limit := statistics.CacheSizeLimit()
		if limit > 0 {
			dbSize := statistics.DBSize()
			if dbSize == -1 || dbSize > limit {
				if !m.dbCacheSizeLimited.Load() {
					m.logger.WarnContext(m.ctx, "statistics database size exceeds limit: ", dbSize, " bytes")
				}
				m.dbCacheSizeLimited.Store(true)
			} else {
				m.dbCacheSizeLimited.Store(false)
			}
		} else {
			m.dbCacheSizeLimited.Store(false)
		}
	}
}

func (m *Manager) resetStatistic() {
	m.uploadTemp.Store(0)
	m.uploadBlip.Store(0)
	m.downloadTemp.Store(0)
	m.downloadBlip.Store(0)
	m.uploadTotalDirect.Store(0)
	m.downloadTotalDirect.Store(0)
}

func (m *Manager) Close() error {
	m.ticker.Stop()
	m.dbSizeTicker.Stop()
	close(m.done)

	if m.pauseCallback != nil {
		m.pause.UnregisterCallback(m.pauseCallback)
		m.pauseCallback = nil
	}

	statistics := service.FromContext[adapter.Statistics](m.ctx)
	if statistics != nil {
		var uploadTemp int64
		var downloadTemp int64

		uploadTemp = m.uploadTemp.Swap(0)
		downloadTemp = m.downloadTemp.Swap(0)
		m.uploadBlip.Store(uploadTemp)
		m.downloadBlip.Store(downloadTemp)

		persistTime := time.Now()
		m.persistConnectionsToDB(m.getClosedConnectionsForPersist(), persistTime, &persistTime)
		m.persistConnectionsToDB(m.getConnectionsForPersist(), persistTime, &persistTime)

		m.addNewEvent("core:stop")
		m.persistDeviceEventsToDB(m.getEventsForPersist(), persistTime)
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
	m.startTime = time.Now()
	m.ResetStatistic()

	return nil
}

func (m *Manager) addNewEvent(name string) {
	if m.dbCacheSizeLimited.Load() {
		return
	}
	id, _ := uuid.NewV4()
	m.persistAccess.Lock()
	defer m.persistAccess.Unlock()
	m.eventsForPersist.PushBack(DeviceEventTracker{CreatedAt: time.Now(), Name: name, ID: id.String()})
}

func (m *Manager) persistConnectionsToDB(connections []TrackerMetadata, persistTime time.Time, closeAt *time.Time) int {
	if len(connections) == 0 {
		return 0
	}
	connectionManager := service.FromContext[adapter.ConnectionManager](m.ctx)
	statistics := service.FromContext[adapter.Statistics](m.ctx)
	if statistics != nil {
		tx, err := statistics.BeginTx()
		if err != nil {
			m.logger.WarnContext(m.ctx, "statistics begin transaction: ", err)
			return 0
		}
		stmt, err := tx.Prepare(prepareSQL())
		if err != nil {
			m.logger.WarnContext(m.ctx, "statistics transaction prepare: ", err)
			return 0
		}
		defer stmt.Close()

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		m.memory = memStats.StackInuse + memStats.HeapInuse + memStats.HeapIdle - memStats.HeapReleased
		m.memoryTotal = memory.Total()
		for _, t := range connections {
			var inbound string
			if t.Metadata.Inbound != "" {
				inbound = t.Metadata.InboundType + "/" + t.Metadata.Inbound
			} else {
				inbound = t.Metadata.InboundType
			}
			var domain string
			var processPath string
			var packageName string

			if !statistics.DataDesensitize() {
				if t.Metadata.Domain != "" {
					domain = t.Metadata.Domain
				} else {
					domain = t.Metadata.Destination.Fqdn
				}
				if t.Metadata.ProcessInfo != nil {
					if t.Metadata.ProcessInfo.ProcessPath != "" {
						processPath = t.Metadata.ProcessInfo.ProcessPath
					} else if len(t.Metadata.ProcessInfo.AndroidPackageNames) > 0 {
						packageName = t.Metadata.ProcessInfo.AndroidPackageNames[0]
					}
					if processPath == "" {
						if t.Metadata.ProcessInfo.UserId != -1 {
							processPath = F.ToString(t.Metadata.ProcessInfo.UserId)
						}
					} else if t.Metadata.ProcessInfo.UserName != "" {
						processPath = F.ToString(processPath, " (", t.Metadata.ProcessInfo.UserName, ")")
					} else if t.Metadata.ProcessInfo.UserId != -1 {
						processPath = F.ToString(processPath, " (", t.Metadata.ProcessInfo.UserId, ")")
					}
				}
			} else {
				domain = "*"
				processPath = "*"
				packageName = "*"
			}

			var rule0 string
			var rule1 string
			var chain0 string
			var chain1 string
			var outboundType string
			if !statistics.DataDesensitize() {
				if t.Rule != nil {
					rule0 = t.Rule.Name()
					rule1 = t.Rule.Action().Target()
				} else {
					rule0 = "final"
				}

				if len(t.Chain) == 1 {
					chain0 = t.Chain[0]
				} else if len(t.Chain) >= 2 {
					chain1 = t.Chain[0]
					chain0 = t.Chain[len(t.Chain)-1]
				}
				outboundType = t.OutboundType
			} else {
				rule0 = "*"
				rule1 = "*"
				chain1 = "*"
				chain0 = "*"
				outboundType = "*"
			}
			var source_ip string
			var destination_ip string
			if t.Metadata.Source.Addr.IsValid() {
				source_ip = safe.SafeAddrString(t.Metadata.Source.Addr) //karing
			}
			if !statistics.DataDesensitize() {
				if t.Metadata.Destination.Addr.IsValid() {
					destination_ip = safe.SafeAddrString(t.Metadata.Destination.Addr) //karing
				}
			} else {
				destination_ip = "*"
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
				persistTime,
				m.uploadTotal.Load(),
				m.downloadTotal.Load(),
				m.uploadBlip.Load(),
				m.downloadBlip.Load(),
				m.uploadTotalDirect.Load(),
				m.downloadTotalDirect.Load(),
				int32(m.connections.Len()),
				int32(connectionManager.Count()),
				int32(runtime.NumGoroutine()),
				int32(gofree.ThreadNum()),
				m.memoryTotal,
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
				outboundType)
			if err != nil {
				m.logger.WarnContext(m.ctx, "db stmt exec: ", err)
			}
		}

		err = tx.Commit()
		if err != nil {
			m.logger.WarnContext(m.ctx, "db transaction commit: ", err)
			return 0
		}
	}
	return len(connections)
}

func (m *Manager) persistDeviceEventsToDB(events []DeviceEventTracker, persistTime time.Time) int {
	if len(events) == 0 {
		return 0
	}
	connectionManager := service.FromContext[adapter.ConnectionManager](m.ctx)
	statistics := service.FromContext[adapter.Statistics](m.ctx)
	if statistics != nil {
		tx, err := statistics.BeginTx()
		if err != nil {
			m.logger.WarnContext(m.ctx, "statistics begin transaction: ", err)
			return 0
		}
		stmt, err := tx.Prepare(prepareSQL())
		if err != nil {
			m.logger.WarnContext(m.ctx, "statistics transaction prepare: ", err)
			return 0
		}
		defer stmt.Close()

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		m.memory = memStats.StackInuse + memStats.HeapInuse + memStats.HeapIdle - memStats.HeapReleased
		m.memoryTotal = memory.Total()
		for _, t := range events {
			_, err = stmt.Exec(
				coreStartTime,
				m.startTime,
				persistTime,
				m.uploadTotal.Load(),
				m.downloadTotal.Load(),
				m.uploadBlip.Load(),
				m.downloadBlip.Load(),
				m.uploadTotalDirect.Load(),
				m.downloadTotalDirect.Load(),
				int32(m.connections.Len()),
				int32(connectionManager.Count()),
				int32(runtime.NumGoroutine()),
				int32(gofree.ThreadNum()),
				m.memoryTotal,
				t.ID,
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
			return 0
		}
	}
	return len(events)
}

func createTableSQL() string {
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

func deleteOldSQL(cacheDays int) string {
	if cacheDays <= 0 {
		cacheDays = 7
	}

	return fmt.Sprintf(`DELETE FROM records WHERE core_start < datetime('now', '-%d day');`, cacheDays)
}

func prepareSQL() string {
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
