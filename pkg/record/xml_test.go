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
	jsonParser := parserModule.NewJson(&config.JsonParser{})
	mockParser := parserModule.NewMock()

	colName := "col_a"
	createRecord := func(
		data []byte,
		config *config.XmlInput,
		parsers parserModule.List,
	) *Xml {
		node, err := xmlquery.Parse(bytes.NewReader(data))
		if err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}

		record, err := NewXml(config, node, parsers, 0)
		if err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}

		return record
	}

	t.Run("string xpath", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>Hello World!</a></root>"),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("string(/root/a)"),
					},
				},
			},
			parserModule.List{"parser": mockParser},
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
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("string(/root/a)"),
					},
				},
			},
			parserModule.List{"parser": integerParser},
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
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("string(/root/a)"),
					},
				},
			},
			parserModule.List{"parser": floatParser},
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
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("string(/root/a)"),
					},
				},
			},
			parserModule.List{"parser": booleanParser},
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
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("number(/root/a)"),
					},
				},
			},
			parserModule.List{"parser": integerParser},
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
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("number(/root/a)"),
					},
				},
			},
			parserModule.List{"parser": floatParser},
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
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("number(/root/a)"),
					},
				},
			},
			parserModule.List{"parser": mockParser},
		)

		_, err := record.Get(colName)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
	t.Run("boolean xpath", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>a</a></root>"),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("boolean(/root/a[text()='a'])"),
					},
				},
			},
			parserModule.List{"parser": booleanParser},
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
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("/root/a[text()='a']"),
					},
				},
			},
			parserModule.List{"parser": integerParser},
		)

		_, err := record.Get(colName)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
	t.Run("node xpath", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>a</a></root>"),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("/root/a"),
					},
				},
			},
			parserModule.List{"parser": mockParser},
		)

		_, err := record.Get(colName)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
	t.Run("missing property", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>a</a></root>"),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("/root/a"),
					},
				},
			},
			parserModule.List{"parser": mockParser},
		)

		_, err := record.Get("not_" + colName)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
	t.Run("integer in array", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>1</a>
				<a>42</a>
				<a>3</a>
			</root>`),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypeArray,
						CompiledXPath: xpath.MustCompile("/root/a"),
						Name:          colName,
						Items: &config.XmlInputProperty{
							Type:          config.XmlInputPropertyTypePrimitive,
							CompiledXPath: xpath.MustCompile("string(/a)"),
							Parser:        "parser",
						},
					},
				},
			},
			parserModule.List{"parser": integerParser},
		)

		got, err := record.Get(colName + ".1")
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		if expect := 42; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("integer in object", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<sub>
					<a>42</a>
				</sub>
			</root>`),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypeObject,
						CompiledXPath: xpath.MustCompile("/root/sub"),
						Name:          colName,
						Properties: []*config.XmlInputProperty{
							{
								Type:          config.XmlInputPropertyTypePrimitive,
								CompiledXPath: xpath.MustCompile("string(/)"),
								Name:          "prop",
								Parser:        "parser",
							},
						},
					},
				},
			},
			parserModule.List{"parser": integerParser},
		)

		got, err := record.Get(colName + ".prop")
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		if expect := 42; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("array instead of object", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>a1</a><a>a2</a></root>"),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypeArray,
						CompiledXPath: xpath.MustCompile("/root/a"),
						Name:          colName,
						Items: &config.XmlInputProperty{
							Type:          config.XmlInputPropertyTypePrimitive,
							CompiledXPath: xpath.MustCompile("string(/)"),
							Parser:        "parser",
						},
					},
				},
			},
			parserModule.List{"parser": mockParser},
		)

		_, err := record.Get(colName)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
	t.Run("array value from node", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>1</a>
				<a>42</a>
			</root>`),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypeArray,
						CompiledXPath: xpath.MustCompile("/root/a"),
						Name:          colName,
						Items: &config.XmlInputProperty{
							Type:          config.XmlInputPropertyTypePrimitive,
							CompiledXPath: xpath.MustCompile("string(/)"),
							Parser:        "parser",
						},
					},
				},
			},
			parserModule.List{"parser": integerParser},
		)

		result, err := record.Get(colName)
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		arrayResult := result.([]interface{})
		if expect := 2; len(arrayResult) != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, arrayResult)
		}
		if expect := 1; arrayResult[0] != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, arrayResult[0])
		}
		if expect := 42; arrayResult[1] != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, arrayResult[1])
		}
	})
	t.Run("object value from node", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<sub>
					<a>42</a>
				</sub>
			</root>`),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypeObject,
						CompiledXPath: xpath.MustCompile("/root/sub"),
						Name:          colName,
						Properties: []*config.XmlInputProperty{
							{
								Type:          config.XmlInputPropertyTypePrimitive,
								CompiledXPath: xpath.MustCompile("string(/)"),
								Name:          "prop",
								Parser:        "parser",
							},
						},
					},
				},
			},
			parserModule.List{"parser": integerParser},
		)

		result, err := record.Get(colName)
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		objectResult := result.(map[string]interface{})
		if expect := 42; objectResult["prop"] != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, objectResult["prop"])
		}
	})
	t.Run("value inside array from parse", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>[1, 42]</a>
			</root>`),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("/root/a"),
						Name:          colName,
					},
				},
			},
			parserModule.List{"parser": jsonParser},
		)

		result, err := record.Get(colName + ".1")
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		if expect := 42; result != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, result)
		}
	})
	t.Run("value inside object from parse", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>{"prop": 42}</a>
			</root>`),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						CompiledXPath: xpath.MustCompile("/root/sub"),
						Name:          colName,
						Parser:        "parser",
					},
				},
			},
			parserModule.List{"parser": jsonParser},
		)

		result, err := record.Get(colName + ".prop")
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		objectResult := result.(map[string]interface{})
		if expect := 42; objectResult["prop"] != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, objectResult["prop"])
		}
	})
	t.Run("array value from parse", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>[1, 42]</a>
			</root>`),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("/root/a"),
						Name:          colName,
					},
				},
			},
			parserModule.List{"parser": jsonParser},
		)

		result, err := record.Get(colName)
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		arrayResult := result.([]interface{})
		if expect := 2; len(arrayResult) != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, arrayResult)
		}
		if expect := 1; arrayResult[0] != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, arrayResult[0])
		}
		if expect := 42; arrayResult[1] != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, arrayResult[1])
		}
	})
	t.Run("object value from parse", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>{"prop": 42}</a>
			</root>`),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						CompiledXPath: xpath.MustCompile("/root/sub"),
						Name:          colName,
						Parser:        "parser",
					},
				},
			},
			parserModule.List{"parser": jsonParser},
		)

		result, err := record.Get(colName)
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		objectResult := result.(map[string]interface{})
		if expect := 42; objectResult["prop"] != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, objectResult["prop"])
		}
	})
	t.Run("primitive xpath with non-primitive type", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>1</a>
			</root>`),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypeArray,
						CompiledXPath: xpath.MustCompile("string(/root/a)"),
						Name:          colName,
						Items: &config.XmlInputProperty{
							Type:          config.XmlInputPropertyTypePrimitive,
							CompiledXPath: xpath.MustCompile("string(/)"),
							Parser:        "parser",
						},
					},
				},
			},
			parserModule.List{"parser": integerParser},
		)

		_, err := record.Get(colName)
		if err == nil {
			t.Errorf("Expected error, got %v", err)
		}
	})
	t.Run("integer in array in object", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<sub>
					<a>1</a>
					<a>42</a>
					<a>3</a>
				</sub>
			</root>`),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypeObject,
						CompiledXPath: xpath.MustCompile("/root/sub"),
						Name:          colName,
						Properties: []*config.XmlInputProperty{
							{
								Type:          config.XmlInputPropertyTypeArray,
								CompiledXPath: xpath.MustCompile("/sub/a"),
								Name:          "prop",
								Items: &config.XmlInputProperty{
									Type:          config.XmlInputPropertyTypePrimitive,
									CompiledXPath: xpath.MustCompile("string(/)"),
									Parser:        "parser",
								},
							},
						},
					},
				},
			},
			parserModule.List{"parser": integerParser},
		)

		got, err := record.Get(colName + ".prop.1")
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		if expect := 42; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("integer in object in array", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<sub>
					<a>1</a>
				</sub>
				<sub>
					<a>42</a>
				</sub>
			</root>`),
			&config.XmlInput{
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypeArray,
						CompiledXPath: xpath.MustCompile("/root/sub"),
						Name:          colName,
						Items: &config.XmlInputProperty{
							Type:          config.XmlInputPropertyTypeObject,
							CompiledXPath: xpath.MustCompile("/sub/a"),
							Properties: []*config.XmlInputProperty{
								{
									Type:          config.XmlInputPropertyTypePrimitive,
									CompiledXPath: xpath.MustCompile("string(/)"),
									Name:          "prop",
									Parser:        "parser",
								},
							},
						},
					},
				},
			},
			parserModule.List{"parser": integerParser},
		)

		got, err := record.Get(colName + ".1.prop")
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		if expect := 42; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
}
