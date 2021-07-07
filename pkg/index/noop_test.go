package index

import (
	"rodb.io/pkg/input"
	"rodb.io/pkg/input/record"
	"rodb.io/pkg/parser"
	"testing"
)

func TestNoopGetRecordPositions(t *testing.T) {
	mockInput := input.NewMock(parser.NewMock(), []record.Record{
		record.NewStringPropertiesMockRecord(map[string]string{
			"col":  "col_a",
			"col2": "col2_b",
		}, 0),
		record.NewStringPropertiesMockRecord(map[string]string{
			"col":  "col_a",
			"col2": "col2_a",
		}, 1),
		record.NewStringPropertiesMockRecord(map[string]string{
			"col":  "col_b",
			"col2": "col2_a",
		}, 2),
		record.NewStringPropertiesMockRecord(map[string]string{
			"col":  "col_a",
			"col2": "col2_a",
		}, 3),
		record.NewStringPropertiesMockRecord(map[string]string{
			"col":  "col_b",
			"col2": "col2_b",
		}, 4),
	})
	index := NewNoop(&NoopConfig{}, map[string]input.Input{
		"input": mockInput,
	})

	t.Run("normal", func(t *testing.T) {
		for _, testCase := range []struct {
			expectedLength  int
			expectedResults record.PositionList
		}{
			{
				expectedLength:  2,
				expectedResults: record.PositionList{1, 3},
			}, {
				expectedLength:  2,
				expectedResults: record.PositionList{1, 3},
			}, {
				expectedLength:  2,
				expectedResults: record.PositionList{1, 3},
			},
		} {
			nextPosition, err := index.GetRecordPositions(mockInput, map[string]interface{}{
				"col":  "col_a",
				"col2": "col2_a",
			})
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
	t.Run("wrong property", func(t *testing.T) {
		nextPosition, err := index.GetRecordPositions(mockInput, map[string]interface{}{
			"wrong_col": "",
		})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		_, err = nextPosition()
		if err == nil {
			t.Fatalf("Expected an error, got %v", err)
		}
	})
}
