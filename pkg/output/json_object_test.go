package output

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"rods/pkg/config"
	"rods/pkg/record"
	"testing"
)

func mockJsonObjectForTests(config *config.JsonObjectOutput) (*JsonObject, error) {
	dataForTests := mockJsonDataForTests()
	jsonObject, err := NewJsonObject(
		config,
		dataForTests.inputs,
		dataForTests.indexes["default"],
		dataForTests.indexes,
		dataForTests.parsers,
	)

	return jsonObject, err
}

func TestJsonObjectHandler(t *testing.T) {
	trueValue := true
	jsonObject, err := mockJsonObjectForTests(&config.JsonObjectOutput{
		Input:    "mock",
		Endpoint: "/foo/{foo_id}",
		Parameters: map[string]*config.JsonObjectOutputParameter{
			"foo_id": {
				Column: "id",
				Parser: "mock",
				Index:  "mock",
			},
		},
		Relationships: map[string]*config.Relationship{
			"child": {
				Input:   "mock",
				IsArray: false,
				Match: []*config.RelationshipMatch{
					{
						ParentColumn: "belongs_to",
						ChildColumn:  "id",
						ChildIndex:   "mock",
					},
				},
				Relationships: map[string]*config.Relationship{
					"subchild": {
						Input:   "mock",
						IsArray: true,
						Limit:   2,
						Sort: []*config.Sort{
							{
								Column:    "id",
								Ascending: &trueValue,
							},
						},
						Match: []*config.RelationshipMatch{
							{
								ParentColumn: "id",
								ChildColumn:  "belongs_to",
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

	getResult := func(id string) (map[string]interface{}, error) {
		buffer := bytes.NewBufferString("")
		err := jsonObject.Handle(
			map[string]string{
				jsonObject.endpointRegexpParamName(0): id,
			},
			[]byte{},
			func(err error) error {
				return err
			},
			func() io.Writer {
				return buffer
			},
		)
		if err != nil {
			return nil, err
		}

		bytesOutput, err := ioutil.ReadAll(buffer)
		if err != nil {
			return nil, err
		}

		data := map[string]interface{}{}
		err = json.Unmarshal(bytesOutput, &data)
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	t.Run("normal", func(t *testing.T) {
		data, err := getResult("2")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := "2", data["id"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'.", expect, got)
		}
		if expect, got := "1", data["belongs_to"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'.", expect, got)
		}
		if _, exists := data["child"]; !exists {
			t.Errorf("Expected to get a 'child' property, got none.")
		}

		child := data["child"].(map[string]interface{})
		if expect, got := "1", child["id"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'.", expect, got)
		}
		if expect, got := "0", child["belongs_to"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'.", expect, got)
		}
		if _, exists := child["subchild"]; !exists {
			t.Errorf("Expected to get a 'subchild' property, got none.")
		}

		subchild := child["subchild"].([]interface{})
		if expect, got := 2, len(subchild); expect != got {
			t.Errorf("Expected to get '%+v' subchilds, got '%+v'.", expect, got)
		}

		subchild0 := subchild[0].(map[string]interface{})
		if expect, got := "2", subchild0["id"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'.", expect, got)
		}

		subchild1 := subchild[1].(map[string]interface{})
		if expect, got := "3", subchild1["id"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'.", expect, got)
		}
	})
	t.Run("no child", func(t *testing.T) {
		data, err := getResult("1")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := "1", data["id"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'.", expect, got)
		}
		if expect, got := "0", data["belongs_to"]; expect != got {
			t.Errorf("Expected to get '%+v', got '%+v'.", expect, got)
		}
		if _, exists := data["child"]; !exists {
			t.Errorf("Expected to get a 'child' property, got none.")
		}
		if child := data["child"]; child != nil {
			t.Errorf("Expected to get a 'child' property equal to nil.")
		}
	})
	t.Run("no child", func(t *testing.T) {
		_, err := getResult("99")
		if err != record.RecordNotFoundError {
			t.Errorf("Expected to get a 404 error, got: '%+v'", err)
		}
	})
}

func TestJsonObjectEndpointRegexp(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonObject, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Input:    "mock",
			Endpoint: "/foo/{foo_id}/bar/{bar_id}",
			Parameters: map[string]*config.JsonObjectOutputParameter{
				"foo_id": {
					Column: "foo",
					Parser: "mock",
					Index:  "mock",
				},
				"bar_id": {
					Column: "bar",
					Parser: "mock",
					Index:  "mock",
				},
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		regexp := jsonObject.Endpoint()
		if got, expect := regexp.String(), "^/foo/(?P<param_0>.*)/bar/(?P<param_1>.*)$"; got != expect {
			t.Errorf("Expected regular expression '%+v', got '%+v'", expect, got)
		}

		if got, expect := len(jsonObject.endpointParams), 2; got != expect {
			t.Errorf("Expected '%+v' endpoint params, got '%+v'", expect, got)
		}

		if got, expect := jsonObject.endpointParams[0], "foo_id"; got != expect {
			t.Errorf("Expected the first endpoint param to be '%+v', got '%+v'", expect, got)
		}
		if got, expect := jsonObject.endpointParams[1], "bar_id"; got != expect {
			t.Errorf("Expected the second endpoint param to be '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("param in the endpoint does not exists", func(t *testing.T) {
		_, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Input:    "mock",
			Endpoint: "/foo/{foo_id}/bar/{bar_id}",
			Parameters: map[string]*config.JsonObjectOutputParameter{
				"foo_id": {
					Column: "foo",
					Parser: "mock",
					Index:  "mock",
				},
			},
		})
		if err == nil {
			t.Errorf("Expected error, got '%+v'", err)
		}
	})
	t.Run("wildcard count lower than param count", func(t *testing.T) {
		jsonObject, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Input:    "mock",
			Endpoint: "/foo/{foo_id}",
			Parameters: map[string]*config.JsonObjectOutputParameter{
				"foo_id": {
					Column: "foo",
					Parser: "mock",
					Index:  "mock",
				},
				"bar_id": {
					Column: "bar",
					Parser: "mock",
					Index:  "mock",
				},
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		regexp := jsonObject.Endpoint()
		if got, expect := regexp.String(), "^/foo/(?P<param_0>.*)$"; got != expect {
			t.Errorf("Expected regular expression '%+v', got '%+v'", expect, got)
		}
	})
}

func TestJsonObjectGetEndpointFiltersPerIndex(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonObject, err := mockJsonObjectForTests(&config.JsonObjectOutput{
			Input:    "mock",
			Endpoint: "/foo/{foo_id}/bar/{bar_id}/baz/{baz_id}",
			Parameters: map[string]*config.JsonObjectOutputParameter{
				"foo_id": {
					Column: "foo",
					Parser: "mock",
					Index:  "a",
				},
				"bar_id": {
					Column: "bar",
					Parser: "mock",
					Index:  "b",
				},
				"baz_id": {
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
