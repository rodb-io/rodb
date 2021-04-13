package parser

import (
	"regexp"
	"rodb.io/pkg/config"
	"testing"
)

func TestFloatParse(t *testing.T) {
	config := &config.FloatParser{
		DecimalSeparator: ".",
		IgnoreCharacters: "$ ,",
	}
	float := NewFloat(config)

	for value, expectedResult := range map[string]interface{}{
		"1.0":           1.0,
		"-1.0":          -1.0,
		"3":             3.0,
		"3.1415":        3.1415,
		"$ 123,456.789": 123456.789,
		"nope":          nil,
	} {
		t.Run(value, func(t *testing.T) {
			got, err := float.Parse(value)
			if expectedResult == nil {
				if err == nil {
					t.Errorf("Expected error, got '%v', '%+v'", got, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got '%v'", err)
				}
				if expectedResult != got {
					t.Errorf("Expected '%+v', got '%v'", expectedResult, got)
				}
			}
		})
	}
}

func TestFloatGetRegexpPattern(t *testing.T) {
	config := &config.FloatParser{
		DecimalSeparator: ".",
		IgnoreCharacters: "$ ,",
	}
	float := NewFloat(config)
	pattern, err := regexp.Compile("^" + float.GetRegexpPattern() + "$")
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	for value, expectedResult := range map[string]interface{}{
		"1":             true,
		"1.0":           true,
		"-2":            true,
		"3.14":          true,
		"$ 123,456.789": true,
		"nope":          false,
		"0%":            false,
	} {
		t.Run(value, func(t *testing.T) {
			got := pattern.MatchString(value)
			if expectedResult != got {
				t.Errorf("Expected '%+v', got '%v' for value '%+v'", expectedResult, got, value)
			}
		})
	}
}
