package index

import (
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/config"
	"rodb.io/pkg/input"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/record"
	"testing"
)

func TestMap(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		index, err := NewMap(
			&config.MapIndex{
				Properties: []string{"col", "col2"},
				Input:      "input",
				Logger:     logrus.NewEntry(logrus.StandardLogger()),
			},
			input.List{
				"input": input.NewMock(parser.NewMock(), []record.Record{
					record.NewStringPropertiesMock(map[string]string{
						"col":  "value_a",
						"col2": "value_2",
					}, 0),
					record.NewStringPropertiesMock(map[string]string{
						"col":  "value_b",
						"col2": "value_2",
					}, 0),
					record.NewStringPropertiesMock(map[string]string{
						"col":  "value_b",
						"col2": "value_2",
					}, 1),
				}),
			},
		)
		if err != nil {
			t.Error(err)
		}

		for col, expectedValues := range map[string]map[string]int{
			"col": {
				"value_a": 1,
				"value_b": 2,
			},
			"col2": {
				"value_2": 3,
			},
		} {
			for key, expected := range expectedValues {
				if got := len(index.index[col][key]); got != expected {
					t.Errorf("Expected to have %v indexed value for '%v', got %v", expected, key, got)
				}
			}
		}

		for col, expectedPositions := range map[string]map[string][]int64{
			"col": {
				"value_a": {0},
				"value_b": {0, 1},
			},
			"col2": {
				"value_2": {0, 0, 1},
			},
		} {
			for key, expected := range expectedPositions {
				for indexOfExpectation, expectedPosition := range expected {
					if got := index.index[col][key][indexOfExpectation]; got != expectedPosition {
						t.Errorf("Expected to have position %v indexed for value '%v'[%v], got %v", expectedPosition, key, indexOfExpectation, got)
					}
				}
			}
		}
	})
}

func TestMapGetRecordPositions(t *testing.T) {
	mockInput := input.NewMock(parser.NewMock(), []record.Record{
		record.NewStringPropertiesMock(map[string]string{
			"col":  "col_a",
			"col2": "col2_b",
		}, 0),
		record.NewStringPropertiesMock(map[string]string{
			"col":  "col_a",
			"col2": "col2_a",
		}, 1),
		record.NewStringPropertiesMock(map[string]string{
			"col":  "col_b",
			"col2": "col2_a",
		}, 2),
		record.NewStringPropertiesMock(map[string]string{
			"col":  "col_a",
			"col2": "col2_a",
		}, 3),
		record.NewStringPropertiesMock(map[string]string{
			"col":  "col_b",
			"col2": "col2_b",
		}, 4),
	})
	index, err := NewMap(
		&config.MapIndex{
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

	t.Run("normal", func(t *testing.T) {
		for _, testCase := range []struct {
			expectedLength  int
			expectedResults record.PositionList
			filters         map[string]interface{}
		}{
			{
				expectedLength:  2,
				expectedResults: record.PositionList{1, 3},
				filters: map[string]interface{}{
					"col":  "col_a",
					"col2": "col2_a",
				},
			}, {
				expectedLength:  1,
				expectedResults: record.PositionList{2},
				filters: map[string]interface{}{
					"col":  "col_b",
					"col2": "col2_a",
				},
			}, {
				expectedLength:  2,
				expectedResults: record.PositionList{0, 4},
				filters: map[string]interface{}{
					"col2": "col2_b",
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
		_, err := index.GetRecordPositions(mockInput, map[string]interface{}{})
		if err == nil {
			t.Errorf("Expected an error, got %v", err)
		}
	})
	t.Run("wrong property", func(t *testing.T) {
		_, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"wrong_col": "",
		})
		if err == nil {
			t.Errorf("Expected an error, got %v", err)
		}
	})
}