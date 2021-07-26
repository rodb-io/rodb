package index

import (
	"database/sql/driver"
	gosqlite "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	sqlitePackage "rodb.io/pkg/index/sqlite"
	"rodb.io/pkg/input"
	"rodb.io/pkg/input/record"
	"rodb.io/pkg/parser"
	"strings"
	"testing"
)

func TestSqlite(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		index, err := NewSqlite(
			&SqliteConfig{
				Name: "testIndex",
				Properties: []SqlitePropertyConfig{
					{Name: "col", Collate: "NOCASE"},
				},
				Dsn:    ":memory:",
				Input:  "input",
				Logger: logrus.NewEntry(logrus.StandardLogger()),
			},
			input.List{
				"input": input.NewMock(parser.NewMock(), []record.Record{
					record.NewStringPropertiesMockRecord(map[string]string{
						"col": "value 1",
					}, 1),
					record.NewStringPropertiesMockRecord(map[string]string{
						"col": "value 2",
					}, 2),
					record.NewStringPropertiesMockRecord(map[string]string{
						"col": "value 3",
					}, 3),
				}),
			},
		)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		rows, err := index.db.Query(`
			SELECT "offset", "property_col"
			FROM "rodb_testIndex_index";
		`, []driver.Value{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		defer rows.Close()

		data := make([]driver.Value, 4)

		if err = rows.Next(data); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got, expect := data[0], int64(1); got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}
		if got, expect := data[1], "value 1"; got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}

		if err = rows.Next(data); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got, expect := data[0], int64(2); got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}
		if got, expect := data[1], "value 2"; got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}

		if err = rows.Next(data); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got, expect := data[0], int64(3); got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}
		if got, expect := data[1], "value 3"; got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}

		// Checking the column type
		tableDefinitionRows, err := index.db.Query(`
			SELECT sql
			FROM sqlite_master
			WHERE type = 'table'
			AND tbl_name = 'rodb_testIndex_index'
		`, []driver.Value{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		defer tableDefinitionRows.Close()

		data = make([]driver.Value, 1)
		if err = tableDefinitionRows.Next(data); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if expect := `"property_col" BLOB COLLATE NOCASE`; !strings.Contains(data[0].(string), expect) {
			t.Fatalf("Expected the column definition to be '%v', got '%v'\n", expect, data[0])
		}
	})
	t.Run("load", func(t *testing.T) {
		config := &SqliteConfig{
			Name: "testIndex",
			Properties: []SqlitePropertyConfig{
				{Name: "col", Collate: "BINARY"},
			},
			Dsn:    "file:memdb1?mode=memory&cache=shared",
			Input:  "input",
			Logger: logrus.NewEntry(logrus.StandardLogger()),
		}
		inputs := input.List{
			"input": input.NewMock(parser.NewMock(), []record.Record{
				record.NewStringPropertiesMockRecord(map[string]string{
					"col": "value 1",
				}, 1),
				record.NewStringPropertiesMockRecord(map[string]string{
					"col": "value 2",
				}, 2),
			}),
		}

		genericDb, err := (&gosqlite.SQLiteDriver{}).Open(config.Dsn)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		defer genericDb.Close()
		db := genericDb.(*gosqlite.SQLiteConn)

		metadata, err := sqlitePackage.NewMetadata(db, "testIndex", inputs["input"])
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		if err := metadata.Save(); err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		// Creating the data table without data
		_, err = db.Exec(`
			CREATE TABLE "rodb_testIndex_index" (
				"offset" INTEGER NOT NULL,
				"property_col" BLOB
			);
		`, []driver.Value{})
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		index, err := NewSqlite(config, inputs)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		rows, err := index.db.Query(`
			SELECT COUNT(*)
			FROM "rodb_testIndex_index";
		`, []driver.Value{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		defer rows.Close()

		data := make([]driver.Value, 1)
		if err = rows.Next(data); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got, expect := data[0], int64(0); got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}
	})
}

func TestSqliteGetRecordPositions(t *testing.T) {
	createTestData := func(t *testing.T, testName string) (*input.Mock, *Sqlite) {
		mockInput := input.NewMock(parser.NewMock(), []record.Record{
			record.NewStringPropertiesMockRecord(map[string]string{
				"col":  "col_a",
				"col2": "col2_b",
			}, 0),
			record.NewStringPropertiesMockRecord(map[string]string{
				"col":  "col_a",
				"col2": "col2_a",
			}, 1),
			record.NewStringPropertiesMockRecord(map[string]string{
				"col":  "col_b",
				"col2": "col2_a",
			}, 2),
			record.NewStringPropertiesMockRecord(map[string]string{
				"col":  "col_a",
				"col2": "col2_a",
			}, 3),
			record.NewStringPropertiesMockRecord(map[string]string{
				"col":  "col_b",
				"col2": "col2_b",
			}, 4),
		})
		index, err := NewSqlite(
			&SqliteConfig{
				Name: "testIndex",
				Properties: []SqlitePropertyConfig{
					{Name: "col", Collate: "BINARY"},
					{Name: "col2", Collate: "BINARY"},
				},
				Dsn:    ":memory:",
				Input:  "input",
				Logger: logrus.NewEntry(logrus.StandardLogger()),
			},
			input.List{
				"input": mockInput,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		return mockInput, index
	}

	t.Run("normal", func(t *testing.T) {
		mockInput, index := createTestData(t, "normal")
		for _, testCase := range []struct {
			expectedLength  int
			expectedResults record.PositionList
			filters         map[string]interface{}
		}{
			{
				expectedLength:  2,
				expectedResults: record.PositionList{1, 3},
				filters: map[string]interface{}{
					"col":  "col_a",
					"col2": "col2_a",
				},
			}, {
				expectedLength:  1,
				expectedResults: record.PositionList{2},
				filters: map[string]interface{}{
					"col":  "col_b",
					"col2": "col2_a",
				},
			}, {
				expectedLength:  2,
				expectedResults: record.PositionList{0, 4},
				filters: map[string]interface{}{
					"col2": "col2_b",
				},
			},
		} {
			nextPosition, err := index.GetRecordPositions(mockInput, testCase.filters)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			positions := make([]record.Position, 0)
			for {
				position, err := nextPosition()
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if position == nil {
					break
				}
				positions = append(positions, *position)
			}

			if got, expect := len(positions), testCase.expectedLength; got != expect {
				t.Fatalf("Expected %v positions, got %v, testCase: %+v", expect, got, testCase)
			}

			for i, position := range testCase.expectedResults {
				if position != positions[i] {
					t.Fatalf("Expected position %v at index %v, got %v", position, i, positions[i])
				}
			}
		}
	})
	t.Run("no filters", func(t *testing.T) {
		mockInput, index := createTestData(t, "no-filters")
		_, err := index.GetRecordPositions(mockInput, map[string]interface{}{})
		if err == nil {
			t.Fatalf("Expected an error, got %v", err)
		}
	})
	t.Run("wrong property", func(t *testing.T) {
		mockInput, index := createTestData(t, "wrong-property")
		_, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"wrong_col": "",
		})
		if err == nil {
			t.Fatalf("Expected an error, got %v", err)
		}
	})
}
