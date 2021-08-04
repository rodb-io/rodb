package sqlite

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func SanitizeIdentifier(db *sql.DB, identifier string) (string, error) {
	row := db.QueryRow("SELECT printf('%w', ?)", identifier)
	if err := row.Err(); err != nil {
		return "", err
	}

	var sanitizedIdentifier string
	if err := row.Scan(&sanitizedIdentifier); err != nil {
		return "", err
	}

	return `"` + sanitizedIdentifier + `"`, nil
}
