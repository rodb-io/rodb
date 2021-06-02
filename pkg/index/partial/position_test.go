package partial

import (
	"io/ioutil"
	"rodb.io/pkg/record"
	"testing"
)

func createTestPositionLinkedList(t *testing.T, positions []record.Position) *PositionLinkedList {
	if len(positions) == 0 {
		return nil
	}

	file, err := ioutil.TempFile("/tmp", "test-index")
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}
	stream := NewStream(file, 0)

	// Dummy byte to avoid issues with the offset 0
	_, err = stream.Add([]byte{0})
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}

	list, err := NewPositionLinkedListFromArray(stream, positions)
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}

	return list
}

func TestPositionLinkedListToIterator(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		list := createTestPositionLinkedList(t, []record.Position{1, 42, 123})
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
		list := createTestPositionLinkedList(t, []record.Position{1, 2, 3})

		copyFirst, copyLast, err := list.Copy()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if got, expect := copyFirst.Position, int64(1); got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}
		if got, expect := copyLast.Position, int64(3); got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}

		next, err := copyFirst.NextPosition()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if got, expect := next.Position, int64(2); got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}

		nextNext, err := next.NextPosition()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if got, expect := nextNext.Position, int64(3); got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}

		nextNextNext, err := next.NextPosition()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if nextNextNext != nil {
			t.Errorf("Expected nil, got %v", nil, nextNextNext)
		}
	})
}

