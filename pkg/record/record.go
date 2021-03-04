package record

import ()

type Position = int64

type Record interface {
	Get(field string) (interface{}, error)
	Position() Position
}
