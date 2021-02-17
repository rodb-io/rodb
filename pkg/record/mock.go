package record

import (
	"errors"
)

type Mock struct {
	strings  map[string]string
	integers map[string]int
	floats   map[string]float64
	booleans map[string]bool
	position Position
}

func NewMock(
	strings map[string]string,
	integers map[string]int,
	floats map[string]float64,
	booleans map[string]bool,
	position Position,
) *Mock {
	return &Mock{
		strings:  strings,
		integers: integers,
		floats:   floats,
		booleans: booleans,
		position: position,
	}
}

func NewSingleStringColumnMock(
	key string,
	value string,
	position Position,
) *Mock {
	return NewMock(
		map[string]string{
			key: value,
		},
		map[string]int{},
		map[string]float64{},
		map[string]bool{},
		position,
	)
}

func (record *Mock) GetString(field string) (*string, error) {
	value, ok := record.strings[field]
	if !ok {
		return nil, errors.New("String column '" + field + "' does not exist in mocked record")
	}

	return &value, nil
}

func (record *Mock) GetInteger(field string) (*int, error) {
	value, ok := record.integers[field]
	if !ok {
		return nil, errors.New("Integer column '" + field + "' does not exist in mocked record")
	}

	return &value, nil
}

func (record *Mock) GetFloat(field string) (*float64, error) {
	value, ok := record.floats[field]
	if !ok {
		return nil, errors.New("Float column '" + field + "' does not exist in mocked record")
	}

	return &value, nil
}

func (record *Mock) GetBoolean(field string) (*bool, error) {
	value, ok := record.booleans[field]
	if !ok {
		return nil, errors.New("Boolean column '" + field + "' does not exist in mocked record")
	}

	return &value, nil
}

func (record *Mock) Get(field string) (interface{}, error) {
	if value, ok := record.strings[field]; ok {
		return &value, nil
	}
	if value, ok := record.integers[field]; ok {
		return &value, nil
	}
	if value, ok := record.floats[field]; ok {
		return &value, nil
	}
	if value, ok := record.booleans[field]; ok {
		return &value, nil
	}

	return nil, errors.New("Generic column '" + field + "' does not exist in mocked record")
}

func (record *Mock) Position() Position {
	return record.position
}
