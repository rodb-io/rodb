package record

import ()

type Position = int64

type Record interface {
	GetString(field string) (*string, error)
	GetInteger(field string) (*int, error)
	GetFloat(field string) (*float64, error)
	GetBoolean(field string) (*bool, error)
	Get(field string) (interface{}, error)
	Position() Position
}
