package record

import (
	"fmt"
	"rods/pkg/config"
	"rods/pkg/parser"
	"testing"
)

func TestCsvGet(t *testing.T) {
	var testConfig *config.CsvInput = &config.CsvInput{
		Columns: []*config.CsvInputColumn{
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
		parser.NewMock(),
	}

	t.Run("normal", func(t *testing.T) {
		record := NewCsv(testConfig, parsers, []string{"string_a", "string_b"}, 0)

		got, err := record.Get("col_a")
		expect := "string_a"
		if fmt.Sprintf("%v", got) != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}

		got, err = record.Get("col_b")
		expect = "string_b"
		if fmt.Sprintf("%v", got) != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("error if col does not exist", func(t *testing.T) {
		record := NewCsv(testConfig, parsers, []string{}, 0)
		got, err := record.Get("col_0")
		if err == nil {
			t.Errorf("Expected error, got '%v'", got)
		}
	})
	t.Run("col not found", func(t *testing.T) {
		record := NewCsv(testConfig, parsers, []string{}, 0)
		got, err := record.Get("col_a")
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
		if got != nil {
			t.Errorf("Expected nil, got '%v'", got)
		}
	})
}
