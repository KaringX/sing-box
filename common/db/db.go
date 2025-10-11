package db

// karing
import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/service/filemanager"
)

type DBFile struct {
	ctx context.Context
	db  *sql.DB
}

func New(ctx context.Context, options *option.DBFileOptions) (*DBFile, error) {
	dbPath := filemanager.BasePath(ctx, options.Path)
	db, err := sql.Open("sqlite3", dbPath) //+":locked.sqlite?cache=shared"
	if err != nil {
		return nil, err
	}
	return &DBFile{ctx: ctx, db: db}, nil
}

func (c *DBFile) Name() string {
	return "db-file"
}

func (c *DBFile) Start(stage adapter.StartStage) error {
	return nil
}

func (d *DBFile) Close() error {
	if d.db == nil {
		return nil
	}
	return d.db.Close()
}

func (d *DBFile) CreateTable(sql string) error {
	if d.db == nil {
		return nil
	}
	_, err := d.db.Exec(sql)
	return err
}

func (d *DBFile) Prepare(tx *sql.Tx, sql string) (*sql.Stmt, error) {
	if d.db == nil || tx == nil {
		return nil, nil
	}
	return tx.Prepare(sql)
}

func (d *DBFile) Exec(stmt *sql.Stmt, args ...any) (sql.Result, error) {
	if d.db == nil || stmt == nil {
		return nil, nil
	}
	return stmt.Exec(args)
}

func (d *DBFile) BeginTx() (*sql.Tx, error) {
	if d.db == nil {
		return nil, nil
	}
	return d.db.Begin()
}

func (d *DBFile) Commit(tx *sql.Tx) error {
	if d.db == nil || tx == nil {
		return nil
	}
	return tx.Commit()
}
