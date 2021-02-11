package config

import (
	"testing"
)

func TestGetAllNonNilFields(t *testing.T) {
	t.Run("one", func(t *testing.T) {
		data := &Service{Http: &HttpService{}}
		if got, expect := len(getAllNonNilFields(data)), 1; got != expect {
			t.Errorf("Expected to get %v field, got %v", expect, got)
		}
	})
	t.Run("many", func(t *testing.T) {
		data := &Output{
			JsonArray:  &JsonArrayOutput{},
			JsonObject: &JsonObjectOutput{},
		}
		if got, expect := len(getAllNonNilFields(data)), 2; got != expect {
			t.Errorf("Expected to get %v field, got %v", expect, got)
		}
	})
	t.Run("empty", func(t *testing.T) {
		data := &Index{}
		if got, expect := len(getAllNonNilFields(data)), 0; got != expect {
			t.Errorf("Expected to get %v field, got %v", expect, got)
		}
	})
}

func TestCheckDuplicateEndpointsPerService(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		data := map[string]Output{
			"Test": {
				JsonArray: &JsonArrayOutput{
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
		data := map[string]Output{
			"Test": {
				JsonArray: &JsonArrayOutput{
					Service:  "test",
					Endpoint: "/",
				},
			},
			"Test2": {
				JsonArray: &JsonArrayOutput{
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
		data := map[string]Output{}
		if err := checkDuplicateEndpointsPerService(data); err != nil {
			t.Errorf("Expected to not get an error, got %v", err)
		}
	})
}
