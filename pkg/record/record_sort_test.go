package record

import (
	"rods/pkg/config"
	"testing"
)

func TestRecordListSort(t *testing.T) {
	// Only testing with strings because parser.Compare is already tested
	trueValue, falseValue := true, false
	records := List{
		NewStringColumnsMock(map[string]string{
			"a": "3",
			"b": "1",
		}, 0),
		NewStringColumnsMock(map[string]string{
			"a": "2",
			"b": "1",
		}, 1),
		NewStringColumnsMock(map[string]string{
			"a": "1",
			"b": "2",
		}, 2),
	}

	t.Run("normal", func(t *testing.T) {
		result := records.Sort([]*config.Sort{
			{
				Column:    "b",
				Ascending: &falseValue,
			}, {
				Column:    "a",
				Ascending: &trueValue,
			},
		})

		if expect, got := 3, len(result); got != expect {
			t.Errorf("Expected length of '%v', got '%v'", expect, got)
		}

		for index, expect := range []int64{2, 1, 0} {
			if got := result[index].Position(); got != expect {
				t.Errorf("Expected to get the record with position = '%v' at index '%v', got '%v'", expect, index, got)
			}
		}
	})
}
