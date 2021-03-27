package record

import (
	"errors"
)

type Position = int64

type PositionList []Position

// Ends when both the record and error are nil at the same time
// a nil position with a non-nil error does not mean it reached the end
// When the end has been reached, the iterator is expected
// to return (nil, nil) even if called again
type PositionIterator func() (*Position, error)

type List []Record

var RecordNotFoundError = errors.New("Record not found")

type Record interface {
	All() (map[string]interface{}, error)
	Get(field string) (interface{}, error)
	Position() Position
}
