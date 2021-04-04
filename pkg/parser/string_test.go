package parser

import (
	"regexp"
	"rods/pkg/config"
	"testing"
)

func TestStringParse(t *testing.T) {
	for value, expectedResult := range map[string]interface{}{
		"abc": "abc",
		"123": "123",
		"":    "",
	} {
		t.Run(value, func(t *testing.T) {
			config := &config.StringParser{}
			stringParser, err := NewString(config)
			if err != nil {
				t.Errorf("Expected no error, got '%v'", err)
			}

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

	t.Run("convertFromCharset", func(t *testing.T) {
		config := &config.StringParser{
			ConvertFromCharset: "Shift_JIS",
		}
		stringParser, err := NewString(config)
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}

		value := string([]byte{147, 140, 139, 158, 147, 115})
		expectedResult := "東京都"

		got, err := stringParser.Parse(value)
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}

		if expectedResult != got {
			t.Errorf("Expected '%+v', got '%v'", expectedResult, got)
		}
	})
}

func TestStringGetRegexpPattern(t *testing.T) {
	config := &config.StringParser{}
	stringParser, err := NewString(config)
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	pattern, err := regexp.Compile("^" + stringParser.GetRegexpPattern() + "$")
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	for value, expectedResult := range map[string]interface{}{
		"abc": true,
		"123": true,
		"":    false,
	} {
		t.Run(value, func(t *testing.T) {
			got := pattern.MatchString(value)
			if expectedResult != got {
				t.Errorf("Expected '%+v', got '%v' for value '%+v'", expectedResult, got, value)
			}
		})
	}
}
