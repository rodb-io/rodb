package output

import (
	"rods/pkg/config"
	"rods/pkg/service"
	"testing"
)

func mockJsonArrayForTests(config *config.JsonArrayOutput) (*JsonArray, *service.Mock, error) {
	dataForTests := mockJsonDataForTests()
	config.Services = []string{"mock"}
	mockService := service.NewMock()
	services := service.List{"mock": mockService}
	jsonArray, err := NewJsonArray(
		config,
		dataForTests.inputs,
		dataForTests.indexes["default"],
		dataForTests.indexes,
		services,
		dataForTests.parsers,
	)

	return jsonArray, mockService, err
}

func TestJsonArray(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		_, mockService, err := mockJsonArrayForTests(&config.JsonArrayOutput{
			Input:    "mock",
			Endpoint: "/test",
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if len(mockService.Routes) != 1 {
			t.Errorf("Expected the output to add a route")
		}
	})
}

func TestJsonArrayGetLimit(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonArray, _, err := mockJsonArrayForTests(&config.JsonArrayOutput{
			Input:    "mock",
			Endpoint: "/test",
			Limit: config.JsonArrayOutputLimit{
				Default:   10,
				Max:       150,
				Parameter: "testlimit",
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		limit, err := jsonArray.getLimit(map[string]string{
			"testlimit": "123",
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if got, expect := limit, uint(123); got != expect {
			t.Errorf("Expected to get '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("max", func(t *testing.T) {
		jsonArray, _, err := mockJsonArrayForTests(&config.JsonArrayOutput{
			Input:    "mock",
			Endpoint: "/test",
			Limit: config.JsonArrayOutputLimit{
				Default:   10,
				Max:       50,
				Parameter: "testlimit",
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		limit, err := jsonArray.getLimit(map[string]string{
			"testlimit": "123",
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if got, expect := limit, uint(50); got != expect {
			t.Errorf("Expected to get '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("default", func(t *testing.T) {
		jsonArray, _, err := mockJsonArrayForTests(&config.JsonArrayOutput{
			Input:    "mock",
			Endpoint: "/test",
			Limit: config.JsonArrayOutputLimit{
				Default:   12,
				Max:       50,
				Parameter: "testlimit",
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		limit, err := jsonArray.getLimit(map[string]string{})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if got, expect := limit, uint(12); got != expect {
			t.Errorf("Expected to get '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("negative", func(t *testing.T) {
		jsonArray, _, err := mockJsonArrayForTests(&config.JsonArrayOutput{
			Input:    "mock",
			Endpoint: "/test",
			Limit: config.JsonArrayOutputLimit{
				Default:   10,
				Max:       50,
				Parameter: "testlimit",
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		_, err = jsonArray.getLimit(map[string]string{
			"testlimit": "-42",
		})
		if err == nil {
			t.Errorf("Expected error, got nil.")
		}
	})
}

func TestJsonArrayGetOffset(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonArray, _, err := mockJsonArrayForTests(&config.JsonArrayOutput{
			Input:    "mock",
			Endpoint: "/test",
			Offset: config.JsonArrayOutputOffset{
				Parameter: "testoffset",
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		offset, err := jsonArray.getOffset(map[string]string{
			"testoffset": "123",
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if got, expect := offset, uint(123); got != expect {
			t.Errorf("Expected to get '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("negative", func(t *testing.T) {
		jsonArray, _, err := mockJsonArrayForTests(&config.JsonArrayOutput{
			Input:    "mock",
			Endpoint: "/test",
			Offset: config.JsonArrayOutputOffset{
				Parameter: "testoffset",
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		_, err = jsonArray.getOffset(map[string]string{
			"testoffset": "-42",
		})
		if err == nil {
			t.Errorf("Expected error, got nil.")
		}
	})
}

func TestJsonArrayGetFiltersPerIndex(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonArray, _, err := mockJsonArrayForTests(&config.JsonArrayOutput{
			Input:    "mock",
			Endpoint: "/test",
			Search: map[string]config.JsonArrayOutputSearch{
				"a": {
					Column: "a",
					Index:  "a",
					Parser: "mock",
				},
				"b": {
					Column: "b",
					Index:  "a",
					Parser: "prefix",
				},
				"c": {
					Column: "c",
					Index:  "b",
					Parser: "prefix",
				},
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		filters, err := jsonArray.getFiltersPerIndex(map[string]string{
			"a": "val-a",
			"b": "val-b",
			"c": "val-c",
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if _, exists := filters["a"]; !exists {
			t.Errorf("Expected to have filters for index 'a'")
		}
		if _, exists := filters["b"]; !exists {
			t.Errorf("Expected to have filters for index 'b'")
		}

		if val, _ := filters["a"]; len(val) != 2 {
			t.Errorf("Expected to have 2 filters for index 'a', got '%v'", len(val))
		}
		if val, _ := filters["b"]; len(val) != 1 {
			t.Errorf("Expected to have 1 filter for index 'b', got '%v'", len(val))
		}

		if _, exists := filters["a"]["a"]; !exists {
			t.Errorf("Expected to have a filter 'a' for index 'a'")
		}
		if _, exists := filters["a"]["b"]; !exists {
			t.Errorf("Expected to have a filter 'b' for index 'a'")
		}
		if _, exists := filters["b"]["c"]; !exists {
			t.Errorf("Expected to have a filter 'c' for index 'b'")
		}

		if expect, got := "val-a", filters["a"]["a"]; got != expect {
			t.Errorf("Expected to have value '%v' for filter 'a' of index 'a', got '%v'", expect, got)
		}
		if expect, got := "prefix_val-b", filters["a"]["b"]; got != expect {
			t.Errorf("Expected to have value '%v' for filter 'b' of index 'a', got '%v'", expect, got)
		}
		if expect, got := "prefix_val-c", filters["b"]["c"]; got != expect {
			t.Errorf("Expected to have value '%v' for filter 'c' of index 'b', got '%v'", expect, got)
		}
	})
}
