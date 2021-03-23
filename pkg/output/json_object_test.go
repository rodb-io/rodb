package output

import (
	"rods/pkg/config"
	"rods/pkg/service"
	"testing"
)

func mockJsonObjectForTests(config *config.JsonObjectOutput) (*JsonObject, *service.Mock, error) {
	dataForTests := mockJsonDataForTests()
	config.Services = []string{"mock"}
	mockService := service.NewMock()
	services := service.List{"mock": mockService}
	jsonObject, err := NewJsonObject(
		config,
		dataForTests.inputs,
		dataForTests.indexes["default"],
		dataForTests.indexes,
		services,
		dataForTests.parsers,
	)

	return jsonObject, mockService, err
}

func TestJsonObject(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		_, mockService, err := mockJsonObjectForTests(&config.JsonObjectOutput{
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

func TestJsonObjectEndpointRegexp(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonObject, _, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Input:    "mock",
			Endpoint: "/foo/?/bar/?",
			Parameters: []*config.JsonObjectOutputParameter{
				{
					Column: "foo",
					Parser: "mock",
					Index:  "mock",
				}, {
					Column: "bar",
					Parser: "mock",
					Index:  "mock",
				},
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		regexp := jsonObject.endpointRegexp()
		if got, expect := regexp.String(), "^/foo/(?P<param_0>.*)/bar/(?P<param_1>.*)$"; got != expect {
			t.Errorf("Expected regular expression '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("param count lower than wildcard count", func(t *testing.T) {
		jsonObject, _, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Input:    "mock",
			Endpoint: "/foo/?/bar/?",
			Parameters: []*config.JsonObjectOutputParameter{
				{
					Column: "foo",
					Parser: "mock",
					Index:  "mock",
				},
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		regexp := jsonObject.endpointRegexp()
		if got, expect := regexp.String(), "^/foo/(?P<param_0>.*)/bar/(.*)$"; got != expect {
			t.Errorf("Expected regular expression '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("wildcard count lower than param count", func(t *testing.T) {
		jsonObject, _, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Input:    "mock",
			Endpoint: "/foo/?",
			Parameters: []*config.JsonObjectOutputParameter{
				{
					Column: "foo",
					Parser: "mock",
					Index:  "mock",
				}, {
					Column: "bar",
					Parser: "mock",
					Index:  "mock",
				},
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		regexp := jsonObject.endpointRegexp()
		if got, expect := regexp.String(), "^/foo/(?P<param_0>.*)$"; got != expect {
			t.Errorf("Expected regular expression '%+v', got '%+v'", expect, got)
		}
	})
}

func TestJsonObjectGetEndpointFiltersPerIndex(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonObject, _, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Input:    "mock",
			Endpoint: "/foo/?/bar/?/baz/?",
			Parameters: []*config.JsonObjectOutputParameter{
				{
					Column: "foo",
					Parser: "mock",
					Index:  "a",
				}, {
					Column: "bar",
					Parser: "mock",
					Index:  "b",
				}, {
					Column: "baz",
					Parser: "mock",
					Index:  "a",
				},
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		fooParamValue := "fooValue"
		barParamValue := "barValue"
		bazParamValue := "bazValue"
		filtersPerIndex, err := jsonObject.getEndpointFiltersPerIndex(map[string]string{
			jsonObject.endpointRegexpParamName(0): fooParamValue,
			jsonObject.endpointRegexpParamName(1): barParamValue,
			jsonObject.endpointRegexpParamName(2): bazParamValue,
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if got, exists := filtersPerIndex["a"]; !exists {
			t.Errorf("Expected to get filters for index 'a', got '%+v'", got)
		}
		if got, exists := filtersPerIndex["b"]; !exists {
			t.Errorf("Expected to get filters for index 'b', got '%+v'", got)
		}

		if got, expect := len(filtersPerIndex["a"]), 2; got != expect {
			t.Errorf("Expected to get '%+v' filters for index 'a', got '%+v'", expect, got)
		}
		if got, expect := len(filtersPerIndex["b"]), 1; got != expect {
			t.Errorf("Expected to get '%+v' filters for index 'b', got '%+v'", expect, got)
		}

		if got, exists := filtersPerIndex["a"]["foo"]; !exists || got != fooParamValue {
			t.Errorf("Expected to get '%+v' value for filter, got '%+v'", fooParamValue, got)
		}
		if got, exists := filtersPerIndex["a"]["baz"]; !exists || got != bazParamValue {
			t.Errorf("Expected to get '%+v' value for filter, got '%+v'", barParamValue, got)
		}
		if got, exists := filtersPerIndex["b"]["bar"]; !exists || got != barParamValue {
			t.Errorf("Expected to get '%+v' value for filter, got '%+v'", bazParamValue, got)
		}
	})
}

func TestJsonObjectClose(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonObject, mockService, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Input:    "mock",
			Endpoint: "/test",
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if len(mockService.Routes) != 1 {
			t.Errorf("Expected the output to add a route")
		}

		err = jsonObject.Close()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if len(mockService.Routes) != 0 {
			t.Errorf("Expected the .Close call to remove the route from the service")
		}
	})
}
