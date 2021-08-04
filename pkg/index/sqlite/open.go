package sqlite

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// We use the driver directly rather than the database/sql interfaces, because the generic
// way tries to autodetect and normalize the return types. But we want to get the real types.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec(`PRAGMA synchronous = OFF;`); err != nil {
		return nil, err
	}
	if _, err = db.Exec(`PRAGMA journal_mode = OFF;`); err != nil {
		return nil, err
	}

	return db, nil
}
