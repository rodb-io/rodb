package output

import (
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
		}, 1)},
		{Record: record.NewStringColumnsMock(map[string]string{
			"id":         "3",
			"belongs_to": "1",
		}, 2)},
		{Record: record.NewStringColumnsMock(map[string]string{
			"id":         "4",
			"belongs_to": "1",
		}, 3)},
	})
	mockIndex := index.NewNoop(
		input.List{"mock": mockInput},
	)
	mockIndex2 := index.NewNoop(
		input.List{"mock": mockInput},
	)
	mockService := service.NewMock()
	mockParser := parser.NewMock()
	config.Services = []string{"mock"}
	jsonObject, err := NewJsonObject(
		config,
		input.List{"mock": mockInput},
		index.List{"mock": mockIndex, "mock2": mockIndex2},
		service.List{"mock": mockService},
		parser.List{"mock": mockParser},
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

func TestJsonObjectGetRelationshipFiltersPerIndex(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonObject, _, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Input:    "mock",
			Endpoint: "/test",
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		filtersPerIndex, err := jsonObject.getRelationshipFiltersPerIndex(
			map[string]interface{}{
				"foo": "3",
				"bar": "1",
			},
			[]*config.RelationshipMatch{
				{
					ParentColumn: "foo",
					ChildColumn:  "foo",
					ChildIndex:   "a",
				}, {
					ParentColumn: "foo",
					ChildColumn:  "foo",
					ChildIndex:   "b",
				}, {
					ParentColumn: "bar",
					ChildColumn:  "bar",
					ChildIndex:   "b",
				},
			},
			"test",
		)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if got, exists := filtersPerIndex["a"]; !exists {
			t.Errorf("Expected to get filters for index 'a', got '%+v'", got)
		}
		if got, exists := filtersPerIndex["b"]; !exists {
			t.Errorf("Expected to get filters for index 'b', got '%+v'", got)
		}

		if got, expect := len(filtersPerIndex["a"]), 1; got != expect {
			t.Errorf("Expected to get '%+v' filters for index 'a', got '%+v'", expect, got)
		}
		if got, expect := len(filtersPerIndex["b"]), 2; got != expect {
			t.Errorf("Expected to get '%+v' filters for index 'b', got '%+v'", expect, got)
		}

		if got, exists := filtersPerIndex["a"]["foo"]; !exists || got != "3" {
			t.Errorf("Expected to get '%+v' value for filter, got '%+v'", "3", got)
		}
		if got, exists := filtersPerIndex["a"]["foo"]; !exists || got != "3" {
			t.Errorf("Expected to get '%+v' value for filter, got '%+v'", "3", got)
		}
		if got, exists := filtersPerIndex["b"]["bar"]; !exists || got != "1" {
			t.Errorf("Expected to get '%+v' value for filter, got '%+v'", "1", got)
		}
	})
}

func TestJsonObjectGetFilteredRecordPositionsPerIndex(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonObject, _, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Input:    "mock",
			Endpoint: "/test",
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		filtersPerIndex := map[string]map[string]interface{}{
			"mock": {
				"id": "2",
			},
			"mock2": {
				"belongs_to": "1",
			},
		}

		recordLists, err := jsonObject.getFilteredRecordPositionsPerIndex(0, filtersPerIndex)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := 2, len(recordLists); got != expect {
			t.Errorf("Expected to get '%+v' entries in the array, got '%+v'", expect, got)
		}

		// Not working, because the map does not guarantee the order
		// if expect, got := 1, len(recordLists[0]); got != expect {
		// 	t.Errorf("Expected to get '%+v' entries in the first array, got '%+v'", expect, got)
		// }
		// if expect, got := 3, len(recordLists[1]); got != expect {
		// 	t.Errorf("Expected to get '%+v' entries in the second array, got '%+v'", expect, got)
		// }

		if expect, got := int64(1), recordLists[0][0]; got != expect {
			t.Errorf("Expected to get position '%+v' for the first result of the first index, got '%+v'", expect, got)
		}
		if expect, got := int64(1), recordLists[1][0]; got != expect {
			t.Errorf("Expected to get position '%+v' for the first result of the second index, got '%+v'", expect, got)
		}
	})
}

func TestJsonObjectLoadRelationships(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		ascendingSort := false
		jsonObject, _, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Endpoint: "/test",
			Input:    "mock",
			Relationships: map[string]*config.Relationship{
				"children": {
					Input:   "mock",
					IsArray: true,
					Limit:   2,
					Sort: []*config.Sort{
						{
							Column:    "id",
							Ascending: &ascendingSort,
						},
					},
					Match: []*config.RelationshipMatch{
						{
							ParentColumn: "id",
							ChildColumn:  "belongs_to",
							ChildIndex:   "mock",
						},
					},
					Relationships: map[string]*config.Relationship{
						"subchild": {
							Input:   "mock",
							IsArray: false,
							Match: []*config.RelationshipMatch{
								{
									ParentColumn: "belongs_to",
									ChildColumn:  "id",
									ChildIndex:   "mock",
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

		// The sort result is only quickly tested, because record.List.Sort is already tested
		if expect, got := "3", children[0]["id"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'", expect, got)
		}
		if expect, got := "2", children[1]["id"]; expect != got {
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
