package util

import (
	"testing"
)

func TestGetTypeFromConfigUnmarshaler(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		unmarshal := func(data interface{}) error {
			dataMap := data.(map[string]interface{})
			dataMap["type"] = "test"
			return nil
		}

		got, err := GetTypeFromConfigUnmarshaler(unmarshal)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		if expect := "test"; got != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("missing type", func(t *testing.T) {
		unmarshal := func(data interface{}) error {
			return nil
		}

		_, err := GetTypeFromConfigUnmarshaler(unmarshal)
		if err == nil {
			t.Fatalf("Expected error, got '%v'", err)
		}
	})
	t.Run("wrong type", func(t *testing.T) {
		unmarshal := func(data interface{}) error {
			dataMap := data.(map[string]interface{})
			dataMap["type"] = 42
			return nil
		}

		_, err := GetTypeFromConfigUnmarshaler(unmarshal)
		if err == nil {
			t.Fatalf("Expected error, got '%v'", err)
		}
	})
	t.Run("empty", func(t *testing.T) {
		unmarshal := func(data interface{}) error {
			dataMap := data.(map[string]interface{})
			dataMap["type"] = ""
			return nil
		}

		_, err := GetTypeFromConfigUnmarshaler(unmarshal)
		if err == nil {
			t.Fatalf("Expected error, got '%v'", err)
		}
	})
}
