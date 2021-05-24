package index

import (
	"rodb.io/pkg/record"
)

type partialIndexPositionLinkedList struct {
	Position     record.Position
	NextPosition *partialIndexPositionLinkedList
}

func (list *partialIndexPositionLinkedList) Iterate() record.PositionIterator {
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

func (list *partialIndexPositionLinkedList) Copy() (
	first *partialIndexPositionLinkedList,
	last *partialIndexPositionLinkedList,
) {
	first = &partialIndexPositionLinkedList{
		Position:     list.Position,
		NextPosition: nil,
	}
	last = first
	for current := list.NextPosition; current != nil; current = current.NextPosition {
		newCurrent := &partialIndexPositionLinkedList{
			Position:     current.Position,
			NextPosition: nil,
		}
		last.NextPosition = newCurrent
		last = newCurrent
	}

	return first, last
}
