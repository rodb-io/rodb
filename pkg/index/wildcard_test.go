package index

import (
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/index/wildcard"
	"rodb.io/pkg/input"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/input/record"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func stringifyWildcardTree(t *testing.T, root *wildcard.TreeNode) string {
	positionsToString := func(positions *wildcard.PositionLinkedList) string {
		result := ""
		currentPosition := positions
		var err error
		for currentPosition != nil {
			if currentPosition != positions {
				result += ","
			}
			result += strconv.Itoa(int(currentPosition.Position))

			currentPosition, err = currentPosition.NextPosition()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		}

		return result
	}

	results := make([]string, 0)

	var stringify func(prefix string, node *wildcard.TreeNode)
	stringify = func(prefix string, node *wildcard.TreeNode) {
		child, err := node.FirstChild()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		for child != nil {
			firstPosition, err := child.FirstPosition()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			value, err := child.Value()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			positions := positionsToString(firstPosition)
			currentValue := prefix + ">" + string(value)
			results = append(results, currentValue+"="+positions)
			stringify(currentValue, child)

			child, err = child.NextSibling()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		}
	}

	stringify("", root)
	sort.Strings(results)

	return strings.Join(results, "\n")
}

func TestWildcard(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		falseValue := false
		index, err := NewWildcard(
			&WildcardConfig{
				Properties: []string{"col"},
				Path:       "/tmp/test-index-wildcard-create.rodb",
				IgnoreCase: &falseValue,
				Input:      "input",
				Logger:     logrus.NewEntry(logrus.StandardLogger()),
			},
			input.List{
				"input": input.NewMock(parser.NewMock(), []record.Record{
					input.NewStringPropertiesMockRecord(map[string]string{
						"col": "BANANA",
					}, 1),
					input.NewStringPropertiesMockRecord(map[string]string{
						"col": "BANANO",
					}, 2),
					input.NewStringPropertiesMockRecord(map[string]string{
						"col": "PLANT",
					}, 3),
				}),
			},
		)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expect := strings.Join([]string{
			">A=1,2,3",
			">A>N=1,2,3",
			">A>N>A=1,2",
			">A>N>A>N=1,2",
			">A>N>A>N>A=1",
			">A>N>A>N>O=2",
			">A>N>O=2",
			">A>N>T=3",
			">BANAN=1,2",
			">BANAN>A=1",
			">BANAN>O=2",
			">LANT=3",
			">N=1,2,3",
			">N>A=1,2",
			">N>A>N=1,2",
			">N>A>N>A=1",
			">N>A>N>O=2",
			">N>O=2",
			">N>T=3",
			">O=2",
			">PLANT=3",
			">T=3",
		}, "\n")
		got := stringifyWildcardTree(t, index.index["col"])
		if got != expect {
			t.Fatalf("Unexpected list of results. Expected :\n=====\n%v\n=====\nbut got:\n=====\n%v\n", expect, got)
		}
	})
	t.Run("load", func(t *testing.T) {
		falseValue := false
		config := &WildcardConfig{
			Properties: []string{"col"},
			Path:       "/tmp/test-index-wildcard-load.rodb",
			IgnoreCase: &falseValue,
			Input:      "input",
			Logger:     logrus.NewEntry(logrus.StandardLogger()),
		}
		inputs := input.List{
			"input": input.NewMock(parser.NewMock(), []record.Record{
				input.NewStringPropertiesMockRecord(map[string]string{
					"col": "BANANA",
				}, 1),
				input.NewStringPropertiesMockRecord(map[string]string{
					"col": "BANANO",
				}, 2),
				input.NewStringPropertiesMockRecord(map[string]string{
					"col": "PLANT",
				}, 3),
			}),
		}

		// Creating the index file, then closing it
		indexToInitFile, err := NewWildcard(config, inputs)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if err := indexToInitFile.Close(); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Loading the index file
		index, err := NewWildcard(config, inputs)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expect := strings.Join([]string{
			">A=1,2,3",
			">A>N=1,2,3",
			">A>N>A=1,2",
			">A>N>A>N=1,2",
			">A>N>A>N>A=1",
			">A>N>A>N>O=2",
			">A>N>O=2",
			">A>N>T=3",
			">BANAN=1,2",
			">BANAN>A=1",
			">BANAN>O=2",
			">LANT=3",
			">N=1,2,3",
			">N>A=1,2",
			">N>A>N=1,2",
			">N>A>N>A=1",
			">N>A>N>O=2",
			">N>O=2",
			">N>T=3",
			">O=2",
			">PLANT=3",
			">T=3",
		}, "\n")
		got := stringifyWildcardTree(t, index.index["col"])
		if got != expect {
			t.Fatalf("Unexpected list of results. Expected :\n=====\n%v\n=====\nbut got:\n=====\n%v\n", expect, got)
		}
	})
}

