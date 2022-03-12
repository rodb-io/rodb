package wildcard

import (
	"encoding/binary"
	"github.com/rodb-io/rodb/pkg/input/record"
)

type PositionLinkedListOffset int64

const PositionLinkedListSize int = 16

type PositionLinkedList struct {
	stream             *Stream
	offset             PositionLinkedListOffset
	Position           record.Position
	nextPositionOffset PositionLinkedListOffset
}

func (list *PositionLinkedList) NextPosition() (*PositionLinkedList, error) {
	return GetPositionLinkedList(list.stream, list.nextPositionOffset)
}

func NewPositionLinkedList(
	stream *Stream,
	position record.Position,
) (*PositionLinkedList, error) {
	node := &PositionLinkedList{
		stream:             stream,
		offset:             0,
		Position:           position,
		nextPositionOffset: 0,
	}

	if err := node.Save(); err != nil {
		return nil, err
	}

	return node, nil
}

func NewPositionLinkedListFromArray(
	stream *Stream,
	positions []record.Position,
) (*PositionLinkedList, error) {
	if len(positions) == 0 {
		return nil, nil
	}

	list, err := NewPositionLinkedList(stream, positions[0])
	if err != nil {
		return nil, err
	}

	current := list
	for i := 1; i < len(positions); i++ {
		newCurrent, err := NewPositionLinkedList(stream, positions[i])
		if err != nil {
			return nil, err
		}

		current.nextPositionOffset = newCurrent.offset
		if err := current.Save(); err != nil {
			return nil, err
		}

		current = newCurrent
	}

	return list, nil
}

func GetPositionLinkedList(
	stream *Stream,
	offset PositionLinkedListOffset,
) (*PositionLinkedList, error) {
	if offset == 0 {
		return nil, nil
	}

	serialized, err := stream.Get(int64(offset), PositionLinkedListSize)
	if err != nil {
		return nil, err
	}

	position := &PositionLinkedList{
		stream: stream,
		offset: offset,
	}

	position.Unserialize(serialized)

	return position, nil
}

func (list *PositionLinkedList) Serialize() []byte {
	buffer := make([]byte, PositionLinkedListSize)

	binary.BigEndian.PutUint64(buffer[0:8], uint64(list.Position))
	binary.BigEndian.PutUint64(buffer[8:16], uint64(list.nextPositionOffset))

	return buffer
}

func (list *PositionLinkedList) Unserialize(data []byte) {
	list.Position = int64(binary.BigEndian.Uint64(data[0:8]))
	list.nextPositionOffset = PositionLinkedListOffset(binary.BigEndian.Uint64(data[8:16]))
}

func (list *PositionLinkedList) Save() error {
	if list.offset == 0 {
		newOffset, err := list.stream.Add(list.Serialize())
		if err != nil {
			return err
		}
		list.offset = PositionLinkedListOffset(newOffset)
	} else {
		if err := list.stream.Replace(int64(list.offset), list.Serialize()); err != nil {
			return err
		}
	}

	return nil
}

func (list *PositionLinkedList) Iterate() record.PositionIterator {
	current := list
	return func() (*record.Position, error) {
		var err error
		for current != nil {
			position := current.Position
			current, err = current.NextPosition()
			if err != nil {
				return nil, err
			}
			return &position, nil
		}

		return nil, nil
	}
}

func (list *PositionLinkedList) Copy() (
	first *PositionLinkedList,
	last *PositionLinkedList,
	err error,
) {
	first, err = NewPositionLinkedList(list.stream, list.Position)
	if err != nil {
		return nil, nil, err
	}

	current, err := list.NextPosition()
	if err != nil {
		return nil, nil, err
	}

	last = first
	for current != nil {
		newCurrent, err := NewPositionLinkedList(current.stream, current.Position)
		if err != nil {
			return nil, nil, err
		}

		last.nextPositionOffset = newCurrent.offset
		if err := last.Save(); err != nil {
			return nil, nil, err
		}

		last = newCurrent

		current, err = current.NextPosition()
		if err != nil {
			return nil, nil, err
		}
	}

	return first, last, nil
}
