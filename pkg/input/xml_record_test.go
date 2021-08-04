package input

import (
	"bytes"
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	parserPackage "rodb.io/pkg/parser"
	"testing"
)

func TestXmlRecordAll(t *testing.T) {
	booleanParser := parserPackage.NewBoolean(&parserPackage.BooleanConfig{
		TrueValues:  []string{"true"},
		FalseValues: []string{"false"},
	})
	floatParser := parserPackage.NewFloat(&parserPackage.FloatConfig{
		DecimalSeparator: ".",
	})
	integerParser := parserPackage.NewInteger(&parserPackage.IntegerConfig{})
	jsonParser := parserPackage.NewJson(&parserPackage.JsonConfig{})
	stringParser, err := parserPackage.NewString(&parserPackage.StringConfig{})
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	node, err := xmlquery.Parse(bytes.NewReader([]byte(`
		<root>
			<integer>42</integer>
			<float>3.14</float>
			<boolean>true</boolean>
			<json>{"json_prop": "json val"}</json>
			<array>
				<item>arr a</item>
				<item>arr b</item>
			</array>
			<object>
				<prop_a>obj val a</prop_a>
				<prop_b>obj val b</prop_b>
			</object>
			<arrayOfObjects>
				<object>
					<prop>array of obj val a</prop>
				</object>
				<object>
					<prop>array of obj val b</prop>
				</object>
			</arrayOfObjects>
		</root>
	`)))
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	config := &XmlConfig{
		Properties: []*XmlPropertyConfig{
			{
				Type:          XmlInputPropertyTypePrimitive,
				Name:          "integer",
				Parser:        "integer",
				CompiledXPath: xpath.MustCompile("number(/root/integer)"),
			},
			{
				Type:          XmlInputPropertyTypePrimitive,
				Name:          "float",
				Parser:        "float",
				CompiledXPath: xpath.MustCompile("number(/root/float)"),
			},
			{
				Type:          XmlInputPropertyTypePrimitive,
				Name:          "boolean",
				Parser:        "boolean",
				CompiledXPath: xpath.MustCompile("boolean(/root/boolean[text()='true'])"),
			},
			{
				Type:          XmlInputPropertyTypePrimitive,
				Name:          "json",
				Parser:        "json",
				CompiledXPath: xpath.MustCompile("string(/root/json)"),
			},
			{
				Type:          XmlInputPropertyTypeArray,
				Name:          "array",
				CompiledXPath: xpath.MustCompile("/root/array/item"),
				Items: &XmlPropertyConfig{
					Type:          XmlInputPropertyTypePrimitive,
					CompiledXPath: xpath.MustCompile("string(/)"),
					Parser:        "string",
				},
			},
			{
				Type:          XmlInputPropertyTypeObject,
				Name:          "object",
				CompiledXPath: xpath.MustCompile("/root/object"),
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						CompiledXPath: xpath.MustCompile("string(/prop_a)"),
						Name:          "prop_a",
						Parser:        "string",
					},
					{
						Type:          XmlInputPropertyTypePrimitive,
						CompiledXPath: xpath.MustCompile("string(/prop_b)"),
						Name:          "prop_b",
						Parser:        "string",
					},
				},
			},
			{
				Type:          XmlInputPropertyTypeArray,
				Name:          "arrayOfObjects",
				CompiledXPath: xpath.MustCompile("/root/arrayOfObjects/object"),
				Items: &XmlPropertyConfig{
					Type:          XmlInputPropertyTypeObject,
					CompiledXPath: xpath.MustCompile("/"),
					Properties: []*XmlPropertyConfig{
						{
							Type:          XmlInputPropertyTypePrimitive,
							CompiledXPath: xpath.MustCompile("string(/prop)"),
							Name:          "prop",
							Parser:        "string",
						},
					},
				},
			},
			{
				Type:          XmlInputPropertyTypePrimitive,
				Name:          "nodeAsString",
				Parser:        "string",
				CompiledXPath: xpath.MustCompile("/root/integer"),
			},
		},
	}

	parsers := parserPackage.List{
		"string":  stringParser,
		"json":    jsonParser,
		"boolean": booleanParser,
		"float":   floatParser,
		"integer": integerParser,
	}

	record, err := NewXmlRecord(config, node, parsers, 0)
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	result, err := record.All()
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	if expect, got := int64(42), result["integer"].(int64); got != expect {
		t.Fatalf("Expected to get '%v', got '%v'", expect, got)
	}
	if expect, got := 3.14, result["float"].(float64); got != expect {
		t.Fatalf("Expected to get '%v', got '%v'", expect, got)
	}
	if expect, got := true, result["boolean"].(bool); got != expect {
		t.Fatalf("Expected to get '%v', got '%v'", expect, got)
	}

	json := result["json"].(map[string]interface{})
	if expect, got := "json val", json["json_prop"].(string); got != expect {
		t.Fatalf("Expected to get '%v', got '%v'", expect, got)
	}

	array := result["array"].([]interface{})
	if expect, got := 2, len(array); got != expect {
		t.Fatalf("Expected a length of '%v', got '%v'", expect, got)
	}
	if expect, got := "arr a", array[0].(string); got != expect {
		t.Fatalf("Expected to get '%v', got '%v'", expect, got)
	}
	if expect, got := "arr b", array[1].(string); got != expect {
		t.Fatalf("Expected to get '%v', got '%v'", expect, got)
	}

	object := result["object"].(map[string]interface{})
	if expect, got := "obj val a", object["prop_a"].(string); got != expect {
		t.Fatalf("Expected to get '%v', got '%v'", expect, got)
	}
	if expect, got := "obj val b", object["prop_b"].(string); got != expect {
		t.Fatalf("Expected to get '%v', got '%v'", expect, got)
	}

	arrayOfObjects := result["arrayOfObjects"].([]interface{})
	if expect, got := 2, len(arrayOfObjects); got != expect {
		t.Fatalf("Expected a length of '%v', got '%v'", expect, got)
	}
	arrayOfObjects0 := arrayOfObjects[0].(map[string]interface{})
	if expect, got := "array of obj val a", arrayOfObjects0["prop"].(string); got != expect {
		t.Fatalf("Expected to get '%v', got '%v'", expect, got)
	}
	arrayOfObjects1 := arrayOfObjects[1].(map[string]interface{})
	if expect, got := "array of obj val b", arrayOfObjects1["prop"].(string); got != expect {
		t.Fatalf("Expected to get '%v', got '%v'", expect, got)
	}

	if expect, got := "<integer>42</integer>", result["nodeAsString"].(string); got != expect {
		t.Fatalf("Expected to get '%v', got '%v'", expect, got)
	}
}

