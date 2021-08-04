package index

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
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
				Properties: []*SqlitePropertyConfig{
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
		`)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		defer rows.Close()

		var offset int64
		var value string

		rows.Next()
		if err = rows.Scan(&offset, &value); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got, expect := offset, int64(1); got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}
		if got, expect := value, "value 1"; got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}

		rows.Next()
		if err = rows.Scan(&offset, &value); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got, expect := offset, int64(2); got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}
		if got, expect := value, "value 2"; got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}

		rows.Next()
		if err = rows.Scan(&offset, &value); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got, expect := offset, int64(3); got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}
		if got, expect := value, "value 3"; got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}

		if err := rows.Close(); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Checking the column type
		tableDefinitionRow := index.db.QueryRow(`
			SELECT sql
			FROM sqlite_master
			WHERE type = 'table'
			AND tbl_name = 'rodb_testIndex_index'
		`)
		if err := tableDefinitionRow.Err(); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		var definition string
		if err := tableDefinitionRow.Scan(&definition); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if expect := `"property_col" BLOB COLLATE NOCASE`; !strings.Contains(definition, expect) {
			t.Fatalf("Expected the column definition to be '%v', got '%v'\n", expect, definition)
		}
	})
	t.Run("load", func(t *testing.T) {
		config := &SqliteConfig{
			Name: "testIndex",
			Properties: []*SqlitePropertyConfig{
				{Name: "col", Collate: "BINARY"},
			},
			Dsn:    "file:memorysqlite?mode=memory&cache=shared",
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

		db, err := sql.Open("sqlite3", config.Dsn)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		defer db.Close()

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
		`)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		index, err := NewSqlite(config, inputs)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		row := index.db.QueryRow(`
			SELECT COUNT(*)
			FROM "rodb_testIndex_index";
		`)
		if err := row.Err(); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		var count int64
		if err = row.Scan(&count); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got, expect := count, int64(0); got != expect {
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
				Properties: []*SqlitePropertyConfig{
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
