//go:build with_karing

package statistics

import (
	"context"
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/service/filemanager"
)

type Statistics struct {
	ctx             context.Context
	db              *sql.DB
	dbPath          string
	dataDesensitize bool
	cacheDays       int
	cacheSizeLimit  int64
}

// dbbackup go-sqlite3\_example\hook\hook.go
func New(ctx context.Context, options *option.StatisticsOptions) (*Statistics, error) {
	dbPath := filemanager.BasePath(ctx, options.Path)
	db, err := sql.Open("sqlite3", dbPath) //+":locked.sqlite?cache=shared"
	if err != nil {
		return nil, err
	}
	return &Statistics{ctx: ctx, db: db, dbPath: dbPath, dataDesensitize: options.DataDesensitize, cacheDays: options.CacheDays, cacheSizeLimit: options.CacheSizeLimit}, nil
}

func (c *Statistics) Name() string {
	return "statistics"
}

func (c *Statistics) Start(stage adapter.StartStage) error {
	return nil
}

func (d *Statistics) Close() error {
	if d.db == nil {
		return nil
	}
	return d.db.Close()
}

func (d *Statistics) Exec(sql string) (sql.Result, error) {
	if d.db == nil {
		return nil, nil
	}
	result, err := d.db.Exec(sql)
	return result, err
}

func (d *Statistics) Query(query string, args ...any) (*sql.Rows, error) {
	if d.db == nil {
		return nil, nil
	}
	return d.db.Query(query, args)
}

func (d *Statistics) Prepare(tx *sql.Tx, sql string) (*sql.Stmt, error) {
	if d.db == nil || tx == nil {
		return nil, nil
	}
	return tx.Prepare(sql)
}

func (d *Statistics) ExecStmt(stmt *sql.Stmt, args ...any) (sql.Result, error) {
	if d.db == nil || stmt == nil {
		return nil, nil
	}
	return stmt.Exec(args)
}

func (d *Statistics) BeginTx() (*sql.Tx, error) {
	if d.db == nil {
		return nil, nil
	}
	return d.db.Begin()
}

func (d *Statistics) Commit(tx *sql.Tx) error {
	if d.db == nil || tx == nil {
		return nil
	}
	return tx.Commit()
}

func (d *Statistics) DataDesensitize() bool {
	return d.dataDesensitize
}

func (d *Statistics) CacheDays() int {
	return d.cacheDays
}

func (d *Statistics) CacheSizeLimit() int64 {
	return d.cacheSizeLimit
}

func (d *Statistics) DBSize() int64 {
	if d.db == nil || d.dbPath == "" {
		return 0
	}
	info, err := os.Stat(d.dbPath)
	if err != nil {
		return -1
	}
	return info.Size()
}
