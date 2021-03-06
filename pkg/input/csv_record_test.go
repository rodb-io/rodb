package input

import (
	"fmt"
	"github.com/rodb-io/rodb/pkg/parser"
	"testing"
)

func TestCsvAll(t *testing.T) {
	var testConfig *CsvConfig = &CsvConfig{
		Columns: []*CsvColumnConfig{
			{Name: "col_a"},
			{Name: "col_b"},
		},
		ColumnIndexByName: map[string]int{
			"col_a": 0,
			"col_b": 1,
		},
	}

	parsers := []parser.Parser{
		parser.NewMock(),
		parser.NewJson(&parser.JsonConfig{}),
	}

	t.Run("normal", func(t *testing.T) {
		record := NewCsvRecord(testConfig, parsers, []string{"string_a", `{"b": "string_b"}`}, 0)
		data, err := record.All()
		if err != nil {
			t.Fatalf("Got error: '%v'", err)
		}

		if expect := "string_a"; fmt.Sprintf("%v", data["col_a"]) != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, data["col_b"])
		}

		if expect := map[string]interface{}{"b": "string_b"}; fmt.Sprintf("%v", data["col_b"]) != fmt.Sprintf("%v", expect) {
			t.Fatalf("Expected to get '%v', got '%v'", expect, data["col_b"])
		}
	})
	t.Run("error if col does not exist", func(t *testing.T) {
		record := NewCsvRecord(testConfig, parsers, []string{}, 0)
		got, err := record.Get("col_0")
		if err == nil {
			t.Fatalf("Expected error, got '%v'", got)
		}
	})
	t.Run("col not found", func(t *testing.T) {
		record := NewCsvRecord(testConfig, parsers, []string{}, 0)
		got, err := record.Get("col_a")
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}
		if got != nil {
			t.Fatalf("Expected nil, got '%v'", got)
		}
	})
}

func TestCsvRecordGet(t *testing.T) {
	var testConfig *CsvConfig = &CsvConfig{
		Columns: []*CsvColumnConfig{
			{Name: "col"},
		},
		ColumnIndexByName: map[string]int{
			"col": 0,
		},
	}

	parsers := []parser.Parser{
		parser.NewJson(&parser.JsonConfig{}),
	}

	testCases := []struct {
		name        string
		json        string
		path        string
		expectError bool
		expectValue interface{}
	}{
		{
			name:        "primitive",
			json:        `"test"`,
			path:        "col",
			expectError: false,
			expectValue: "test",
		}, {
			name:        "primitive in array",
			json:        `["test"]`,
			path:        "col.0",
			expectError: false,
			expectValue: "test",
		}, {
			name:        "primitive in object",
			json:        `{"key": "test"}`,
			path:        "col.key",
			expectError: false,
			expectValue: "test",
		}, {
			name:        "primitive in array of objects",
			json:        `[{"key": "test"}]`,
			path:        "col.0.key",
			expectError: false,
			expectValue: "test",
		}, {
			name:        "primitive in array in an object",
			json:        `{"key": ["test"]}`,
			path:        "col.key.0",
			expectError: false,
			expectValue: "test",
		}, {
			name:        "array",
			json:        `["test"]`,
			path:        "col",
			expectError: false,
			expectValue: []string{"test"},
		}, {
			name:        "array in array",
			json:        `[42, ["test"]]`,
			path:        "col.1",
			expectError: false,
			expectValue: []string{"test"},
		}, {
			name:        "array in object",
			json:        `{"key": ["test"]}`,
			path:        "col.key",
			expectError: false,
			expectValue: []string{"test"},
		}, {
			name:        "object",
			json:        `{"key": "test"}`,
			path:        "col",
			expectError: false,
			expectValue: map[string]interface{}{"key": "test"},
		}, {
			name:        "object in array",
			json:        `[{"key": "test"}]`,
			path:        "col.0",
			expectError: false,
			expectValue: map[string]interface{}{"key": "test"},
		}, {
			name:        "object in object",
			json:        `{"keyRoot": {"key": "test"}}`,
			path:        "col.keyRoot",
			expectError: false,
			expectValue: map[string]interface{}{"key": "test"},
		}, {
			name:        "index out of range",
			json:        `["a", "b"]`,
			path:        "col.5",
			expectError: false,
			expectValue: nil,
		}, {
			name:        "missing property",
			json:        `{"a": null, "b": 42}`,
			path:        "col.nope",
			expectError: false,
			expectValue: nil,
		}, {
			name:        "wrong path from root",
			json:        `"test"`,
			path:        "a.b",
			expectError: true,
		}, {
			name:        "non-numeric key",
			json:        `["test"]`,
			path:        "col.key",
			expectError: true,
		}, {
			name:        "wrong path",
			json:        `{"key": "value"}`,
			path:        "col.key.test.42",
			expectError: true,
		}, {
			name:        "deep path",
			json:        `{"a": {"b": {"c": {"d": {"e": {"f": {"g": true}}}}}}}`,
			path:        "col.a.b.c.d.e.f.g",
			expectError: false,
			expectValue: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			record := NewCsvRecord(testConfig, parsers, []string{testCase.json}, 0)

			got, err := record.Get(testCase.path)
			if testCase.expectError {
				if err == nil {
					t.Fatalf("Expected error, got: '%v'", err)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: '%v'", err)
				}
			}

			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", testCase.expectValue) {
				t.Fatalf("Expected to get '%v', got '%v'", testCase.expectValue, got)
			}
		})
	}
}
