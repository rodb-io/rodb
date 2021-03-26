package record

import (
	"testing"
)

func TestJoinPositionIterators(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		lists  []PositionList
		expect PositionList
	}{
		{
			name: "all elements from the first one",
			lists: []PositionList{
				{2},
				{2, 3},
				{1, 2},
				{1, 2, 4},
			},
			expect: PositionList{2},
		}, {
			name: "few elements from the first one",
			lists: []PositionList{
				{0, 1, 2, 3, 4, 5},
				{3, 4, 5},
				{0, 3, 5, 6, 7},
			},
			expect: PositionList{3, 5},
		}, {
			name: "single list",
			lists: []PositionList{
				{42, 123},
			},
			expect: PositionList{42, 123},
		}, {
			name:   "no lists",
			lists:  []PositionList{},
			expect: PositionList{},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			lists := make([]PositionIterator, 0)
			for _, values := range testCase.lists {
				list := (func(values PositionList) PositionIterator {
					i := 0
					return func() (*Position, error) {
						for i < len(values) {
							currentIndex := i
							i++
							return &values[currentIndex], nil
						}

						return nil, nil
					}
				})(values)

				lists = append(lists, list)
			}

			resultIterator := JoinPositionIterators(lists...)

			result := make([]Position, 0)
			for {
				position, err := resultIterator()
				if err != nil {
					t.Errorf("Unexpected error: '%+v'", err)
				}
				if position == nil {
					break
				}
				result = append(result, *position)
			}

			if expect, got := len(testCase.expect), len(result); got != expect {
				t.Errorf("Expected length of '%v', got '%v'", expect, got)
			}

			for i, expect := range testCase.expect {
				if got := result[i]; got != expect {
					t.Errorf("Expected value of '%v' at index '%v', got '%v'", expect, i, got)
				}
			}
		})
	}
}
