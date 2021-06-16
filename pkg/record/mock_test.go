package record

import (
	"fmt"
	"testing"
)

func TestMockGet(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		record := NewMock(
			map[string]string{
				"col_a": "string",
			},
			map[string]int{},
			map[string]float64{},
			map[string]bool{},
			0,
		)
		got, err := record.Get("col_a")
		expect := "string"
		if fmt.Sprintf("%v", got) != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Fatalf("Got error: '%v'", err)
		}
	})
	t.Run("integer", func(t *testing.T) {
		record := NewMock(
			map[string]string{},
			map[string]int{
				"col_a": 42,
			},
			map[string]float64{},
			map[string]bool{},
			0,
		)
		got, err := record.Get("col_a")
		expect := 42
		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", expect) {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Fatalf("Got error: '%v'", err)
		}
	})
	t.Run("float", func(t *testing.T) {
		record := NewMock(
			map[string]string{},
			map[string]int{},
			map[string]float64{
				"col_a": 42,
			},
			map[string]bool{},
			0,
		)
		got, err := record.Get("col_a")
		expect := float64(42)
		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", expect) {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Fatalf("Got error: '%v'", err)
		}
	})
	t.Run("boolean", func(t *testing.T) {
		record := NewMock(
			map[string]string{},
			map[string]int{},
			map[string]float64{},
			map[string]bool{
				"col_a": false,
			},
			0,
		)
		got, err := record.Get("col_a")
		expect := false
		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", expect) {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Fatalf("Got error: '%v'", err)
		}
	})
}
