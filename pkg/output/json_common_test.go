package output

import (
	"rodb.io/pkg/config"
	"rodb.io/pkg/index"
	"rodb.io/pkg/input"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/input/record"
	"testing"
)

type jsonDataForTests struct {
	mockResults      []record.Record
	mockInput        *input.Mock
	mockIndex        *index.Noop
	mockIndex2       *index.Noop
	mockParser       *parser.Mock
	mockParserPrefix *parser.Mock
	inputs           input.List
	indexes          index.List
	parsers          parser.List
}

func mockJsonDataForTests() jsonDataForTests {
	mockResults := []record.Record{
		input.NewStringPropertiesMockRecord(map[string]string{
			"id":         "1",
			"belongs_to": "0",
		}, 0),
		input.NewStringPropertiesMockRecord(map[string]string{
			"id":         "2",
			"belongs_to": "1",
		}, 1),
		input.NewStringPropertiesMockRecord(map[string]string{
			"id":         "3",
			"belongs_to": "1",
		}, 2),
		input.NewStringPropertiesMockRecord(map[string]string{
			"id":         "4",
			"belongs_to": "1",
		}, 3),
	}

	mockInput := input.NewMock(parser.NewMock(), mockResults)
	mockIndex := index.NewNoop(&index.NoopConfig{}, input.List{"mock": mockInput})
	mockIndex2 := index.NewNoop(&index.NoopConfig{}, input.List{"mock": mockInput})
	mockParser := parser.NewMock()
	mockParserPrefix := parser.NewMockWithPrefix("prefix_")

	inputs := input.List{"mock": mockInput}
	indexes := index.List{"default": mockIndex, "mock": mockIndex, "mock2": mockIndex2}
	parsers := parser.List{"mock": mockParser, "prefix": mockParserPrefix}

	return jsonDataForTests{
		mockResults,
		mockInput,
		mockIndex,
		mockIndex2,
		mockParser,
		mockParserPrefix,
		inputs,
		indexes,
		parsers,
	}
}

func TestJsonObjectGetRelationshipFiltersPerIndex(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		filtersPerIndex, err := getRelationshipFiltersPerIndex(
			map[string]interface{}{
				"foo": "3",
				"bar": "1",
			},
			[]*config.RelationshipMatch{
				{
					ParentProperty: "foo",
					ChildProperty:  "foo",
					ChildIndex:     "a",
				}, {
					ParentProperty: "foo",
					ChildProperty:  "foo",
					ChildIndex:     "b",
				}, {
					ParentProperty: "bar",
					ChildProperty:  "bar",
					ChildIndex:     "b",
				},
			},
			"test",
		)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if got, exists := filtersPerIndex["a"]; !exists {
			t.Fatalf("Expected to get filters for index 'a', got '%+v'", got)
		}
		if got, exists := filtersPerIndex["b"]; !exists {
			t.Fatalf("Expected to get filters for index 'b', got '%+v'", got)
		}

		if got, expect := len(filtersPerIndex["a"]), 1; got != expect {
			t.Fatalf("Expected to get '%+v' filters for index 'a', got '%+v'", expect, got)
		}
		if got, expect := len(filtersPerIndex["b"]), 2; got != expect {
			t.Fatalf("Expected to get '%+v' filters for index 'b', got '%+v'", expect, got)
		}

		if got, exists := filtersPerIndex["a"]["foo"]; !exists || got != "3" {
			t.Fatalf("Expected to get '%+v' value for filter, got '%+v'", "3", got)
		}
		if got, exists := filtersPerIndex["a"]["foo"]; !exists || got != "3" {
			t.Fatalf("Expected to get '%+v' value for filter, got '%+v'", "3", got)
		}
		if got, exists := filtersPerIndex["b"]["bar"]; !exists || got != "1" {
			t.Fatalf("Expected to get '%+v' value for filter, got '%+v'", "1", got)
		}
	})
}

