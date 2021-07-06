package input

import (
	"fmt"
	"rodb.io/pkg/input/record"
)

type MockRecord struct {
	strings  map[string]string
	integers map[string]int
	floats   map[string]float64
	booleans map[string]bool
	position record.Position
}

func NewMockRecord(
	strings map[string]string,
	integers map[string]int,
	floats map[string]float64,
	booleans map[string]bool,
	position record.Position,
) *MockRecord {
	return &MockRecord{
		strings:  strings,
		integers: integers,
		floats:   floats,
		booleans: booleans,
		position: position,
	}
}

func NewStringPropertiesMockRecord(
	values map[string]string,
	position record.Position,
) *MockRecord {
	return NewMockRecord(
		values,
		map[string]int{},
		map[string]float64{},
		map[string]bool{},
		position,
	)
}

func (record *MockRecord) All() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for field, value := range record.strings {
		result[field] = value
	}
	for field, value := range record.integers {
		result[field] = value
	}
	for field, value := range record.floats {
		result[field] = value
	}
	for field, value := range record.booleans {
		result[field] = value
	}

	return result, nil
}

func (record *MockRecord) Get(path string) (interface{}, error) {
	if value, ok := record.strings[path]; ok {
		return value, nil
	}
	if value, ok := record.integers[path]; ok {
		return value, nil
	}
	if value, ok := record.floats[path]; ok {
		return value, nil
	}
	if value, ok := record.booleans[path]; ok {
		return value, nil
	}

	return nil, fmt.Errorf("Property '%v' does not exist in mocked record", path)
}

func (record *MockRecord) Position() record.Position {
	return record.position
}