func TestWildcardGetRecordPositions(t *testing.T) {
	createTestData := func(t *testing.T, testName string) (*input.Mock, *Wildcard) {
		mockInput := input.NewMock(parser.NewMock(), []record.Record{
			input.NewStringPropertiesMockRecord(map[string]string{
				"col":  "BANANA",
				"col2": "col2_b",
			}, 0),
			input.NewStringPropertiesMockRecord(map[string]string{
				"col":  "BANANT",
				"col2": "col2_a",
			}, 1),
			input.NewStringPropertiesMockRecord(map[string]string{
				"col":  "PLANT",
				"col2": "col2_a",
			}, 2),
			input.NewStringPropertiesMockRecord(map[string]string{
				"col":  "BANANA",
				"col2": "col2_a",
			}, 3),
			input.NewStringPropertiesMockRecord(map[string]string{
				"col":  "PLANT",
				"col2": "col2_b",
			}, 4),
		})
		falseValue := false
		index, err := NewWildcard(
			&WildcardConfig{
				Properties: []string{"col", "col2"},
				Path:       "/tmp/test-index-wildcard-get-record-positions" + testName + ".rodb",
				IgnoreCase: &falseValue,
				Input:      "input",
				Logger:     logrus.NewEntry(logrus.StandardLogger()),
			},
			input.List{
				"input": mockInput,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		return mockInput, index
	}

	t.Run("normal", func(t *testing.T) {
		mockInput, index := createTestData(t, "normal")
		for _, testCase := range []struct {
			expectedLength  int
			expectedResults record.PositionList
			filters         map[string]interface{}
		}{
			{
				expectedLength:  2,
				expectedResults: record.PositionList{1, 3},
				filters: map[string]interface{}{
					"col":  "BANAN",
					"col2": "col2_a",
				},
			}, {
				expectedLength:  1,
				expectedResults: record.PositionList{2},
				filters: map[string]interface{}{
					"col":  "PLANT",
					"col2": "col2_a",
				},
			}, {
				expectedLength:  2,
				expectedResults: record.PositionList{0, 4},
				filters: map[string]interface{}{
					"col2": "col2_b",
				},
			}, {
				expectedLength:  3,
				expectedResults: record.PositionList{1, 2, 4},
				filters: map[string]interface{}{
					"col": "ANT",
				},
			},
		} {
			nextPosition, err := index.GetRecordPositions(mockInput, testCase.filters)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			positions := make([]record.Position, 0)
			for {
				position, err := nextPosition()
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if position == nil {
					break
				}
				positions = append(positions, *position)
			}

			if got, expect := len(positions), testCase.expectedLength; got != expect {
				t.Fatalf("Expected %v positions, got %v, testCase: %+v", expect, got, testCase)
			}

			for i, position := range testCase.expectedResults {
				if position != positions[i] {
					t.Fatalf("Expected position %v at index %v, got %v", position, i, positions[i])
				}
			}
		}
	})
	t.Run("no filters", func(t *testing.T) {
		mockInput, index := createTestData(t, "no-filters")
		_, err := index.GetRecordPositions(mockInput, map[string]interface{}{})
		if err == nil {
			t.Fatalf("Expected an error, got %v", err)
		}
	})
	t.Run("wrong property", func(t *testing.T) {
		mockInput, index := createTestData(t, "wrong-property")
		_, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"wrong_col": "",
		})
		if err == nil {
			t.Fatalf("Expected an error, got %v", err)
		}
	})

	createTestDataForIgnoreCase := func(t *testing.T, ignoreCase bool, testName string) (*input.Mock, *Wildcard) {
		mockInput := input.NewMock(parser.NewMock(), []record.Record{
			input.NewStringPropertiesMockRecord(map[string]string{
				"col": "BANANÉ",
			}, 42),
		})
		index, err := NewWildcard(
			&WildcardConfig{
				Properties: []string{"col"},
				Path:       "/tmp/test-index-wildcard-get-record-positions" + testName + ".rodb",
				Input:      "input",
				IgnoreCase: &ignoreCase,
				Logger:     logrus.NewEntry(logrus.StandardLogger()),
			},
			input.List{
				"input": mockInput,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		return mockInput, index
	}

	t.Run("IgnoreCase is true, search lower case", func(t *testing.T) {
		mockInput, index := createTestDataForIgnoreCase(t, true, "ignorecase-true-search-lower")
		iterator, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"col": "ané",
		})
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		recordPosition, err := iterator()
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		if recordPosition == nil {
			t.Fatalf("Expected to get one record, got %v", recordPosition)
		}
		if *recordPosition != 42 {
			t.Fatalf("Expected to get record 42, got %v", recordPosition)
		}
	})
	t.Run("IgnoreCase is true, search upper case", func(t *testing.T) {
		mockInput, index := createTestDataForIgnoreCase(t, true, "ignorecase-true-search-upper")
		iterator, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"col": "ANÉ",
		})
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		recordPosition, err := iterator()
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}
		if recordPosition == nil {
			t.Fatalf("Expected to get one record, got %v", recordPosition)
		}
		if *recordPosition != 42 {
			t.Fatalf("Expected to get record 42, got %v", recordPosition)
		}
	})
	t.Run("IgnoreCase is false, search lower case", func(t *testing.T) {
		mockInput, index := createTestDataForIgnoreCase(t, false, "ignorecase-false-search-lower")

		iterator, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"col": "ané",
		})
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}

		recordPosition, err := iterator()
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}

		if recordPosition != nil {
			t.Fatalf("Expected not to get a record, got %v", recordPosition)
		}
	})
	t.Run("IgnoreCase is false, search upper case", func(t *testing.T) {
		mockInput, index := createTestDataForIgnoreCase(t, false, "ignorecase-false-search-upper")

		iterator, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"col": "ANÉ",
		})
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}

		recordPosition, err := iterator()
		if err != nil {
			t.Fatalf("Unexpected error, got %v", err)
		}

		if recordPosition == nil {
			t.Fatalf("Expected to get one record, got %v", recordPosition)
		}
		if *recordPosition != 42 {
			t.Fatalf("Expected to get record 42, got %v", recordPosition)
		}
	})
}
