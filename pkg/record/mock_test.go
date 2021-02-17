package record

import (
	"fmt"
	"rods/pkg/util"
	"testing"
)

func TestMockGetString(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		record := NewMock(
			map[string]string{
				"col_a": "string",
			},
			map[string]int{},
			map[string]float64{},
			map[string]bool{},
			0,
		)
		expect := "string"
		got, err := record.GetString("col_a")
		if *got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("via .Get", func(t *testing.T) {
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
		expect := util.PString("string")
		if fmt.Sprintf("%v", got) == *expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
}

func TestMockGetInteger(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		record := NewMock(
			map[string]string{},
			map[string]int{
				"col_a": 42,
			},
			map[string]float64{},
			map[string]bool{},
			0,
		)
		expect := 42
		got, err := record.GetInteger("col_a")
		if *got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("via .Get", func(t *testing.T) {
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
		if fmt.Sprintf("%v", got) == fmt.Sprintf("%v", expect) {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
}

func TestMockGetFloat(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		record := NewMock(
			map[string]string{},
			map[string]int{},
			map[string]float64{
				"col_a": 42,
			},
			map[string]bool{},
			0,
		)
		expect := float64(42)
		got, err := record.GetFloat("col_a")
		if *got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("via .Get", func(t *testing.T) {
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
		if fmt.Sprintf("%v", got) == fmt.Sprintf("%v", expect) {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
}

func TestMockGetBoolean(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		record := NewMock(
			map[string]string{},
			map[string]int{},
			map[string]float64{},
			map[string]bool{
				"col_a": true,
			},
			0,
		)
		expect := true
		got, err := record.GetBoolean("col_a")
		if *got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
	t.Run("via .Get", func(t *testing.T) {
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
		if fmt.Sprintf("%v", got) == fmt.Sprintf("%v", expect) {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
		if err != nil {
			t.Errorf("Got error: '%v'", err)
		}
	})
}
