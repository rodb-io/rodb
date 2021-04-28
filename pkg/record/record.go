package record

import (
	"errors"
)

type Position = int64

type PositionList []Position

type PositionLinkedList struct{
	Position Position
	NextPosition *PositionLinkedList
}

// Ends when both the position and error are nil at the same time
// a nil position with a non-nil error does not mean it reached the end
// When the end has been reached, the iterator is expected
// to return (nil, nil), even if called again
type PositionIterator func() (*Position, error)

type List []Record

// Ends when both the record and error are nil at the same time
// a nil record with a non-nil error does not mean it reached the end
// When the end has been reached, the iterator is expected
// to return (nil, nil), even if called again
type Iterator func() (Record, error)

var RecordNotFoundError = errors.New("Record not found")

type Record interface {
	// Returns all the record's data. Each value may be a
	// []interface{} or map[string]interface{}, recursively
	All() (map[string]interface{}, error)

	// Returns the value matching the given path. The path is a dot-separated string.
	// Array indexes does not have a specific syntax, ie foo.0.bar.1 ...
	Get(path string) (interface{}, error)

	Position() Position
}

func (list PositionList) Iterate() PositionIterator {
	var i int = 0
	return func() (*Position, error) {
		for i < len(list) {
			position := list[i]
			i++
			return &position, nil
		}

		return nil, nil
	}
}
