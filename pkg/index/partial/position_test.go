package partial

import (
	"testing"
)

func TestPositionLinkedListToIterator(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		list := &PositionLinkedList{
			Position: 1,
			NextPosition: &PositionLinkedList{
				Position: 42,
				NextPosition: &PositionLinkedList{
					Position:     123,
					NextPosition: nil,
				},
			},
		}
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

func TestPositionLinkedListCopy(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		list := &PositionLinkedList{
			Position: 1,
			NextPosition: &PositionLinkedList{
				Position: 2,
				NextPosition: &PositionLinkedList{
					Position:     3,
					NextPosition: nil,
				},
			},
		}
		copyFirst, copyLast := list.Copy()

		if got, expect := copyFirst.Position, list.Position; got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}
		if got, expect := copyFirst.NextPosition.Position, list.NextPosition.Position; got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}
		if got, expect := copyFirst.NextPosition.NextPosition.Position, list.NextPosition.NextPosition.Position; got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}
		if got, expect := copyLast, copyFirst.NextPosition.NextPosition; got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}
	})
}

