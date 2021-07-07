package record

import (
	inputPackage "rodb.io/pkg/input"
	"testing"
)

func TestRecordListSort(t *testing.T) {
	// Only testing with strings because parser.Compare is already tested
	trueValue, falseValue := true, false
	records := List{
		inputPackage.NewStringPropertiesMockRecord(map[string]string{
			"a": "3",
			"b": "1",
		}, 0),
		inputPackage.NewStringPropertiesMockRecord(map[string]string{
			"a": "2",
			"b": "1",
		}, 1),
		inputPackage.NewStringPropertiesMockRecord(map[string]string{
			"a": "1",
			"b": "2",
		}, 2),
	}

	t.Run("normal", func(t *testing.T) {
		result := records.Sort([]*SortConfig{
			{
				Property:  "b",
				Ascending: &falseValue,
			}, {
				Property:  "a",
				Ascending: &trueValue,
			},
		})

		if expect, got := 3, len(result); got != expect {
			t.Fatalf("Expected length of '%v', got '%v'", expect, got)
		}

		for index, expect := range []int64{2, 1, 0} {
			if got := result[index].Position(); got != expect {
				t.Fatalf("Expected to get the record with position = '%v' at index '%v', got '%v'", expect, index, got)
			}
		}
	})
}
