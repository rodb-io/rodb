package wildcard

import (
	"fmt"
	"rodb.io/pkg/input/record"
	"testing"
)

func createTestPositionLinkedList(t *testing.T, positions []record.Position) *PositionLinkedList {
	stream := createTestStream(t)

	list, err := NewPositionLinkedListFromArray(stream, positions)
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	return list
}

func TestNewPositionLinkedListFromArray(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		stream := createTestStream(t)
		list0, err := NewPositionLinkedListFromArray(stream, []record.Position{1, 42, 123})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if got, expect := list0.Position, int64(1); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		list1, err := list0.NextPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if got, expect := list1.Position, int64(42); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		list2, err := list1.NextPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if got, expect := list2.Position, int64(123); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		list3, err := list2.NextPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if list3 != nil {
			t.Fatalf("Expected nil, got %v", list3)
		}
	})
}

func TestPositionLinkedListSerialize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		list := PositionLinkedList{
			Position:           1,
			nextPositionOffset: 1234,
		}

		expect := []byte{
			0, 0, 0, 0, 0, 0, 0, 0x01,
			0, 0, 0, 0, 0, 0, 0x4, 0xD2,
		}
		if expect, got := fmt.Sprintf("%x", expect), fmt.Sprintf("%x", list.Serialize()); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
	t.Run("empty next", func(t *testing.T) {
		list := PositionLinkedList{
			Position:           1,
			nextPositionOffset: 0,
		}

		expect := []byte{
			0, 0, 0, 0, 0, 0, 0, 0x01,
			0, 0, 0, 0, 0, 0, 0, 0,
		}
		if expect, got := fmt.Sprintf("%x", expect), fmt.Sprintf("%x", list.Serialize()); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestPositionLinkedListUnserialize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		list := PositionLinkedList{}
		data := []byte{
			0, 0, 0, 0, 0, 0, 0, 0x01,
			0, 0, 0, 0, 0, 0, 0x4, 0xD2,
		}
		list.Unserialize(data)
		if expect, got := int64(1), list.Position; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := PositionLinkedListOffset(1234), list.nextPositionOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
	t.Run("empty next", func(t *testing.T) {
		list := PositionLinkedList{}
		data := []byte{
			0, 0, 0, 0, 0, 0, 0, 0x01,
			0, 0, 0, 0, 0, 0, 0, 0,
		}
		list.Unserialize(data)
		if expect, got := int64(1), list.Position; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := PositionLinkedListOffset(0), list.nextPositionOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
	t.Run("from serialize", func(t *testing.T) {
		list1 := PositionLinkedList{
			Position:           1,
			nextPositionOffset: 1234,
		}

		list2 := PositionLinkedList{}
		list2.Unserialize(list1.Serialize())

		if expect, got := int64(1), list2.Position; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := PositionLinkedListOffset(1234), list2.nextPositionOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestPositionLinkedListSave(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		stream := createTestStream(t)
		initialSize := stream.streamSize

		list := PositionLinkedList{
			stream:             stream,
			offset:             0,
			Position:           1,
			nextPositionOffset: 1234,
		}
		if err := list.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := initialSize, int64(list.offset); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		expectBytes := []byte{
			0, 0, 0, 0, 0, 0, 0, 0x01,
			0, 0, 0, 0, 0, 0, 0x4, 0xD2,
		}
		gotBytes, err := stream.Get(initialSize, 16)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := fmt.Sprintf("%x", expectBytes), fmt.Sprintf("%x", gotBytes); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})

	t.Run("update", func(t *testing.T) {
		stream := createTestStream(t)
		offset, err := stream.Add([]byte{
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		list := PositionLinkedList{
			stream:             stream,
			offset:             PositionLinkedListOffset(offset),
			Position:           1,
			nextPositionOffset: 1234,
		}
		if err := list.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := offset, int64(list.offset); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		expectBytes := []byte{
			0, 0, 0, 0, 0, 0, 0, 0x01,
			0, 0, 0, 0, 0, 0, 0x4, 0xD2,
		}
		gotBytes, err := stream.Get(offset, 16)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := fmt.Sprintf("%x", expectBytes), fmt.Sprintf("%x", gotBytes); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestPositionLinkedListToIterator(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		list := createTestPositionLinkedList(t, []record.Position{1, 42, 123})
		iterator := list.Iterate()

		list0, err := iterator()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if got, expect := list0, int64(1); got == nil || *got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		list1, err := iterator()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if got, expect := list1, int64(42); got == nil || *got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		list2, err := iterator()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if got, expect := list2, int64(123); got == nil || *got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		list3, err := iterator()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if list3 != nil {
			t.Fatalf("Expected nil, got %v", list3)
		}
	})
}

func TestPositionLinkedListCopy(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		list := createTestPositionLinkedList(t, []record.Position{1, 2, 3})

		copyFirst, copyLast, err := list.Copy()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if got, expect := copyFirst.Position, int64(1); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if got, expect := copyLast.Position, int64(3); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		next, err := copyFirst.NextPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if got, expect := next.Position, int64(2); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		nextNext, err := next.NextPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if got, expect := nextNext.Position, int64(3); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		nextNextNext, err := nextNext.NextPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if nextNextNext != nil {
			t.Fatalf("Expected nil, got %v", nextNextNext)
		}
	})
}
