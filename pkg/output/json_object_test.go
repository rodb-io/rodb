package output

import (
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/index"
	"rods/pkg/input"
	"rods/pkg/parser"
	"rods/pkg/record"
	"rods/pkg/service"
	"testing"
)

func mockJsonObjectForTests(config *config.JsonObjectOutput) (*JsonObject, *service.Mock, error) {
	mockInput := input.NewMock([]input.IterateAllResult{
		{Record: record.NewStringColumnsMock(map[string]string{
			"id":         "1",
			"belongs_to": "0",
		}, 0)},
		{Record: record.NewStringColumnsMock(map[string]string{
			"id":         "2",
			"belongs_to": "1",
		}, 0)},
		{Record: record.NewStringColumnsMock(map[string]string{
			"id":         "3",
			"belongs_to": "1",
		}, 0)},
		{Record: record.NewStringColumnsMock(map[string]string{
			"id":         "4",
			"belongs_to": "1",
		}, 0)},
	})
	mockIndex := index.NewNoop(
		input.List{"mock": mockInput},
		logrus.StandardLogger(),
	)
	mockService := service.NewMock()
	mockParser := parser.NewMock()
	config.Index = "mock"
	jsonObject, err := NewJsonObject(
		config,
		index.List{"mock": mockIndex},
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
			Parameters: []*config.JsonObjectOutputParameter{
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
			Parameters: []*config.JsonObjectOutputParameter{
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
			Parameters: []*config.JsonObjectOutputParameter{
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
			Parameters: []*config.JsonObjectOutputParameter{
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

func TestJsonObjectLoadRelationships(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonObject, _, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Endpoint: "/test",
			Relationships: map[string]*config.JsonObjectOutputRelationship{
				"children": {
					Input:   "mock",
					Index:   "mock",
					IsArray: true,
					Limit:   2,
					Match: []*config.JsonObjectOutputRelationshipMatch{
						{
							ParentColumn: "id",
							ChildColumn:  "belongs_to",
						},
					},
					Relationships: map[string]*config.JsonObjectOutputRelationship{
						"subchild": {
							Input:   "mock",
							Index:   "mock",
							IsArray: false,
							Match: []*config.JsonObjectOutputRelationshipMatch{
								{
									ParentColumn: "belongs_to",
									ChildColumn:  "id",
								},
							},
						},
					},
				},
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		data := map[string]interface{}{
			"id": "1",
		}
		data, err = jsonObject.loadRelationships(
			data,
			jsonObject.config.Relationships,
		)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := "1", data["id"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'", expect, got)
		}

		if got, ok := data["children"]; !ok {
			t.Errorf("Expected to get an array, got '%+v'", got)
		}
		if got, ok := data["children"].([]map[string]interface{}); !ok {
			t.Errorf("Expected to get an array, got '%+v'", got)
		}

		children := data["children"].([]map[string]interface{})
		if expect, got := 2, len(children); expect != got {
			t.Errorf("Expected length of '%+v', got '%+v'", expect, got)
		}

		if expect, got := "2", children[0]["id"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'", expect, got)
		}
		if expect, got := "3", children[1]["id"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'", expect, got)
		}

		if got, ok := children[0]["subchild"]; !ok {
			t.Errorf("Expected to get an object, got '%+v'", got)
		}
		if got, ok := children[1]["subchild"]; !ok {
			t.Errorf("Expected to get an object, got '%+v'", got)
		}

		if got, ok := children[0]["subchild"].(map[string]interface{}); !ok {
			t.Errorf("Expected to get an object, got '%+v'", got)
		}
		if got, ok := children[1]["subchild"].(map[string]interface{}); !ok {
			t.Errorf("Expected to get an object, got '%+v'", got)
		}

		subchild0 := children[0]["subchild"].(map[string]interface{})
		subchild1 := children[1]["subchild"].(map[string]interface{})

		if expect, got := "1", subchild0["id"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'", expect, got)
		}
		if expect, got := "1", subchild1["id"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'", expect, got)
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
