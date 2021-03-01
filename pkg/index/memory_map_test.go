package index

import (
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/input"
	"rods/pkg/record"
	"testing"
)

func TestMemoryMap(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		index, err := NewMemoryMap(
			&config.MemoryMapIndex{
				Columns: []string{"col", "col2"},
				Input:   "input",
			},
			input.NewMock([]input.IterateAllResult{
				{Record: record.NewStringColumnsMock(map[string]string{
					"col":  "value_a",
					"col2": "value_2",
				}, 0)},
				{Record: record.NewStringColumnsMock(map[string]string{
					"col":  "value_b",
					"col2": "value_2",
				}, 0)},
				{Record: record.NewStringColumnsMock(map[string]string{
					"col":  "value_b",
					"col2": "value_2",
				}, 1)},
			}),
			logrus.StandardLogger(),
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

func TestMemoryMapGetRecords(t *testing.T) {
	index, err := NewMemoryMap(
		&config.MemoryMapIndex{
			Columns: []string{"col", "col2"},
			Input:   "input",
		},
		input.NewMock([]input.IterateAllResult{
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
		}),
		logrus.StandardLogger(),
	)
	if err != nil {
		t.Error(err)
	}

	t.Run("normal", func(t *testing.T) {
		for _, testCase := range []struct {
			limit           uint
			expectedLength  int
			expectedResults []record.Position
		}{
			{
				limit:           0,
				expectedLength:  2,
				expectedResults: []record.Position{1, 3},
			},
			{
				limit:           1,
				expectedLength:  1,
				expectedResults: []record.Position{1},
			},
			{
				limit:           2,
				expectedLength:  2,
				expectedResults: []record.Position{1, 3},
			},
			{
				limit:           10,
				expectedLength:  2,
				expectedResults: []record.Position{1, 3},
			},
		} {
			records, err := index.GetRecords("input", map[string]interface{}{
				"col":  "col_a",
				"col2": "col2_a",
			}, testCase.limit)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if got, expect := len(records), testCase.expectedLength; got != expect {
				t.Errorf("Expected %v records, got %v, testCase: %+v", expect, got, testCase)
			}

			for i, position := range testCase.expectedResults {
				if position != records[i].Position() {
					t.Errorf("Expected position %v at index %v, got %v", position, i, records[i])
				}
			}
		}
	})
	t.Run("wrong input", func(t *testing.T) {
		_, err := index.GetRecords("wrong_input", map[string]interface{}{
			"col": "",
		}, 1)
		if err == nil {
			t.Errorf("Expected an error, got %v", err)
		}
	})
	t.Run("no filters", func(t *testing.T) {
		_, err := index.GetRecords("input", map[string]interface{}{}, 1)
		if err == nil {
			t.Errorf("Expected an error, got %v", err)
		}
	})
	t.Run("wrong column", func(t *testing.T) {
		_, err := index.GetRecords("input", map[string]interface{}{
			"wrong_col": "",
		}, 1)
		if err == nil {
			t.Errorf("Expected an error, got %v", err)
		}
	})
}
