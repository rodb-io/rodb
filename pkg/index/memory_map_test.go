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
				Columns: []string{"col"},
				Input:   "input",
			},
			input.NewMock([]input.IterateAllResult{
				{Record: record.NewSingleStringColumnMock("col", "value_a", 0)},
				{Record: record.NewSingleStringColumnMock("col", "value_b", 0)},
				{Record: record.NewSingleStringColumnMock("col", "value_b", 1)},
			}),
			logrus.StandardLogger(),
		)

		err := index.Prepare()
		if err != nil {
			t.Error(err)
		}

		for key, expected := range map[string]int{
			"value_a": 1,
			"value_b": 2,
		} {
			if got := len(index.index["col"][key]); got != expected {
				t.Errorf("Expected to have %v indexed value for '%v', got %v", expected, key, got)
			}
		}

		for key, expected := range map[string][]int64{
			"value_a": {0},
			"value_b": {0, 1},
		} {
			for indexOfExpectation, expectedPosition := range expected {
				if got := index.index["col"][key][indexOfExpectation]; got != expectedPosition {
					t.Errorf("Expected to have position %v indexed for value '%v'[%v], got %v", expectedPosition, key, indexOfExpectation, got)
				}
			}
		}
	})
}
