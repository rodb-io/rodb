package parser

import (
	"regexp"
	configModule "rodb.io/pkg/config"
	"testing"
)

func TestSplitParse(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		delimiter := "//"
		falseValue := false
		config := &configModule.SplitParser{
			Delimiter:         &delimiter,
			DelimiterIsRegexp: &falseValue,
			Parser:            "integer",
		}
		splitParser := NewSplit(config, List{
			"integer": NewInteger(&configModule.IntegerParser{}),
		})

		data, err := splitParser.Parse(`1//42`)
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}

		dataArray, isArray := data.([]interface{})
		if !isArray {
			t.Errorf("Expected data to be an array, got '%#v'", data)
		}

		if len(dataArray) != 2 {
			t.Errorf("Expected array to have 2 values, got '%#v'", dataArray)
		}

		data0, isInt := dataArray[0].(int)
		if !isInt {
			t.Errorf("Expected property to be a string, got '%#v'", dataArray)
		}
		if data0 != 1 {
			t.Errorf("Expected array value at index 0 to be '1', got '%v'", data0)
		}

		data1, isInt := dataArray[1].(int)
		if !isInt {
			t.Errorf("Expected property to be a string, got '%#v'", dataArray)
		}
		if data1 != 42 {
			t.Errorf("Expected array value at index 1 to be '42', got '%v'", data1)
		}
	})
	t.Run("regexp", func(t *testing.T) {
		delimiter := "[^0-9]+"
		trueValue := true
		config := &configModule.SplitParser{
			Delimiter:         &delimiter,
			DelimiterIsRegexp: &trueValue,
			DelimiterRegexp:   regexp.MustCompile(delimiter),
			Parser:            "integer",
		}
		splitParser := NewSplit(config, List{
			"integer": NewInteger(&configModule.IntegerParser{}),
		})

		data, err := splitParser.Parse(`1//42`)
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}

		dataArray, isArray := data.([]interface{})
		if !isArray {
			t.Errorf("Expected data to be an array, got '%#v'", data)
		}

		if len(dataArray) != 2 {
			t.Errorf("Expected array to have 2 values, got '%#v'", dataArray)
		}

		data0, isInt := dataArray[0].(int)
		if !isInt {
			t.Errorf("Expected property to be a string, got '%#v'", dataArray)
		}
		if data0 != 1 {
			t.Errorf("Expected array value at index 0 to be '1', got '%v'", data0)
		}

		data1, isInt := dataArray[1].(int)
		if !isInt {
			t.Errorf("Expected property to be a string, got '%#v'", dataArray)
		}
		if data1 != 42 {
			t.Errorf("Expected array value at index 1 to be '42', got '%v'", data1)
		}
	})
}
