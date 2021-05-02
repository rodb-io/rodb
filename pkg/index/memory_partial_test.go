package index

import (
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/config"
	"rodb.io/pkg/input"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/record"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func stringifyMemoryPartialTree(root *partialIndexTreeNode) string {
	positionsToString := func(positions *record.PositionLinkedList) string {
		result := ""
		for currentPosition := positions; currentPosition != nil; currentPosition = currentPosition.NextPosition {
			if currentPosition != positions {
				result += ","
			}
			result += strconv.Itoa(int(currentPosition.Position))
		}

		return result
	}

	results := make([]string, 0)

	var stringify func(prefix string, node *partialIndexTreeNode)
	stringify = func(prefix string, node *partialIndexTreeNode) {
		for child := node.firstChild; child != nil; child = child.nextSibling {
			positions := positionsToString(child.firstPosition)
			currentValue := prefix + string(child.value)
			results = append(results, currentValue+"="+positions)
			stringify(currentValue, child)
		}
	}

	stringify("", root)
	sort.Strings(results)

	return strings.Join(results, "\n")
}

func TestMemoryPartial(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		index, err := NewMemoryPartial(
			&config.MemoryPartialIndex{
				Properties: []string{"col"},
				Input:      "input",
				Logger:     logrus.NewEntry(logrus.StandardLogger()),
			},
			input.List{
				"input": input.NewMock(parser.NewMock(), []record.Record{
					record.NewStringPropertiesMock(map[string]string{
						"col": "BANANA",
					}, 1),
					record.NewStringPropertiesMock(map[string]string{
						"col": "BANANO",
					}, 2),
					record.NewStringPropertiesMock(map[string]string{
						"col": "PLANT",
					}, 3),
				}),
			},
		)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expect := strings.Join([]string{
			"A=1,2,3",
			"AN=1,2,3",
			"ANA=1,2",
			"ANAN=1,2",
			"ANANA=1",
			"ANANO=2",
			"ANO=2",
			"ANT=3",
			"B=1,2",
			"BA=1,2",
			"BAN=1,2",
			"BANA=1,2",
			"BANAN=1,2",
			"BANANA=1",
			"BANANO=2",
			"L=3",
			"LA=3",
			"LAN=3",
			"LANT=3",
			"N=1,2,3",
			"NA=1,2",
			"NAN=1,2",
			"NANA=1",
			"NANO=2",
			"NO=2",
			"NT=3",
			"P=3",
			"PL=3",
			"PLA=3",
			"PLAN=3",
			"PLANT=3",
		}, "\n")
		got := stringifyMemoryPartialTree(index.index["col"])
		if got != expect {
			t.Errorf("Unexpected list of results. Expected :\n=====\n%v\n=====\nbut got:\n=====\n%v\n", expect, got)
		}
	})
}

func TestMemoryPartialGetRecordPositions(t *testing.T) {
	createTestData := func() (*input.Mock, *MemoryPartial) {
		mockInput := input.NewMock(parser.NewMock(), []record.Record{
			record.NewStringPropertiesMock(map[string]string{
				"col":  "BANANA",
				"col2": "col2_b",
			}, 0),
			record.NewStringPropertiesMock(map[string]string{
				"col":  "BANANT",
				"col2": "col2_a",
			}, 1),
			record.NewStringPropertiesMock(map[string]string{
				"col":  "PLANT",
				"col2": "col2_a",
			}, 2),
			record.NewStringPropertiesMock(map[string]string{
				"col":  "BANANA",
				"col2": "col2_a",
			}, 3),
			record.NewStringPropertiesMock(map[string]string{
				"col":  "PLANT",
				"col2": "col2_b",
			}, 4),
		})
		index, err := NewMemoryPartial(
			&config.MemoryPartialIndex{
				Properties: []string{"col", "col2"},
				Input:      "input",
				Logger:     logrus.NewEntry(logrus.StandardLogger()),
			},
			input.List{
				"input": mockInput,
			},
		)
		if err != nil {
			t.Error(err)
		}

		return mockInput, index
	}

	t.Run("normal", func(t *testing.T) {
		mockInput, index := createTestData()
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
				t.Errorf("Expected no error, got %v", err)
			}

			positions := make([]record.Position, 0)
			for {
				position, err := nextPosition()
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if position == nil {
					break
				}
				positions = append(positions, *position)
			}

			if got, expect := len(positions), testCase.expectedLength; got != expect {
				t.Errorf("Expected %v positions, got %v, testCase: %+v", expect, got, testCase)
			}

			for i, position := range testCase.expectedResults {
				if position != positions[i] {
					t.Errorf("Expected position %v at index %v, got %v", position, i, positions[i])
				}
			}
		}
	})
	t.Run("no filters", func(t *testing.T) {
		mockInput, index := createTestData()
		_, err := index.GetRecordPositions(mockInput, map[string]interface{}{})
		if err == nil {
			t.Errorf("Expected an error, got %v", err)
		}
	})
	t.Run("wrong property", func(t *testing.T) {
		mockInput, index := createTestData()
		_, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"wrong_col": "",
		})
		if err == nil {
			t.Errorf("Expected an error, got %v", err)
		}
	})

	createTestDataForIgnoreCase := func(ignoreCase bool) (*input.Mock, *MemoryPartial) {
		mockInput := input.NewMock(parser.NewMock(), []record.Record{
			record.NewStringPropertiesMock(map[string]string{
				"col": "BANANÉ",
			}, 42),
		})
		index, err := NewMemoryPartial(
			&config.MemoryPartialIndex{
				Properties: []string{"col"},
				Input:      "input",
				IgnoreCase: &ignoreCase,
				Logger:     logrus.NewEntry(logrus.StandardLogger()),
			},
			input.List{
				"input": mockInput,
			},
		)
		if err != nil {
			t.Error(err)
		}

		return mockInput, index
	}

	t.Run("IgnoreCase is true, search lower case", func(t *testing.T) {
		mockInput, index := createTestDataForIgnoreCase(true)
		iterator, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"col": "ané",
		})
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		recordPosition, err := iterator()
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		if recordPosition == nil {
			t.Errorf("Expected to get one record, got %v", recordPosition)
		}
		if *recordPosition != 42 {
			t.Errorf("Expected to get record 42, got %v", recordPosition)
		}
	})
	t.Run("IgnoreCase is true, search upper case", func(t *testing.T) {
		mockInput, index := createTestDataForIgnoreCase(true)
		iterator, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"col": "ANÉ",
		})
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		recordPosition, err := iterator()
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}
		if recordPosition == nil {
			t.Errorf("Expected to get one record, got %v", recordPosition)
		}
		if *recordPosition != 42 {
			t.Errorf("Expected to get record 42, got %v", recordPosition)
		}
	})
	t.Run("IgnoreCase is false, search lower case", func(t *testing.T) {
		mockInput, index := createTestDataForIgnoreCase(false)

		iterator, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"col": "ané",
		})
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}

		recordPosition, err := iterator()
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}

		if recordPosition != nil {
			t.Errorf("Expected not to get a record, got %v", recordPosition)
		}
	})
	t.Run("IgnoreCase is false, search upper case", func(t *testing.T) {
		mockInput, index := createTestDataForIgnoreCase(false)

		iterator, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"col": "ANÉ",
		})
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}

		recordPosition, err := iterator()
		if err != nil {
			t.Errorf("Unexpected error, got %v", err)
		}

		if recordPosition == nil {
			t.Errorf("Expected to get one record, got %v", recordPosition)
		}
		if *recordPosition != 42 {
			t.Errorf("Expected to get record 42, got %v", recordPosition)
		}
	})
}