func TestJsonObjectGetFilteredRecordPositionsPerIndex(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonDataForTests := mockJsonDataForTests()

		filtersPerIndex := map[string]map[string]interface{}{
			"mock": {
				"id": "2",
			},
			"mock2": {
				"belongs_to": "1",
			},
		}

		recordLists, err := getFilteredRecordPositionsPerIndex(
			jsonDataForTests.indexes["default"],
			jsonDataForTests.indexes,
			jsonDataForTests.mockInput,
			filtersPerIndex,
		)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := 2, len(recordLists); got != expect {
			t.Fatalf("Expected to get '%+v' entries in the array, got '%+v'", expect, got)
		}

		// Not working, because the map does not guarantee the order
		// if expect, got := 1, len(recordLists[0]); got != expect {
		// 	t.Fatalf("Expected to get '%+v' entries in the first array, got '%+v'", expect, got)
		// }
		// if expect, got := 3, len(recordLists[1]); got != expect {
		// 	t.Fatalf("Expected to get '%+v' entries in the second array, got '%+v'", expect, got)
		// }

		position0, err := recordLists[0]()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect, got := int64(1), *position0; got != expect {
			t.Fatalf("Expected to get position '%+v' for the first result of the first index, got '%+v'", expect, got)
		}

		position1, err := recordLists[1]()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect, got := int64(1), *position1; got != expect {
			t.Fatalf("Expected to get position '%+v' for the first result of the second index, got '%+v'", expect, got)
		}
	})
	t.Run("no filters", func(t *testing.T) {
		jsonDataForTests := mockJsonDataForTests()

		recordIterators, err := getFilteredRecordPositionsPerIndex(
			jsonDataForTests.indexes["default"],
			jsonDataForTests.indexes,
			jsonDataForTests.mockInput,
			map[string]map[string]interface{}{},
		)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := 1, len(recordIterators); got != expect {
			t.Fatalf("Expected to get '%+v' record lists, got '%+v'", expect, got)
		}

		recordCount := 0
		for {
			position, err := recordIterators[0]()
			if err != nil {
				t.Fatalf("Unexpected error: '%+v'", err)
			}
			if position == nil {
				break
			}
			recordCount++
		}

		if expect, got := len(jsonDataForTests.mockResults), recordCount; got != expect {
			t.Fatalf("Expected to get '%+v' records in the first list, got '%+v'", expect, got)
		}
	})
}

func TestJsonObjectLoadRelationships(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonDataForTests := mockJsonDataForTests()

		falseValue := false
		relationshipsConfig := map[string]*config.Relationship{
			"children": {
				Input:   "mock",
				IsArray: true,
				Limit:   2,
				Sort: []*config.Sort{
					{
						Property:  "id",
						Ascending: &falseValue,
					},
				},
				Match: []*config.RelationshipMatch{
					{
						ParentProperty: "id",
						ChildProperty:  "belongs_to",
						ChildIndex:     "mock",
					},
				},
				Relationships: map[string]*config.Relationship{
					"subchild": {
						Input:   "mock",
						IsArray: false,
						Match: []*config.RelationshipMatch{
							{
								ParentProperty: "belongs_to",
								ChildProperty:  "id",
								ChildIndex:     "mock",
							},
						},
					},
				},
			},
		}

		data := map[string]interface{}{
			"id": "1",
		}
		data, err := loadRelationships(
			data,
			relationshipsConfig,
			jsonDataForTests.indexes["default"],
			jsonDataForTests.indexes,
			jsonDataForTests.inputs,
			"mock",
		)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := "1", data["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'", expect, got)
		}

		if got, ok := data["children"]; !ok {
			t.Fatalf("Expected to get an array, got '%+v'", got)
		}
		if got, ok := data["children"].([]map[string]interface{}); !ok {
			t.Fatalf("Expected to get an array, got '%+v'", got)
		}

		children := data["children"].([]map[string]interface{})
		if expect, got := 2, len(children); expect != got {
			t.Fatalf("Expected length of '%+v', got '%+v'", expect, got)
		}

		// The sort result is only quickly tested, because record.List.Sort is already tested
		if expect, got := "4", children[0]["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'", expect, got)
		}
		if expect, got := "3", children[1]["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'", expect, got)
		}

		if got, ok := children[0]["subchild"]; !ok {
			t.Fatalf("Expected to get an object, got '%+v'", got)
		}
		if got, ok := children[1]["subchild"]; !ok {
			t.Fatalf("Expected to get an object, got '%+v'", got)
		}

		if got, ok := children[0]["subchild"].(map[string]interface{}); !ok {
			t.Fatalf("Expected to get an object, got '%+v'", got)
		}
		if got, ok := children[1]["subchild"].(map[string]interface{}); !ok {
			t.Fatalf("Expected to get an object, got '%+v'", got)
		}

		subchild0 := children[0]["subchild"].(map[string]interface{})
		subchild1 := children[1]["subchild"].(map[string]interface{})

		if expect, got := "1", subchild0["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'", expect, got)
		}
		if expect, got := "1", subchild1["id"]; expect != got {
			t.Fatalf("Expected to get '%+v', got '%+v'", expect, got)
		}
	})
}
