package record

import (
	"testing"
	"rods/pkg/config"
	"rods/pkg/utils"
)

var testConfig *config.CsvInputConfig = &config.CsvInputConfig{
	Columns: []config.CsvInputColumnConfig {
		{
			Name: "col_a",
			Type: "string",
		}, {
			Name: "col_b",
			Type: "integer",
			IgnoreCharacters: ",$ ",
		}, {
			Name: "col_c",
			Type: "float",
			IgnoreCharacters: ",€",
			DecimalSeparator: ".",
		}, {
			Name: "col_d",
			Type: "boolean",
			TrueValues: []string{ "yes" },
			FalseValues: []string{ "no" },
		},
	},
	ColumnIndexByName: map[string]int{
		"col_a": 0,
		"col_b": 1,
		"col_c": 2,
		"col_d": 3,
	},
}

func TestGetString(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		record := NewCsvRecord(testConfig, []string { "string" })
		expect := utils.PString("string")
		got, err := record.GetString("col_a")
		if *got != *expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("error if col does not exist", func(t *testing.T) {
		record := NewCsvRecord(testConfig, []string {})
		got, err := record.GetString("col_0")
		if err == nil {
			t.Errorf("Expected error, got '%v'", got)
		}
	})
	t.Run("col not found", func(t *testing.T) {
		record := NewCsvRecord(testConfig, []string {})
		got, err := record.GetString("col_a")
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
		if got != nil {
			t.Errorf("Expected nil, got '%v'", got)
		}
	})
}

func TestGetInteger(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		record := NewCsvRecord(testConfig, []string { "", "$ 123,456" })
		expect := utils.PInt(123456)
		got, err := record.GetInteger("col_b")
		if *got != *expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("error if col does not exist", func(t *testing.T) {
		record := NewCsvRecord(testConfig, []string { "", "123" })
		got, err := record.GetInteger("col_0")
		if err == nil {
			t.Errorf("Expected error, got '%v'", got)
		}
	})
	t.Run("col not found", func(t *testing.T) {
		record := NewCsvRecord(testConfig, []string { "" })
		got, err := record.GetInteger("col_b")
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
		if got != nil {
			t.Errorf("Expected nil, got '%v'", got)
		}
	})
}

func TestGetFloat(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		record := NewCsvRecord(testConfig, []string { "", "", "1,234.56€" })
		expect := utils.PFloat(1234.56)
		got, err := record.GetFloat("col_c")
		if *got != *expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("error if col does not exist", func(t *testing.T) {
		record := NewCsvRecord(testConfig, []string { "", "", "1,234.56€" })
		got, err := record.GetFloat("col_0")
		if err == nil {
			t.Errorf("Expected error, got '%v'", got)
		}
	})
	t.Run("col not found", func(t *testing.T) {
		record := NewCsvRecord(testConfig, []string { "", "" })
		got, err := record.GetFloat("col_c")
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
		if got != nil {
			t.Errorf("Expected nil, got '%v'", got)
		}
	})
}

func TestGetBoolean(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		record := NewCsvRecord(testConfig, []string { "", "", "", "yes" })
		expect := utils.PBool(true)
		got, err := record.GetBoolean("col_d")
		if *got != *expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("false", func(t *testing.T) {
		record := NewCsvRecord(testConfig, []string { "", "", "", "no" })
		expect := utils.PBool(false)
		got, err := record.GetBoolean("col_d")
		if *got != *expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("error if col does not exist", func(t *testing.T) {
		record := NewCsvRecord(testConfig, []string { "", "", "", "yes" })
		got, err := record.GetBoolean("col_0")
		if err == nil {
			t.Errorf("Expected error, got '%v'", got)
		}
	})
	t.Run("col not found", func(t *testing.T) {
		record := NewCsvRecord(testConfig, []string { "", "", "" })
		got, err := record.GetBoolean("col_d")
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
		if got != nil {
			t.Errorf("Expected nil, got '%v'", got)
		}
	})
}
