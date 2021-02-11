package config

import (
	"testing"
)

func TestGetAllNonNilFields(t *testing.T) {
	t.Run("one", func(t *testing.T) {
		data := &ServiceConfig{Http: &HttpServiceConfig{}}
		if got, expect := len(getAllNonNilFields(data)), 1; got != expect {
			t.Errorf("Expected to get %v field, got %v", expect, got)
		}
	})
	t.Run("many", func(t *testing.T) {
		data := &OutputConfig{
			JsonArray:  &JsonArrayOutputConfig{},
			JsonObject: &JsonObjectOutputConfig{},
		}
		if got, expect := len(getAllNonNilFields(data)), 2; got != expect {
			t.Errorf("Expected to get %v field, got %v", expect, got)
		}
	})
	t.Run("empty", func(t *testing.T) {
		data := &IndexConfig{}
		if got, expect := len(getAllNonNilFields(data)), 0; got != expect {
			t.Errorf("Expected to get %v field, got %v", expect, got)
		}
	})
}

func TestCheckDuplicateEndpointsPerService(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		data := map[string]OutputConfig{
			"Test": {
				JsonArray: &JsonArrayOutputConfig{
					Service:  "test",
					Endpoint: "/",
				},
			},
		}
		if err := checkDuplicateEndpointsPerService(data); err != nil {
			t.Errorf("Expected to not get an error, got %v", err)
		}
	})
	t.Run("duplicates", func(t *testing.T) {
		data := map[string]OutputConfig{
			"Test": {
				JsonArray: &JsonArrayOutputConfig{
					Service:  "test",
					Endpoint: "/",
				},
			},
			"Test2": {
				JsonArray: &JsonArrayOutputConfig{
					Service:  "test",
					Endpoint: "/",
				},
			},
		}
		if err := checkDuplicateEndpointsPerService(data); err == nil {
			t.Errorf("Expected to get an error, got nil")
		}
	})
	t.Run("empty", func(t *testing.T) {
		data := map[string]OutputConfig{}
		if err := checkDuplicateEndpointsPerService(data); err != nil {
			t.Errorf("Expected to not get an error, got %v", err)
		}
	})
}
