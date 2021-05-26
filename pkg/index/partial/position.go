package partial

import (
	"bytes"
	"encoding/binary"
	"rodb.io/pkg/record"
)

type PositionLinkedListOffset int64

const PositionLinkedListSize int = 42 // TODO

type PositionLinkedList struct {
	stream        *Stream
	offset        PositionLinkedListOffset
	Position     record.Position
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
		stream:        stream,
		offset:        0,
		Position: position,
		nextPositionOffset:  0,
	}

	err := node.Save()
	if err != nil {
		return nil, err
	}

	return node, nil
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
		stream:        stream,
		offset: offset,
	}

	err = position.Unserialize(serialized)
	if err != nil {
		return nil, err
	}

	return position, nil
}

func (list *PositionLinkedList) Serialize() ([]byte, error) {
	var err error
	buffer := &bytes.Buffer{}

	if err = binary.Write(buffer, binary.LittleEndian, list.Position); err != nil {
		return nil, err
	}
	if err = binary.Write(buffer, binary.LittleEndian, list.nextPositionOffset); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (list *PositionLinkedList) Unserialize(data []byte) error {
	var err error
	buffer := bytes.NewBuffer(data)

	if err = binary.Read(buffer, binary.LittleEndian, &list.Position); err != nil {
		return err
	}
	if err = binary.Read(buffer, binary.LittleEndian, &list.nextPositionOffset); err != nil {
		return err
	}

	return nil
}

func (list *PositionLinkedList) Save() error {
	serialized, err := list.Serialize()
	if err != nil {
		return err
	}
	// TODO set size PositionLinkedListSize

	if list.offset == 0 {
		newOffset, err := list.stream.Add(serialized)
		if err != nil {
			return err
		}
		list.offset = PositionLinkedListOffset(newOffset)
	} else {
		err := list.stream.Replace(int64(list.offset), serialized)
		if err != nil {
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
		err = last.Save()
		if err != nil {
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
