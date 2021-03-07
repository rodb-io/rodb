package output

import (
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/index"
	"rods/pkg/input"
	"rods/pkg/parser"
	"rods/pkg/service"
	"testing"
)

func mockJsonObjectForTests(config *config.JsonObjectOutput) (*JsonObject, *service.Mock, error) {
	mockInput := input.NewMock([]input.IterateAllResult{})
	mockIndex := index.NewDumb(
		input.List{"mock": mockInput},
		logrus.StandardLogger(),
	)
	mockService := service.NewMock()
	mockParser := parser.NewMock()
	jsonObject, err := NewJsonObject(
		config,
		mockIndex,
		[]service.Service{mockService},
		parser.List{"mock": mockParser},
		logrus.StandardLogger(),
	)

	return jsonObject, mockService, err
}

func TestJsonObject(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		_, mockService, err := mockJsonObjectForTests(&config.JsonObjectOutput{
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
			Endpoint: "/foo/?/bar/?",
			Parameters: []config.JsonObjectOutputParams{
				{
					Column: "foo",
					Parser: "mock",
				}, {
					Column: "bar",
					Parser: "mock",
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
			Endpoint: "/foo/?/bar/?",
			Parameters: []config.JsonObjectOutputParams{
				{
					Column: "foo",
					Parser: "mock",
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
			Endpoint: "/foo/?",
			Parameters: []config.JsonObjectOutputParams{
				{
					Column: "foo",
					Parser: "mock",
				}, {
					Column: "bar",
					Parser: "mock",
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

func TestJsonObjectGetEndpointFilters(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonObject, _, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Endpoint: "/foo/?/bar/?",
			Parameters: []config.JsonObjectOutputParams{
				{
					Column: "foo",
					Parser: "mock",
				}, {
					Column: "bar",
					Parser: "mock",
				},
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		fooParamValue := "fooValue"
		barParamValue := "barValue"
		filters, err := jsonObject.getEndpointFilters(map[string]string{
			jsonObject.endpointRegexpParamName(0): fooParamValue,
			jsonObject.endpointRegexpParamName(1): barParamValue,
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if got, expect := len(filters), 2; got != expect {
			t.Errorf("Expected to get '%+v' filters, got '%+v'", expect, got)
		}

		if got, exists := filters["foo"]; !exists || got != fooParamValue {
			t.Errorf("Expected to get '%+v' value for filter, got '%+v'", fooParamValue, got)
		}
		if got, exists := filters["bar"]; !exists || got != barParamValue {
			t.Errorf("Expected to get '%+v' value for filter, got '%+v'", barParamValue, got)
		}
	})
}

func TestJsonObjectClose(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonObject, mockService, err := mockJsonObjectForTests(&config.JsonObjectOutput{
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
