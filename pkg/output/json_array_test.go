package output

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	parameterPackage "rodb.io/pkg/output/parameter"
	relationshipPackage "rodb.io/pkg/output/relationship"
	"testing"
)

func mockJsonArrayForTests(config *JsonArrayConfig) (*JsonArray, error) {
	dataForTests := mockJsonDataForTests()
	jsonArray, err := NewJsonArray(
		config,
		dataForTests.inputs,
		dataForTests.indexes["default"],
		dataForTests.indexes,
		dataForTests.parsers,
	)

	return jsonArray, err
}

func TestJsonArrayHandler(t *testing.T) {
	jsonArray, err := mockJsonArrayForTests(&JsonArrayConfig{
		Input: "mock",
		Limit: *&JsonArrayLimitConfig{
			Max:       100,
			Default:   10,
			Parameter: "limit",
		},
		Offset: *&JsonArrayOffsetConfig{
			Parameter: "offset",
		},
		Parameters: map[string]parameterPackage.ParameterConfig{
			"belongs_to_param": {
				Property: "belongs_to",
				Parser:   "mock",
				Index:    "mock",
			},
		},
		Relationships: map[string]*relationshipPackage.RelationshipConfig{
			"child": {
				Input:   "mock",
				IsArray: false,
				Match: []*relationshipPackage.RelationshipMatchConfig{
					{
						ParentProperty: "belongs_to",
						ChildProperty:  "id",
						ChildIndex:     "mock",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	getResult := func(params map[string]string) ([]interface{}, error) {
		buffer := bytes.NewBufferString("")
		err := jsonArray.Handle(
			params,
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

		data := []interface{}{}
		if err := json.Unmarshal(bytesOutput, &data); err != nil {
			return nil, err
		}

		return data, nil
	}

	t.Run("normal", func(t *testing.T) {
		data, err := getResult(map[string]string{})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect, got := 4, len(data); expect != got {
			t.Fatalf("Expected to get '%+v' items, got '%+v'.", expect, got)
		}

		row0 := data[0].(map[string]interface{})
		if expect, got := "1", row0["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}

		row1 := data[1].(map[string]interface{})
		if expect, got := "2", row1["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}

		row2 := data[2].(map[string]interface{})
		if expect, got := "3", row2["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}

		row3 := data[3].(map[string]interface{})
		if expect, got := "4", row3["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}
		if _, exists := row3["child"]; !exists {
			t.Fatalf("Expected to get a 'child' property, got none.")
		}

		row3Child := row3["child"].(map[string]interface{})
		if expect, got := "1", row3Child["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}
		if expect, got := "0", row3Child["belongs_to"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}
	})
	t.Run("filter", func(t *testing.T) {
		data, err := getResult(map[string]string{
			"belongs_to_param": "1",
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect, got := 3, len(data); expect != got {
			t.Fatalf("Expected to get '%+v' items, got '%+v'.", expect, got)
		}

		row0 := data[0].(map[string]interface{})
		if expect, got := "2", row0["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}

		row1 := data[1].(map[string]interface{})
		if expect, got := "3", row1["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}

		row2 := data[2].(map[string]interface{})
		if expect, got := "4", row2["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}
	})
	t.Run("limit", func(t *testing.T) {
		data, err := getResult(map[string]string{
			"limit": "2",
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect, got := 2, len(data); expect != got {
			t.Fatalf("Expected to get '%+v' items, got '%+v'.", expect, got)
		}

		row0 := data[0].(map[string]interface{})
		if expect, got := "1", row0["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}

		row1 := data[1].(map[string]interface{})
		if expect, got := "2", row1["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}
	})
	t.Run("offset", func(t *testing.T) {
		data, err := getResult(map[string]string{
			"offset": "2",
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect, got := 2, len(data); expect != got {
			t.Fatalf("Expected to get '%+v' items, got '%+v'.", expect, got)
		}

		row0 := data[0].(map[string]interface{})
		if expect, got := "3", row0["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}

		row1 := data[1].(map[string]interface{})
		if expect, got := "4", row1["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}
	})
	t.Run("filter+offset+limit", func(t *testing.T) {
		data, err := getResult(map[string]string{
			"belongs_to_param": "1",
			"offset":           "1",
			"limit":            "2",
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect, got := 2, len(data); expect != got {
			t.Fatalf("Expected to get '%+v' items, got '%+v'.", expect, got)
		}

		row0 := data[0].(map[string]interface{})
		if expect, got := "3", row0["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}

		row1 := data[1].(map[string]interface{})
		if expect, got := "4", row1["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'.", expect, got)
		}
	})
}

func TestJsonArrayGetLimit(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonArray, err := mockJsonArrayForTests(&JsonArrayConfig{
			Input: "mock",
			Limit: JsonArrayLimitConfig{
				Default:   10,
				Max:       150,
				Parameter: "testlimit",
			},
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		limit, err := jsonArray.getLimit(map[string]string{
			"testlimit": "123",
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if got, expect := limit, uint(123); got != expect {
			t.Fatalf("Expected to get '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("max", func(t *testing.T) {
		jsonArray, err := mockJsonArrayForTests(&JsonArrayConfig{
			Input: "mock",
			Limit: JsonArrayLimitConfig{
				Default:   10,
				Max:       50,
				Parameter: "testlimit",
			},
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		limit, err := jsonArray.getLimit(map[string]string{
			"testlimit": "123",
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if got, expect := limit, uint(50); got != expect {
			t.Fatalf("Expected to get '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("default", func(t *testing.T) {
		jsonArray, err := mockJsonArrayForTests(&JsonArrayConfig{
			Input: "mock",
			Limit: JsonArrayLimitConfig{
				Default:   12,
				Max:       50,
				Parameter: "testlimit",
			},
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		limit, err := jsonArray.getLimit(map[string]string{})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if got, expect := limit, uint(12); got != expect {
			t.Fatalf("Expected to get '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("negative", func(t *testing.T) {
		jsonArray, err := mockJsonArrayForTests(&JsonArrayConfig{
			Input: "mock",
			Limit: JsonArrayLimitConfig{
				Default:   10,
				Max:       50,
				Parameter: "testlimit",
			},
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		_, err = jsonArray.getLimit(map[string]string{
			"testlimit": "-42",
		})
		if err == nil {
			t.Fatalf("Expected error, got nil.")
		}
	})
}

func TestJsonArrayGetOffset(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonArray, err := mockJsonArrayForTests(&JsonArrayConfig{
			Input: "mock",
			Offset: JsonArrayOffsetConfig{
				Parameter: "testoffset",
			},
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		offset, err := jsonArray.getOffset(map[string]string{
			"testoffset": "123",
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if got, expect := offset, uint(123); got != expect {
			t.Fatalf("Expected to get '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("negative", func(t *testing.T) {
		jsonArray, err := mockJsonArrayForTests(&JsonArrayConfig{
			Input: "mock",
			Offset: JsonArrayOffsetConfig{
				Parameter: "testoffset",
			},
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		_, err = jsonArray.getOffset(map[string]string{
			"testoffset": "-42",
		})
		if err == nil {
			t.Fatalf("Expected error, got nil.")
		}
	})
}

func TestJsonArrayGetFiltersPerIndex(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonArray, err := mockJsonArrayForTests(&JsonArrayConfig{
			Input: "mock",
			Parameters: map[string]parameterPackage.ParameterConfig{
				"a": {
					Property: "a",
					Index:    "a",
					Parser:   "mock",
				},
				"b": {
					Property: "b",
					Index:    "a",
					Parser:   "prefix",
				},
				"c": {
					Property: "c",
					Index:    "b",
					Parser:   "prefix",
				},
			},
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		filters, err := jsonArray.getFiltersPerIndex(map[string]string{
			"a": "val-a",
			"b": "val-b",
			"c": "val-c",
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if _, exists := filters["a"]; !exists {
			t.Fatalf("Expected to have filters for index 'a'")
		}
		if _, exists := filters["b"]; !exists {
			t.Fatalf("Expected to have filters for index 'b'")
		}

		if val, _ := filters["a"]; len(val) != 2 {
			t.Fatalf("Expected to have 2 filters for index 'a', got '%v'", len(val))
		}
		if val, _ := filters["b"]; len(val) != 1 {
			t.Fatalf("Expected to have 1 filter for index 'b', got '%v'", len(val))
		}

		if _, exists := filters["a"]["a"]; !exists {
			t.Fatalf("Expected to have a filter 'a' for index 'a'")
		}
		if _, exists := filters["a"]["b"]; !exists {
			t.Fatalf("Expected to have a filter 'b' for index 'a'")
		}
		if _, exists := filters["b"]["c"]; !exists {
			t.Fatalf("Expected to have a filter 'c' for index 'b'")
		}

		if expect, got := "val-a", filters["a"]["a"]; got != expect {
			t.Fatalf("Expected to have value '%v' for filter 'a' of index 'a', got '%v'", expect, got)
		}
		if expect, got := "prefix_val-b", filters["a"]["b"]; got != expect {
			t.Fatalf("Expected to have value '%v' for filter 'b' of index 'a', got '%v'", expect, got)
		}
		if expect, got := "prefix_val-c", filters["b"]["c"]; got != expect {
			t.Fatalf("Expected to have value '%v' for filter 'c' of index 'b', got '%v'", expect, got)
		}
	})
}
