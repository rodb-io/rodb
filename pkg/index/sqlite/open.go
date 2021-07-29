package sqlite

import (
	"database/sql/driver"
	gosqlite "github.com/mattn/go-sqlite3"
)

// We use the driver directly rather than the database/sql interfaces, because the generic
// way tries to autodetect and normalize the return types. But we want to get the real types.
func Open(dsn string) (*gosqlite.SQLiteConn, error) {
	genericDb, err := (&gosqlite.SQLiteDriver{}).Open(dsn)
	if err != nil {
		return nil, err
	}

	db := genericDb.(*gosqlite.SQLiteConn)

	if _, err = db.Exec(`PRAGMA synchronous = OFF;`, []driver.Value{}); err != nil {
		return nil, err
	}
	if _, err = db.Exec(`PRAGMA journal_mode = OFF;`, []driver.Value{}); err != nil {
		return nil, err
	}

	return db, nil
}
