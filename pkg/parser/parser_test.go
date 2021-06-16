package parser

import (
	"testing"
)

func TestCompare(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		a           interface{}
		b           interface{}
		expectNil   bool
		expectValue bool
	}{
		{"string a < b", "a", "b", false, true},
		{"string a = b", "a", "a", true, false},
		{"string a > b", "b", "a", false, false},

		{"integer a < b", 1, 2, false, true},
		{"integer a = b", 1, 1, true, false},
		{"integer a > b", 2, 1, false, false},

		{"float a < b", 1.0, 2.0, false, true},
		{"float a = b", 1.0, 1.0, true, false},
		{"float a > b", 2.0, 1.0, false, false},

		{"bool a < b", false, true, false, true},
		{"bool a = b", false, false, true, false},
		{"bool a > b", true, false, false, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := Compare(testCase.a, testCase.b)
			if err != nil {
				t.Fatalf("Expected no error, got '%v'", err)
			}
			if testCase.expectNil {
				if result != nil {
					t.Fatalf("Expected to get nil, got '%v'", result)
				}
			} else {
				if result == nil {
					t.Fatalf("Expected to get '%v', got '%v'", testCase.expectValue, result)
				}
				if *result != testCase.expectValue {
					t.Fatalf("Expected to get '%v', got '%v'", testCase.expectValue, *result)
				}
			}
		})
	}
	t.Run("different", func(t *testing.T) {
		_, err := Compare(42, "test")
		if err == nil {
			t.Fatalf("Expected error, got '%v'", err)
		}
	})
	t.Run("unknown", func(t *testing.T) {
		_, err := Compare(struct{}{}, map[string]int{})
		if err == nil {
			t.Fatalf("Expected error, got '%v'", err)
		}
	})
}
