package parser

import (
	"rodb.io/pkg/config"
	"testing"
)

func TestJsonParse(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		config := &config.JsonParser{}
		jsonParser := NewJson(config)

		data, err := jsonParser.Parse(`{
			"foo": {
				"bar": [
					"baz"
				]
			}
		}`)
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}

		dataMap, isMap := data.(map[string]interface{})
		if !isMap {
			t.Errorf("Expected data to be a map, got '%#v'", data)
		}

		foo, fooMapExists := dataMap["foo"]
		if !fooMapExists {
			t.Errorf("Expected map to have a 'foo' key, got '%#v'", data)
		}

		fooMap, isMap := foo.(map[string]interface{})
		if !isMap {
			t.Errorf("Expected property to be a map, got '%#v'", foo)
		}

		bar, barArrayExists := fooMap["bar"]
		if !barArrayExists {
			t.Errorf("Expected map to have a 'bar' key, got '%#v'", fooMap)
		}

		barArray, isArray := bar.([]interface{})
		if !isArray {
			t.Errorf("Expected property to be an array, got '%#v'", bar)
		}

		if len(barArray) != 1 {
			t.Errorf("Expected array to have 1 value, got '%#v'", bar)
		}

		baz, isString := barArray[0].(string)
		if !isString {
			t.Errorf("Expected property to be a string, got '%#v'", baz)
		}

		if baz != "baz" {
			t.Errorf("Expected array value to be 'baz', got '%#v'", baz)
		}
	})
}
