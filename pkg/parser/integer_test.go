package parser

import (
	"regexp"
	"rodb.io/pkg/config"
	"testing"
)

func TestIntegerParse(t *testing.T) {
	config := &config.IntegerParser{
		IgnoreCharacters: "$ ,",
	}
	integer := NewInteger(config)

	for value, expectedResult := range map[string]interface{}{
		"1":         1,
		"-42":       -42,
		"3":         3,
		"$ 123,456": 123456,
		"nope":      nil,
	} {
		t.Run(value, func(t *testing.T) {
			got, err := integer.Parse(value)
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

func TestIntegerGetRegexpPattern(t *testing.T) {
	config := &config.IntegerParser{
		IgnoreCharacters: "$ ,",
	}
	integer := NewInteger(config)
	pattern, err := regexp.Compile("^" + integer.GetRegexpPattern() + "$")
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	for value, expectedResult := range map[string]interface{}{
		"1":         true,
		"-2":        true,
		"42":        true,
		"$ 123,456": true,
		"nope":      false,
		"0%":        false,
	} {
		t.Run(value, func(t *testing.T) {
			got := pattern.MatchString(value)
			if expectedResult != got {
				t.Errorf("Expected '%+v', got '%v' for value '%+v'", expectedResult, got, value)
			}
		})
	}
}
