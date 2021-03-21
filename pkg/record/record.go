package record

type Position = int64

type PositionList []Position

type List []Record

type Record interface {
	All() (map[string]interface{}, error)
	Get(field string) (interface{}, error)
	Position() Position
}
