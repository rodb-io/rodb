package record

import (
	"fmt"
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

func NewStringColumnsMock(
	values map[string]string,
	position Position,
) *Mock {
	return NewMock(
		values,
		map[string]int{},
		map[string]float64{},
		map[string]bool{},
		position,
	)
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

	return nil, fmt.Errorf("Column '%v' does not exist in mocked record", field)
}

func (record *Mock) Position() Position {
	return record.position
}
