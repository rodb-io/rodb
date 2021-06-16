package parser

import (
	"regexp"
	"rodb.io/pkg/config"
	"testing"
)

func TestBooleanParse(t *testing.T) {
	config := &config.BooleanParser{
		TrueValues:  []string{"true", "yes"},
		FalseValues: []string{"false", "no"},
	}
	boolean := NewBoolean(config)

	for value, expectedResult := range map[string]interface{}{
		"true":  true,
		"yes":   true,
		"false": false,
		"no":    false,
		"0":     nil,
		"1":     nil,
		"TRUE":  nil,
		"FALSE": nil,
	} {
		t.Run(value, func(t *testing.T) {
			got, err := boolean.Parse(value)
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

func TestBooleanGetRegexpPattern(t *testing.T) {
	config := &config.BooleanParser{
		TrueValues:  []string{"true", "yes"},
		FalseValues: []string{"false", "no"},
	}
	boolean := NewBoolean(config)
	pattern, err := regexp.Compile("^" + boolean.GetRegexpPattern() + "$")
	if err != nil {
		t.Fatalf("Expected no error, got '%v'", err)
	}

	for value, expectedResult := range map[string]interface{}{
		"true":  true,
		"yes":   true,
		"false": true,
		"no":    true,
		"0":     false,
		"1":     false,
		"TRUE":  false,
		"FALSE": false,
	} {
		t.Run(value, func(t *testing.T) {
			got := pattern.MatchString(value)
			if expectedResult != got {
				t.Fatalf("Expected '%+v', got '%v' for value '%+v'", expectedResult, got, value)
			}
		})
	}
}
