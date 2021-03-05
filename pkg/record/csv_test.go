package record

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/parser"
	"testing"
)

func TestCsvGet(t *testing.T) {
	var testConfig *config.CsvInput = &config.CsvInput{
		Columns: []config.CsvInputColumn{
			{
				Name:   "col_a",
				Parser: "string",
			}, {
				Name:   "col_b",
				Parser: "integer",
			}, {
				Name:   "col_c",
				Parser: "float",
			}, {
				Name:   "col_d",
				Parser: "boolean",
			},
		},
		ColumnIndexByName: map[string]int{
			"col_a": 0,
			"col_b": 1,
			"col_c": 2,
			"col_d": 3,
		},
	}

	parsers := []parser.Parser{
		parser.NewString(&config.StringParser{}, logrus.StandardLogger()),
		parser.NewInteger(&config.IntegerParser{
			IgnoreCharacters: ",$ ",
		}, logrus.StandardLogger()),
		parser.NewFloat(&config.FloatParser{
			IgnoreCharacters: ",€",
			DecimalSeparator: ".",
		}, logrus.StandardLogger()),
		parser.NewBoolean(&config.BooleanParser{
			TrueValues:  []string{"yes"},
			FalseValues: []string{"no"},
		}, logrus.StandardLogger()),
	}

	t.Run("string", func(t *testing.T) {
		record := NewCsv(testConfig, parsers, []string{"string"}, 0)
		got, err := record.Get("col_a")
		expect := "string"
		if fmt.Sprintf("%v", got) != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("integer", func(t *testing.T) {
		record := NewCsv(testConfig, parsers, []string{"", "$ 123,456"}, 0)
		got, err := record.Get("col_b")
		expect := "123456"
		if fmt.Sprintf("%v", got) != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("float", func(t *testing.T) {
		record := NewCsv(testConfig, parsers, []string{"", "", "1,234.56€"}, 0)
		got, err := record.Get("col_c")
		expect := "1234.56"
		if fmt.Sprintf("%v", got) != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("true", func(t *testing.T) {
		record := NewCsv(testConfig, parsers, []string{"", "", "", "yes"}, 0)
		expect := "true"
		got, err := record.Get("col_d")
		if fmt.Sprintf("%v", got) != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("false", func(t *testing.T) {
		record := NewCsv(testConfig, parsers, []string{"", "", "", "no"}, 0)
		expect := "false"
		got, err := record.Get("col_d")
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
