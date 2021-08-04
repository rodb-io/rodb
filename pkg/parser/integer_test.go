package parser

import (
	"regexp"
	"testing"
)

func TestIntegerParse(t *testing.T) {
	config := &IntegerConfig{
		IgnoreCharacters: "$ ,",
	}
	integer := NewInteger(config)

	for value, expectedResult := range map[string]interface{}{
		"1":         int64(1),
		"-42":       int64(-42),
		"3":         int64(3),
		"$ 123,456": int64(123456),
		"nope":      nil,
	} {
		t.Run(value, func(t *testing.T) {
			got, err := integer.Parse(value)
			if expectedResult == nil {
				if err == nil {
					t.Fatalf("Expected error, got '%v', '%+v'", got, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got '%v'", err)
				}
				if expectedResult != got {
					t.Fatalf("Expected '%+v', got '%v'", expectedResult, got)
				}
			}
		})
	}
}

func TestIntegerGetRegexpPattern(t *testing.T) {
	config := &IntegerConfig{
		IgnoreCharacters: "$ ,",
	}
	integer := NewInteger(config)
	pattern, err := regexp.Compile("^" + integer.GetRegexpPattern() + "$")
	if err != nil {
		t.Fatalf("Expected no error, got '%v'", err)
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
				t.Fatalf("Expected '%+v', got '%v' for value '%+v'", expectedResult, got, value)
			}
		})
	}
}