func TestXmlRecordGet(t *testing.T) {
	booleanParser := parserPackage.NewBoolean(&parserPackage.BooleanConfig{
		TrueValues:  []string{"true"},
		FalseValues: []string{"false"},
	})
	floatParser := parserPackage.NewFloat(&parserPackage.FloatConfig{
		DecimalSeparator: ".",
	})
	integerParser := parserPackage.NewInteger(&parserPackage.IntegerConfig{})
	jsonParser := parserPackage.NewJson(&parserPackage.JsonConfig{})
	mockParser := parserPackage.NewMock()
	stringParser, err := parserPackage.NewString(&parserPackage.StringConfig{})
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	colName := "col_a"
	createRecord := func(
		data []byte,
		config *XmlConfig,
		parsers parserPackage.List,
	) *XmlRecord {
		node, err := xmlquery.Parse(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		record, err := NewXmlRecord(config, node, parsers, 0)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		return record
	}

	t.Run("string xpath", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>Hello World!</a></root>"),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("string(/root/a)"),
					},
				},
			},
			parserPackage.List{"parser": mockParser},
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		if expect := "Hello World!"; got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("string xpath on integer property", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>42</a></root>"),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("string(/root/a)"),
					},
				},
			},
			parserPackage.List{"parser": integerParser},
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		if expect := int64(42); got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("string xpath on float property", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>42.1</a></root>"),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("string(/root/a)"),
					},
				},
			},
			parserPackage.List{"parser": floatParser},
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		if expect := 42.1; got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("string xpath on boolean property", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>true</a></root>"),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("string(/root/a)"),
					},
				},
			},
			parserPackage.List{"parser": booleanParser},
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		if expect := true; got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("number xpath on integer property", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>42</a></root>"),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("number(/root/a)"),
					},
				},
			},
			parserPackage.List{"parser": integerParser},
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		if expect := int64(42); got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("number xpath on float property", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>42.1</a></root>"),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("number(/root/a)"),
					},
				},
			},
			parserPackage.List{"parser": floatParser},
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		if expect := 42.1; got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("number xpath on string property", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>42</a></root>"),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("number(/root/a)"),
					},
				},
			},
			parserPackage.List{"parser": mockParser},
		)

		_, err := record.Get(colName)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})
	t.Run("boolean xpath", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>a</a></root>"),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("boolean(/root/a[text()='a'])"),
					},
				},
			},
			parserPackage.List{"parser": booleanParser},
		)

		got, err := record.Get(colName)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		if expect := true; got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("boolean xpath on integer property", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>a</a></root>"),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("/root/a[text()='a']"),
					},
				},
			},
			parserPackage.List{"parser": integerParser},
		)

		_, err := record.Get(colName)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})
	t.Run("node xpath", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>a</a></root>"),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("/root/a"),
					},
				},
			},
			parserPackage.List{"parser": mockParser},
		)

		_, err := record.Get(colName)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})
	t.Run("missing property", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>a</a></root>"),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						Name:          colName,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("/root/a"),
					},
				},
			},
			parserPackage.List{"parser": mockParser},
		)

		_, err := record.Get("not_" + colName)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})
	t.Run("integer in array", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>1</a>
				<a>42</a>
				<a>3</a>
			</root>`),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypeArray,
						CompiledXPath: xpath.MustCompile("/root/a"),
						Name:          colName,
						Items: &XmlPropertyConfig{
							Type:          XmlInputPropertyTypePrimitive,
							CompiledXPath: xpath.MustCompile("string(/)"),
							Parser:        "parser",
						},
					},
				},
			},
			parserPackage.List{"parser": integerParser},
		)

		got, err := record.Get(colName + ".1")
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		if expect := int64(42); got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("integer in object", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<sub>
					<a>42</a>
				</sub>
			</root>`),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypeObject,
						CompiledXPath: xpath.MustCompile("/root/sub"),
						Name:          colName,
						Properties: []*XmlPropertyConfig{
							{
								Type:          XmlInputPropertyTypePrimitive,
								CompiledXPath: xpath.MustCompile("string(/a)"),
								Name:          "prop",
								Parser:        "parser",
							},
						},
					},
				},
			},
			parserPackage.List{"parser": integerParser},
		)

		got, err := record.Get(colName + ".prop")
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		if expect := int64(42); got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("object type, but multiple results", func(t *testing.T) {
		record := createRecord(
			[]byte("<root><a>a1</a><a>a2</a></root>"),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypeObject,
						CompiledXPath: xpath.MustCompile("/root/a"),
						Name:          colName,
						Properties: []*XmlPropertyConfig{
							{
								Type:          XmlInputPropertyTypePrimitive,
								CompiledXPath: xpath.MustCompile("string(/a)"),
								Name:          "prop",
								Parser:        "parser",
							},
						},
					},
				},
			},
			parserPackage.List{"parser": mockParser},
		)

		_, err := record.Get(colName)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})
	t.Run("array value from node", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>1</a>
				<a>42</a>
			</root>`),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypeArray,
						CompiledXPath: xpath.MustCompile("/root/a"),
						Name:          colName,
						Items: &XmlPropertyConfig{
							Type:          XmlInputPropertyTypePrimitive,
							CompiledXPath: xpath.MustCompile("string(/)"),
							Parser:        "parser",
						},
					},
				},
			},
			parserPackage.List{"parser": integerParser},
		)

		result, err := record.Get(colName)
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		arrayResult := result.([]interface{})
		if expect := 2; len(arrayResult) != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, arrayResult)
		}
		if expect := int64(1); arrayResult[0] != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, arrayResult[0])
		}
		if expect := int64(42); arrayResult[1] != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, arrayResult[1])
		}
	})
	t.Run("object value from node", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<sub>
					<a>42</a>
				</sub>
			</root>`),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypeObject,
						CompiledXPath: xpath.MustCompile("/root/sub"),
						Name:          colName,
						Properties: []*XmlPropertyConfig{
							{
								Type:          XmlInputPropertyTypePrimitive,
								CompiledXPath: xpath.MustCompile("string(/a)"),
								Name:          "prop",
								Parser:        "parser",
							},
						},
					},
				},
			},
			parserPackage.List{"parser": integerParser},
		)

		result, err := record.Get(colName)
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		objectResult := result.(map[string]interface{})
		if expect := int64(42); objectResult["prop"] != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, objectResult["prop"])
		}
	})
	t.Run("value inside array from parse", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>[1, 42]</a>
			</root>`),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("string(/root/a)"),
						Name:          colName,
					},
				},
			},
			parserPackage.List{"parser": jsonParser},
		)

		result, err := record.Get(colName + ".1")
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		if expect := float64(42); result != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, result)
		}
	})
	t.Run("value inside object from parse", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>{"prop": 42}</a>
			</root>`),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						CompiledXPath: xpath.MustCompile("string(/root/a)"),
						Name:          colName,
						Parser:        "parser",
					},
				},
			},
			parserPackage.List{"parser": jsonParser},
		)

		result, err := record.Get(colName + ".prop")
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		if expect := float64(42); result != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, result)
		}
	})
	t.Run("array value from parse", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>[1, 42]</a>
			</root>`),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						Parser:        "parser",
						CompiledXPath: xpath.MustCompile("string(/root/a)"),
						Name:          colName,
					},
				},
			},
			parserPackage.List{"parser": jsonParser},
		)

		result, err := record.Get(colName)
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		arrayResult := result.([]interface{})
		if expect := 2; len(arrayResult) != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, arrayResult)
		}
		if expect := float64(1); arrayResult[0] != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, arrayResult[0])
		}
		if expect := float64(42); arrayResult[1] != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, arrayResult[1])
		}
	})
	t.Run("object value from parse", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>{"prop": 42}</a>
			</root>`),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						CompiledXPath: xpath.MustCompile("string(/root/a)"),
						Name:          colName,
						Parser:        "parser",
					},
				},
			},
			parserPackage.List{"parser": jsonParser},
		)

		result, err := record.Get(colName)
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		objectResult := result.(map[string]interface{})
		if expect := float64(42); objectResult["prop"] != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, objectResult["prop"])
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
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypeObject,
						CompiledXPath: xpath.MustCompile("/root/sub"),
						Name:          colName,
						Properties: []*XmlPropertyConfig{
							{
								Type:          XmlInputPropertyTypeArray,
								CompiledXPath: xpath.MustCompile("/a"),
								Name:          "prop",
								Items: &XmlPropertyConfig{
									Type:          XmlInputPropertyTypePrimitive,
									CompiledXPath: xpath.MustCompile("string(/)"),
									Parser:        "parser",
								},
							},
						},
					},
				},
			},
			parserPackage.List{"parser": integerParser},
		)

		got, err := record.Get(colName + ".prop.1")
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		if expect := int64(42); got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
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
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypeArray,
						CompiledXPath: xpath.MustCompile("/root/sub"),
						Name:          colName,
						Items: &XmlPropertyConfig{
							Type:          XmlInputPropertyTypeObject,
							CompiledXPath: xpath.MustCompile("/a"),
							Properties: []*XmlPropertyConfig{
								{
									Type:          XmlInputPropertyTypePrimitive,
									CompiledXPath: xpath.MustCompile("string(/)"),
									Name:          "prop",
									Parser:        "parser",
								},
							},
						},
					},
				},
			},
			parserPackage.List{"parser": integerParser},
		)

		got, err := record.Get(colName + ".1.prop")
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		if expect := int64(42); got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("nodes as string", func(t *testing.T) {
		record := createRecord(
			[]byte(`<root>
				<a>a1</a>
				<a>a2</a>
			</root>`),
			&XmlConfig{
				Properties: []*XmlPropertyConfig{
					{
						Type:          XmlInputPropertyTypePrimitive,
						CompiledXPath: xpath.MustCompile("/root/a"),
						Name:          colName,
						Parser:        "parser",
					},
				},
			},
			parserPackage.List{"parser": stringParser},
		)

		result, err := record.Get(colName)
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		if expect, got := "<a>a1</a><a>a2</a>", result.(string); got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
}
