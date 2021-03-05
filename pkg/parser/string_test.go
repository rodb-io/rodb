package parser

import (
	"github.com/sirupsen/logrus"
	"regexp"
	"rods/pkg/config"
	"testing"
)

func TestStringParse(t *testing.T) {
	config := &config.StringParser{}
	stringParser := NewString(config, logrus.StandardLogger())

	for value, expectedResult := range map[string]interface{}{
		"abc": "abc",
		"123": "123",
		"":    "",
	} {
		t.Run(value, func(t *testing.T) {
			got, err := stringParser.Parse(value)
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

func TestStringGetRegexpPattern(t *testing.T) {
	config := &config.StringParser{}
	stringParser := NewString(config, logrus.StandardLogger())
	pattern := regexp.MustCompile("^" + stringParser.GetRegexpPattern() + "$")

	for value, expectedResult := range map[string]interface{}{
		"abc": true,
		"123": true,
		"":    true,
	} {
		t.Run(value, func(t *testing.T) {
			got := pattern.MatchString(value)
			if expectedResult != got {
				t.Errorf("Expected '%+v', got '%v' for value '%+v'", expectedResult, got, value)
			}
		})
	}
}
