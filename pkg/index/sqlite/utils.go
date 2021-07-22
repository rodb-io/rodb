package sqlite

import (
	"database/sql/driver"
	gosqlite "github.com/mattn/go-sqlite3"
)

func SanitizeIdentifier(db *gosqlite.SQLiteConn, identifier string) (string, error) {
	rows, err := db.Query("SELECT printf('%w', ?)", []driver.Value{identifier})
	if err != nil {
		return "", err
	}
	defer rows.Close()

	data := make([]driver.Value, 1)
	if err = rows.Next(data); err != nil {
		return "", err
	}

	return `"`+data[0].(string)+`"`, nil
}
