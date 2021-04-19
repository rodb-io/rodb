package record

import (
	"bytes"
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"rodb.io/pkg/config"
	parserModule "rodb.io/pkg/parser"
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
		data []byte,
		xPath string,
		parser parserModule.Parser,
	) *Xml {
		var testConfig *config.XmlInput = &config.XmlInput{
			Properties: []*config.XmlInputProperty{
				{
					Name:          colName,
					Parser:        "parser",
					CompiledXPath: xpath.MustCompile(xPath),
				},
			},
		}

		node, err := xmlquery.Parse(bytes.NewReader(data))
		if err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}

		parsers := parserModule.List{
			"parser": parser,
		}

		record, err := NewXml(testConfig, node, parsers, 0)
		if err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}

		return record
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
	t.Run("string xpath on integer property", func(t *testing.T) {
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
	t.Run("string xpath on float property", func(t *testing.T) {
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
	t.Run("string xpath on boolean property", func(t *testing.T) {
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
	t.Run("number xpath on integer property", func(t *testing.T) {
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
	t.Run("number xpath on float property", func(t *testing.T) {
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
	t.Run("number xpath on string property", func(t *testing.T) {
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
	t.Run("boolean xpath on integer property", func(t *testing.T) {
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
	t.Run("test", func(t *testing.T) {
		// TODO
		// TODO test getting non-existing node (does it work? return unsupported nil?
		record := createRecord(
			[]byte(`<root>
				<a>a1</a>
				<a>a2</a>
				<a>a3</a>
				<a>a4</a>
			</root>`),
			"/root/a",
			mockParser,
		)

		_, err := record.Get(colName)
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
	})
}
