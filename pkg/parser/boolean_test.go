package parser

import (
	"github.com/sirupsen/logrus"
	"regexp"
	"rods/pkg/config"
	"testing"
)

func TestBooleanParse(t *testing.T) {
	config := &config.BooleanParser{
		TrueValues:  []string{"true", "yes"},
		FalseValues: []string{"false", "no"},
	}
	boolean := NewBoolean(config, logrus.StandardLogger())

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
			got, err := boolean.Parse("true")
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

func TestBooleanGetRegexpPattern(t *testing.T) {
	config := &config.BooleanParser{
		TrueValues:  []string{"true", "yes"},
		FalseValues: []string{"false", "no"},
	}
	boolean := NewBoolean(config, logrus.StandardLogger())

	pattern := regexp.MustCompile(boolean.GetRegexpPattern())

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
				t.Errorf("Expected '%+v', got '%v' for value '%+v'", expectedResult, got, value)
			}
		})
	}
}
