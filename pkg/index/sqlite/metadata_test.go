package sqlite

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"rodb.io/pkg/input"
	"rodb.io/pkg/input/record"
	"rodb.io/pkg/parser"
	"testing"
	"time"
)

func TestMetadataNewMetadata(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		defer db.Close()

		input := input.NewMock(parser.NewMock(), make([]record.Record, 42))
		input.SetModTime(time.Unix(1234, 0))

		metadata, err := NewMetadata(db, "testIndex", input)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := CurrentVersion, metadata.version; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(1234), metadata.inputFileModificationTime.Unix(); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(42), metadata.inputFileSize; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := false, metadata.completed; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestMetadataLoadMetadata(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		defer db.Close()

		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS "rodb_testIndex_metadata" (
				"version" INTEGER NOT NULL,
				"inputFileModificationTime" INTEGER NOT NULL,
				"inputFileSize" INTEGER NOT NULL,
				"completed" BOOLEAN NOT NULL
			);
		`)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		_, err = db.Exec(`
			INSERT INTO "rodb_testIndex_metadata" (
				"version",
				"inputFileModificationTime",
				"inputFileSize",
				"completed"
			) VALUES (2, 1234, 42, 1);
		`)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		metadata, err := LoadMetadata(db, "testIndex")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := uint16(2), metadata.version; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(1234), metadata.inputFileModificationTime.Unix(); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(42), metadata.inputFileSize; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := true, metadata.completed; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestMetadataHasMetadata(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		defer db.Close()

		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS "rodb_testIndex_metadata" (
				"version" INTEGER NOT NULL,
				"inputFileModificationTime" INTEGER NOT NULL,
				"inputFileSize" INTEGER NOT NULL,
				"completed" BOOLEAN NOT NULL
			);
		`)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		_, err = db.Exec(`
			INSERT INTO "rodb_testIndex_metadata" (
				"version",
				"inputFileModificationTime",
				"inputFileSize",
				"completed"
			) VALUES (2, 1234, 42, 1);
		`)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		got, err := HasMetadata(db, "testIndex")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect := true; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestMetadataSave(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		defer db.Close()

		metadata := Metadata{
			db:                        db,
			indexName:                 "testIndex",
			version:                   1,
			inputFileModificationTime: time.Unix(1234, 0),
			inputFileSize:             42,
			completed:                 false,
		}
		if err := metadata.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		row := db.QueryRow(`
			SELECT
				"version",
				"inputFileModificationTime",
				"inputFileSize",
				"completed"
			FROM "rodb_testIndex_metadata";
		`)
		if err := row.Err(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		var version int64
		var inputFileModificationTime int64
		var inputFileSize int64
		var completed bool
		if err = row.Scan(&version, &inputFileModificationTime, &inputFileSize, &completed); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := int64(CurrentVersion), version; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(1234), inputFileModificationTime; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(42), inputFileSize; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := false, completed; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestMetadataAssertValid(t *testing.T) {
	modTime := time.Now()
	data := make([]record.Record, 42)
	input := input.NewMock(parser.NewMock(), data)
	input.SetModTime(modTime)

	metadata := Metadata{
		version:                   CurrentVersion,
		inputFileModificationTime: modTime,
		inputFileSize:             int64(len(data)),
		completed:                 true,
	}

	t.Run("valid", func(t *testing.T) {
		if err := metadata.AssertValid(input); err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
	t.Run("wrong version", func(t *testing.T) {
		metadata.version = CurrentVersion + 1
		if metadata.AssertValid(input) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("wrong time", func(t *testing.T) {
		metadata.inputFileModificationTime = time.Unix(1234, 0)
		if metadata.AssertValid(input) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("wrong size", func(t *testing.T) {
		metadata.inputFileSize = int64(len(data) + 1)
		if metadata.AssertValid(input) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("not completed", func(t *testing.T) {
		metadata.completed = false
		if metadata.AssertValid(input) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
}
