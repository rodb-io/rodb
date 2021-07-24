package sqlite

import (
	gosqlite "github.com/mattn/go-sqlite3"
	"testing"
)

func TestSanitizeIdentifier(t *testing.T) {
	getSanitizedIdentifier := func(identifier string) string {
		db, err := (&gosqlite.SQLiteDriver{}).Open(":memory:")
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		defer db.Close()

		result, err := SanitizeIdentifier(db.(*gosqlite.SQLiteConn), identifier)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		return result
	}

	t.Run("normal", func(t *testing.T) {
		if got, expect := getSanitizedIdentifier("table"), `"table"`; got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("space", func(t *testing.T) {
		if got, expect := getSanitizedIdentifier("some column"), `"some column"`; got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("double-quote", func(t *testing.T) {
		if got, expect := getSanitizedIdentifier(`foo"bar`), `"foo""bar"`; got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
}
