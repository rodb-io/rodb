package record

import (
	"github.com/antchfx/xpath"
	"rods/pkg/config"
	parserModule "rods/pkg/parser"
	"testing"
)

func TestXmlGet(t *testing.T) {
	booleanParser := parserModule.NewBoolean(&config.BooleanParser{
		TrueValues:  []string{"true"},
		FalseValues: []string{"false"},
	})
	floatParser := parserModule.NewFloat(&config.FloatParser{
		DecimalSeparator: ".",
	})
	integerParser := parserModule.NewInteger(&config.IntegerParser{})
	mockParser := parserModule.NewMock()

	colName := "col_a"
	createRecord := func(
		xml []byte,
		xPath string,
		parser parserModule.Parser,
	) *Xml {
		var testConfig *config.XmlInput = &config.XmlInput{
			Columns: []*config.XmlInputColumn{
				{
					Name:          colName,
					CompiledXPath: xpath.MustCompile(xPath),
				},
			},
			ColumnIndexByName: map[string]int{
				colName: 0,
			},
		}

		return NewXml(testConfig, []parserModule.Parser{parser}, xml, 0)
	}

	t.Run("string xpath", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>Hello World!</a></root>"),
			"string(/root/a)",
			mockParser,
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}
		if expect := "Hello World!"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("string xpath on integer column", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>42</a></root>"),
			"string(/root/a)",
			integerParser,
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}
		if expect := 42; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("string xpath on float column", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>42.1</a></root>"),
			"string(/root/a)",
			floatParser,
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}
		if expect := 42.1; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("string xpath on boolean column", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>true</a></root>"),
			"string(/root/a)",
			booleanParser,
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}
		if expect := true; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("number xpath on integer column", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>42</a></root>"),
			"number(/root/a)",
			integerParser,
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}
		if expect := 42; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("number xpath on float column", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>42.1</a></root>"),
			"number(/root/a)",
			floatParser,
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}
		if expect := 42.1; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("number xpath on string column", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>42</a></root>"),
			"number(/root/a)",
			mockParser,
		)

		_, err := record.Get(colName)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
	t.Run("boolean xpath", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>a</a></root>"),
			"boolean(/root/a[text()='a'])",
			booleanParser,
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}
		if expect := true; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("boolean xpath on integer column", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>a</a></root>"),
			"/root/a[text()='a']",
			integerParser,
		)

		_, err := record.Get(colName)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
	t.Run("node xpath", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>a</a></root>"),
			"/root/a",
			mockParser,
		)

		_, err := record.Get(colName)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}
