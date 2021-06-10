package partial

import (
	"fmt"
	"io/ioutil"
	"rodb.io/pkg/record"
	"testing"
)

func createTestPositionLinkedList(t *testing.T, positions []record.Position) *PositionLinkedList {
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

func TestPositionLinkedListSerialize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		list := PositionLinkedList{
			Position:           1,
			nextPositionOffset: 1234,
		}
		got, err := list.Serialize()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		expect := []byte{
			0, 0, 0, 0, 0, 0, 0, 0x01,
			0, 0, 0, 0, 0, 0, 0x4, 0xD2,
		}
		if expect, got := fmt.Sprintf("%x", expect), fmt.Sprintf("%x", got); expect != got {
			t.Errorf("Expected %v, got %v", expect, got)
		}
	})
	t.Run("empty next", func(t *testing.T) {
		list := PositionLinkedList{
			Position:           1,
			nextPositionOffset: 0,
		}
		got, err := list.Serialize()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		expect := []byte{
			0, 0, 0, 0, 0, 0, 0, 0x01,
			0, 0, 0, 0, 0, 0, 0, 0,
		}
		if expect, got := fmt.Sprintf("%x", expect), fmt.Sprintf("%x", got); expect != got {
			t.Errorf("Expected %v, got %v", expect, got)
		}
	})
}

func TestPositionLinkedListUnserialize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		list := PositionLinkedList{}
		err := list.Unserialize([]byte{0, 0, 0, 0, 0, 0, 0, 0x01, 0, 0, 0, 0, 0, 0, 0x4, 0xD2})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if expect, got := int64(1), list.Position; expect != got {
			t.Errorf("Expected %v, got %v", expect, got)
		}
		if expect, got := PositionLinkedListOffset(1234), list.nextPositionOffset; expect != got {
			t.Errorf("Expected %v, got %v", expect, got)
		}
	})
	t.Run("empty next", func(t *testing.T) {
		list := PositionLinkedList{}
		err := list.Unserialize([]byte{0, 0, 0, 0, 0, 0, 0, 0x01, 0, 0, 0, 0, 0, 0, 0, 0})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if expect, got := int64(1), list.Position; expect != got {
			t.Errorf("Expected %v, got %v", expect, got)
		}
		if expect, got := PositionLinkedListOffset(0), list.nextPositionOffset; expect != got {
			t.Errorf("Expected %v, got %v", expect, got)
		}
	})
	t.Run("from serialize", func(t *testing.T) {
		list1 := PositionLinkedList{
			Position:           1,
			nextPositionOffset: 1234,
		}
		serialized, err := list1.Serialize()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		list2 := PositionLinkedList{}
		err = list2.Unserialize(serialized)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := int64(1), list2.Position; expect != got {
			t.Errorf("Expected %v, got %v", expect, got)
		}
		if expect, got := PositionLinkedListOffset(1234), list2.nextPositionOffset; expect != got {
			t.Errorf("Expected %v, got %v", expect, got)
		}
	})
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

		nextNextNext, err := nextNext.NextPosition()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if nextNextNext != nil {
			t.Errorf("Expected nil, got %v", nextNextNext)
		}
	})
}
