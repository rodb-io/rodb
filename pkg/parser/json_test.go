package parser

import (
	"testing"
)

func TestJsonParse(t *testing.T) {
	t.Run("object", func(t *testing.T) {
		config := &JsonConfig{}
		jsonParser := NewJson(config)

		data, err := jsonParser.Parse(`{
			"foo": {
				"bar": [
					"baz"
				]
			}
		}`)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		dataMap, isMap := data.(map[string]interface{})
		if !isMap {
			t.Fatalf("Expected data to be a map, got '%#v'", data)
		}

		foo, fooMapExists := dataMap["foo"]
		if !fooMapExists {
			t.Fatalf("Expected map to have a 'foo' key, got '%#v'", data)
		}

		fooMap, isMap := foo.(map[string]interface{})
		if !isMap {
			t.Fatalf("Expected property to be a map, got '%#v'", foo)
		}

		bar, barArrayExists := fooMap["bar"]
		if !barArrayExists {
			t.Fatalf("Expected map to have a 'bar' key, got '%#v'", fooMap)
		}

		barArray, isArray := bar.([]interface{})
		if !isArray {
			t.Fatalf("Expected property to be an array, got '%#v'", bar)
		}

		if len(barArray) != 1 {
			t.Fatalf("Expected array to have 1 value, got '%#v'", bar)
		}

		baz, isString := barArray[0].(string)
		if !isString {
			t.Fatalf("Expected property to be a string, got '%#v'", baz)
		}

		if baz != "baz" {
			t.Fatalf("Expected array value to be 'baz', got '%#v'", baz)
		}
	})
	t.Run("array", func(t *testing.T) {
		config := &JsonConfig{}
		jsonParser := NewJson(config)

		data, err := jsonParser.Parse(`["baz"]`)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		dataArray, isArray := data.([]interface{})
		if !isArray {
			t.Fatalf("Expected property to be an array, got '%#v'", data)
		}

		if len(dataArray) != 1 {
			t.Fatalf("Expected array to have 1 value, got '%#v'", data)
		}

		dataString, isString := dataArray[0].(string)
		if !isString {
			t.Fatalf("Expected property to be a string, got '%#v'", dataArray)
		}

		if dataString != "baz" {
			t.Fatalf("Expected array value to be 'baz', got '%#v'", dataString)
		}
	})
	t.Run("primitive", func(t *testing.T) {
		config := &JsonConfig{}
		jsonParser := NewJson(config)

		data, err := jsonParser.Parse(`42`)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		if data != float64(42) {
			t.Fatalf("Expected data to be 42, got '%#v'", data)
		}
	})
}
