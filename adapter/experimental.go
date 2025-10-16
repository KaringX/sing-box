package adapter

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"time"

	"github.com/sagernet/sing/common/varbin"
)

type ClashServer interface {
	LifecycleService
	ConnectionTracker
	Mode() string
	ModeList() []string
	HistoryStorage() URLTestHistoryStorage
}

type URLTestHistory struct {
	Time  time.Time `json:"time"`
	Delay uint16    `json:"delay"`
	Err   string    `json:"err,omitempty"` // karing
}

type URLTestHistoryStorage interface {
	SetHook(hook chan<- struct{})
	LoadURLTestHistory(tag string) *URLTestHistory
	DeleteURLTestHistory(tag string)
	StoreURLTestHistory(tag string, history *URLTestHistory)
	Close() error
	GetURLTestHistory() map[string]*URLTestHistory // karing
}

type URLTestResult struct { // karing
	Delay uint16 `json:"delay,omitempty"`
	Err   string `json:"err,omitempty"`
}

type V2RayServer interface {
	LifecycleService
	StatsService() ConnectionTracker
}

type CacheFile interface {
	BeforeStart() error //karing
	LifecycleService

	StoreFakeIP() bool
	FakeIPStorage

	StoreRDRC() bool
	RDRCStore

	LoadMode() string
	StoreMode(mode string) error
	LoadSelected(group string) string
	StoreSelected(group string, selected string) error
	LoadGroupExpand(group string) (isExpand bool, loaded bool)
	StoreGroupExpand(group string, expand bool) error
	LoadRuleSet(tag string) *SavedBinary
	SaveRuleSet(tag string, set *SavedBinary) error
	DeleteRuleSet(tag string)           //karing
	HasRuleSet(tag string) bool         //karing
	GetAllRuleSetKeys() map[string]bool //karing
}

type DBFile interface { //karing
	LifecycleService
	Exec(sql string) (sql.Result, error)
	Query(query string, args ...any) (*sql.Row, error)
	Prepare(tx *sql.Tx, sql string) (*sql.Stmt, error)
	ExecStmt(stmt *sql.Stmt, args ...any) (sql.Result, error)
	BeginTx() (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	CacheDays() int
}

type SavedBinary struct {
	Content     []byte
	LastUpdated time.Time
	LastEtag    string
}

func (s *SavedBinary) MarshalBinary() ([]byte, error) {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, uint8(1))
	if err != nil {
		return nil, err
	}
	err = varbin.Write(&buffer, binary.BigEndian, s.Content)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buffer, binary.BigEndian, s.LastUpdated.Unix())
	if err != nil {
		return nil, err
	}
	err = varbin.Write(&buffer, binary.BigEndian, s.LastEtag)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (s *SavedBinary) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)
	var version uint8
	err := binary.Read(reader, binary.BigEndian, &version)
	if err != nil {
		return err
	}
	err = varbin.Read(reader, binary.BigEndian, &s.Content)
	if err != nil {
		return err
	}
	var lastUpdated int64
	err = binary.Read(reader, binary.BigEndian, &lastUpdated)
	if err != nil {
		return err
	}
	s.LastUpdated = time.Unix(lastUpdated, 0)
	err = varbin.Read(reader, binary.BigEndian, &s.LastEtag)
	if err != nil {
		return err
	}
	return nil
}

type OutboundGroup interface {
	Outbound
	Now() string
	All() []string
}

type URLTestGroup interface {
	OutboundGroup
	URLTest(ctx context.Context, force bool) (map[string]URLTestResult, error) //karing
	UpdateCheck()                                                              //karing
}

func OutboundTag(detour Outbound) string {
	if group, isGroup := detour.(OutboundGroup); isGroup {
		return group.Now()
	}
	return detour.Tag()
}
