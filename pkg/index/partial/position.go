package partial

import (
	"rodb.io/pkg/record"
)

type PositionLinkedList struct {
	Position     record.Position
	NextPosition *PositionLinkedList
}

func (list *PositionLinkedList) Iterate() record.PositionIterator {
	current := list
	return func() (*record.Position, error) {
		for current != nil {
			position := current.Position
			current = current.NextPosition
			return &position, nil
		}

		return nil, nil
	}
}

func (list *PositionLinkedList) Copy() (
	first *PositionLinkedList,
	last *PositionLinkedList,
) {
	first = &PositionLinkedList{
		Position:     list.Position,
		NextPosition: nil,
	}
	last = first
	for current := list.NextPosition; current != nil; current = current.NextPosition {
		newCurrent := &PositionLinkedList{
			Position:     current.Position,
			NextPosition: nil,
		}
		last.NextPosition = newCurrent
		last = newCurrent
	}

	return first, last
}
