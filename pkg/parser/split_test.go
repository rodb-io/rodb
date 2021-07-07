package parser

import (
	"regexp"
	"testing"
)

func TestSplitParse(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		delimiter := "//"
		falseValue := false
		config := &SplitConfig{
			Delimiter:         &delimiter,
			DelimiterIsRegexp: &falseValue,
			Parser:            "integer",
		}
		splitParser := NewSplit(config, List{
			"integer": NewInteger(&IntegerConfig{}),
		})

		data, err := splitParser.Parse(`1//42`)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		dataArray, isArray := data.([]interface{})
		if !isArray {
			t.Fatalf("Expected data to be an array, got '%#v'", data)
		}

		if len(dataArray) != 2 {
			t.Fatalf("Expected array to have 2 values, got '%#v'", dataArray)
		}

		data0, isInt := dataArray[0].(int)
		if !isInt {
			t.Fatalf("Expected property to be a string, got '%#v'", dataArray)
		}
		if data0 != 1 {
			t.Fatalf("Expected array value at index 0 to be '1', got '%v'", data0)
		}

		data1, isInt := dataArray[1].(int)
		if !isInt {
			t.Fatalf("Expected property to be a string, got '%#v'", dataArray)
		}
		if data1 != 42 {
			t.Fatalf("Expected array value at index 1 to be '42', got '%v'", data1)
		}
	})
	t.Run("regexp", func(t *testing.T) {
		delimiter := "[^0-9]+"
		trueValue := true
		config := &SplitConfig{
			Delimiter:         &delimiter,
			DelimiterIsRegexp: &trueValue,
			DelimiterRegexp:   regexp.MustCompile(delimiter),
			Parser:            "integer",
		}
		splitParser := NewSplit(config, List{
			"integer": NewInteger(&IntegerConfig{}),
		})

		data, err := splitParser.Parse(`1//42`)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		dataArray, isArray := data.([]interface{})
		if !isArray {
			t.Fatalf("Expected data to be an array, got '%#v'", data)
		}

		if len(dataArray) != 2 {
			t.Fatalf("Expected array to have 2 values, got '%#v'", dataArray)
		}

		data0, isInt := dataArray[0].(int)
		if !isInt {
			t.Fatalf("Expected property to be a string, got '%#v'", dataArray)
		}
		if data0 != 1 {
			t.Fatalf("Expected array value at index 0 to be '1', got '%v'", data0)
		}

		data1, isInt := dataArray[1].(int)
		if !isInt {
			t.Fatalf("Expected property to be a string, got '%#v'", dataArray)
		}
		if data1 != 42 {
			t.Fatalf("Expected array value at index 1 to be '42', got '%v'", data1)
		}
	})
}
