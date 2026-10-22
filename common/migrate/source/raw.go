package source

import (
	"io"
	"io/fs"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4/source"
	E "github.com/sagernet/sing/common/exceptions"
)

type RawDriver struct {
	migrations    *source.Migrations
	rawMigrations map[string]string
}

func NewRawDriver(rawMigrations map[string]string) *RawDriver {
	return &RawDriver{rawMigrations: rawMigrations}
}

func (d *RawDriver) Init() error {
	ms := source.NewMigrations()
	for key := range d.rawMigrations {
		m, err := source.DefaultParse(key)
		if err != nil {
			continue
		}
		if !ms.Append(m) {
			return source.ErrDuplicateMigration{
				Migration: *m,
			}
		}
	}
	d.migrations = ms
	return nil
}

func (d *RawDriver) Open(url string) (source.Driver, error) {
	return nil, E.New("open() cannot be called")
}

func (d *RawDriver) Close() error {
	return nil
}

func (d *RawDriver) First() (version uint, err error) {
	if version, ok := d.migrations.First(); ok {
		return version, nil
	}
	return 0, &fs.PathError{
		Op:  "first",
		Err: fs.ErrNotExist,
	}
}

func (d *RawDriver) Prev(version uint) (prevVersion uint, err error) {
	if version, ok := d.migrations.Prev(version); ok {
		return version, nil
	}
	return 0, &fs.PathError{
		Op:  "prev for version " + strconv.FormatUint(uint64(version), 10),
		Err: fs.ErrNotExist,
	}
}

func (d *RawDriver) Next(version uint) (nextVersion uint, err error) {
	if version, ok := d.migrations.Next(version); ok {
		return version, nil
	}
	return 0, &fs.PathError{
		Op:  "next for version " + strconv.FormatUint(uint64(version), 10),
		Err: fs.ErrNotExist,
	}
}

func (d *RawDriver) ReadUp(version uint) (r io.ReadCloser, identifier string, err error) {
	if m, ok := d.migrations.Up(version); ok {
		body := ioutil.NopCloser(strings.NewReader(d.rawMigrations[m.Raw]))
		return body, m.Identifier, nil
	}
	return nil, "", &fs.PathError{
		Op:  "read up for version " + strconv.FormatUint(uint64(version), 10),
		Err: fs.ErrNotExist,
	}
}

func (d *RawDriver) ReadDown(version uint) (r io.ReadCloser, identifier string, err error) {
	if m, ok := d.migrations.Down(version); ok {
		body := ioutil.NopCloser(strings.NewReader(d.rawMigrations[m.Raw]))
		return body, m.Identifier, nil
	}
	return nil, "", &fs.PathError{
		Op:  "read down for version " + strconv.FormatUint(uint64(version), 10),
		Err: fs.ErrNotExist,
	}
}
