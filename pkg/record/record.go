package record

import ()

type Position = int64

type Record interface {
	All() (map[string]interface{}, error)
	Get(field string) (interface{}, error)
	Position() Position
}
