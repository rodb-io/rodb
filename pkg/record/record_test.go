package record

import (
	"testing"
)

func TestPositionListToIterator(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		list := PositionList{1, 42, 123}
		iterator := list.Iterate()

		list0, err := iterator()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if got, expect := list0, int64(1); got == nil || *got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}

		list1, err := iterator()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if got, expect := list1, int64(42); got == nil || *got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}

		list2, err := iterator()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if got, expect := list2, int64(123); got == nil || *got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}

		list3, err := iterator()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if list3 != nil {
			t.Errorf("Expected nil, got %v", list3)
		}
	})
}
