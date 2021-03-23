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
}
