package index

import (
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/input"
	"rods/pkg/record"
	"testing"
)

func TestMemoryMapPrepare(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		index := NewMemoryMap(
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

		err := index.Prepare()
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
