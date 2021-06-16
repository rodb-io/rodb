package parser

import (
	"regexp"
	"testing"
)

func TestMockParse(t *testing.T) {
	mock := NewMock()

	for value, expectedResult := range map[string]interface{}{
		"abc": "abc",
		"123": "123",
		"":    "",
	} {
		t.Run(value, func(t *testing.T) {
			got, err := mock.Parse(value)
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

func TestMockGetRegexpPattern(t *testing.T) {
	mock := NewMock()
	pattern, err := regexp.Compile("^" + mock.GetRegexpPattern() + "$")
	if err != nil {
		t.Fatalf("Expected no error, got '%v'", err)
	}

	for value, expectedResult := range map[string]interface{}{
		"abc": true,
		"123": true,
		"":    true,
	} {
		t.Run(value, func(t *testing.T) {
			got := pattern.MatchString(value)
			if expectedResult != got {
				t.Fatalf("Expected '%+v', got '%v' for value '%+v'", expectedResult, got, value)
			}
		})
	}
}

func TestMockWithPrefix(t *testing.T) {
	mock := NewMockWithPrefix("prefix_")

	for value, expectedResult := range map[string]interface{}{
		"abc": "prefix_abc",
		"123": "prefix_123",
		"":    "prefix_",
	} {
		t.Run(value, func(t *testing.T) {
			got, err := mock.Parse(value)
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
