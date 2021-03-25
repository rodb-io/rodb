package record

type Position = int64

type PositionList []Position

// Ends when both the record and error are nil at the same time
// a nil position with a non-nil error does not mean it reached the end
type PositionIterator func() (*Position, error)

type List []Record

type Record interface {
	All() (map[string]interface{}, error)
	Get(field string) (interface{}, error)
	Position() Position
}
