package index

import (
	"rods/pkg/input"
	"rods/pkg/record"
	"testing"
)

func TestNoopGetRecordPositions(t *testing.T) {
	mockInput := input.NewMock([]input.IterateAllResult{
		{Record: record.NewStringColumnsMock(map[string]string{
			"col":  "col_a",
			"col2": "col2_b",
		}, 0)},
		{Record: record.NewStringColumnsMock(map[string]string{
			"col":  "col_a",
			"col2": "col2_a",
		}, 1)},
		{Record: record.NewStringColumnsMock(map[string]string{
			"col":  "col_b",
			"col2": "col2_a",
		}, 2)},
		{Record: record.NewStringColumnsMock(map[string]string{
			"col":  "col_a",
			"col2": "col2_a",
		}, 3)},
		{Record: record.NewStringColumnsMock(map[string]string{
			"col":  "col_b",
			"col2": "col2_b",
		}, 4)},
	})
	index := NewNoop(
		map[string]input.Input{
			"input": mockInput,
		},
	)

	t.Run("normal", func(t *testing.T) {
		for _, testCase := range []struct {
			limit           uint
			expectedLength  int
			expectedResults record.PositionList
		}{
			{
				limit:           0,
				expectedLength:  2,
				expectedResults: record.PositionList{1, 3},
			},
			{
				limit:           1,
				expectedLength:  1,
				expectedResults: record.PositionList{1},
			},
			{
				limit:           2,
				expectedLength:  2,
				expectedResults: record.PositionList{1, 3},
			},
			{
				limit:           10,
				expectedLength:  2,
				expectedResults: record.PositionList{1, 3},
			},
		} {
			positions, err := index.GetRecordPositions(mockInput, map[string]interface{}{
				"col":  "col_a",
				"col2": "col2_a",
			}, testCase.limit)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
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
	t.Run("wrong column", func(t *testing.T) {
		_, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"wrong_col": "",
		}, 1)
		if err == nil {
			t.Errorf("Expected an error, got %v", err)
		}
	})
}
