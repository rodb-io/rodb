package output

import (
	"rods/pkg/config"
	"rods/pkg/index"
	"rods/pkg/input"
	"rods/pkg/parser"
	"rods/pkg/record"
	"testing"
)

type jsonDataForTests struct {
	mockResults      []input.IterateAllResult
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
	mockResults := []input.IterateAllResult{
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
	}

	mockInput := input.NewMock(mockResults)
	mockIndex := index.NewNoop(&config.NoopIndex{}, input.List{"mock": mockInput})
	mockIndex2 := index.NewNoop(&config.NoopIndex{}, input.List{"mock": mockInput})
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

		position0, err := recordLists[0]()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if expect, got := int64(1), *position0; got != expect {
			t.Errorf("Expected to get position '%+v' for the first result of the first index, got '%+v'", expect, got)
		}

		position1, err := recordLists[1]()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if expect, got := int64(1), *position1; got != expect {
			t.Errorf("Expected to get position '%+v' for the first result of the second index, got '%+v'", expect, got)
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
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := 1, len(recordIterators); got != expect {
			t.Errorf("Expected to get '%+v' record lists, got '%+v'", expect, got)
		}

		recordCount := 0
		for {
			position, err := recordIterators[0]()
			if err != nil {
				t.Errorf("Unexpected error: '%+v'", err)
			}
			if position == nil {
				break
			}
			recordCount++
		}

		if expect, got := len(jsonDataForTests.mockResults), recordCount; got != expect {
			t.Errorf("Expected to get '%+v' records in the first list, got '%+v'", expect, got)
		}
	})
}

func TestJsonObjectLoadRelationships(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jsonDataForTests := mockJsonDataForTests()

		ascendingSort := false
		relationshipsConfig := map[string]*config.Relationship{
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
