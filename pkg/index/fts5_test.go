package index

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	sqlitePackage "github.com/rodb-io/rodb/pkg/index/sqlite"
	"github.com/rodb-io/rodb/pkg/input"
	"github.com/rodb-io/rodb/pkg/input/record"
	"github.com/rodb-io/rodb/pkg/parser"
	"testing"
)

func TestFts5(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		index, err := NewFts5(
			&Fts5Config{
				Name:       "testIndex",
				Properties: []string{"col"},
				Dsn:        ":memory:",
				Input:      "input",
				Logger:     logrus.NewEntry(logrus.StandardLogger()),
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
			SELECT "__offset"
			FROM "rodb_testIndex_index"
			WHERE "rodb_testIndex_index" MATCH 'value';
		`)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		defer rows.Close()

		var offset int64

		rows.Next()
		if err = rows.Scan(&offset); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got, expect := offset, int64(1); got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}

		rows.Next()
		if err = rows.Scan(&offset); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got, expect := offset, int64(2); got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}

		rows.Next()
		if err = rows.Scan(&offset); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got, expect := offset, int64(3); got != expect {
			t.Fatalf("Expected %v, got %v\n", expect, got)
		}
	})
	t.Run("load", func(t *testing.T) {
		config := &Fts5Config{
			Name:       "testIndex",
			Properties: []string{"col"},
			Dsn:        "file:memoryfts5?mode=memory&cache=shared",
			Input:      "input",
			Logger:     logrus.NewEntry(logrus.StandardLogger()),
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
		metadata.SetCompleted(true)
		if err := metadata.Save(); err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		// Creating the data table without data
		_, err = db.Exec(`
			CREATE VIRTUAL TABLE "rodb_testIndex_index" USING fts5(
				"offset" UNINDEXED,
				"property_col"
			);
		`)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		index, err := NewFts5(config, inputs)
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

func TestFts5GetRecordPositions(t *testing.T) {
	createTestData := func(t *testing.T, testName string) (*input.Mock, *Fts5) {
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
		index, err := NewFts5(
			&Fts5Config{
				Name:       "testIndex",
				Properties: []string{"col", "col2"},
				Dsn:        ":memory:",
				Input:      "input",
				Logger:     logrus.NewEntry(logrus.StandardLogger()),
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
					"match": "col:col_a AND col2:col2_a",
				},
			}, {
				expectedLength:  1,
				expectedResults: record.PositionList{2},
				filters: map[string]interface{}{
					"match": "col:col_b AND col2:col2_a",
				},
			}, {
				expectedLength:  2,
				expectedResults: record.PositionList{0, 4},
				filters: map[string]interface{}{
					"match": "col2:col2_b",
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
			"not_match": "",
		})
		if err == nil {
			t.Fatalf("Expected an error, got %v", err)
		}
	})
	t.Run("multiple filters", func(t *testing.T) {
		mockInput, index := createTestData(t, "wrong-property")
		_, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"filter1": "",
			"filter2": "",
		})
		if err == nil {
			t.Fatalf("Expected an error, got %v", err)
		}
	})
}
