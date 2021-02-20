package index

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
	"rods/pkg/config"
	"rods/pkg/input"
	"rods/pkg/record"
)

type MemoryMap struct {
	config *config.MemoryMapIndex
	input  input.Input
	logger *logrus.Logger
	index  map[interface{}][]record.Position
}

func NewMemoryMap(
	config *config.MemoryMapIndex,
	input input.Input,
	log *logrus.Logger,
) *MemoryMap {
	return &MemoryMap{
		config: config,
		input:  input,
		logger: log,
	}
}

func (mm *MemoryMap) Prepare() error {
	mm.index = make(map[interface{}][]record.Position)

	for result := range mm.input.IterateAll() {
		if result.Error != nil {
			return result.Error
		}

		value, err := result.Record.Get(mm.config.Column)
		if err != nil {
			return err
		}

		if value != nil {
			value = reflect.ValueOf(value).Elem().Interface()
		}

		valueIndexes, valueIndexesExists := mm.index[value]
		if valueIndexesExists {
			mm.index[value] = append(valueIndexes, result.Record.Position())
		} else {
			mm.index[value] = []record.Position{result.Record.Position()}
		}
	}

	return nil
}

func (mm *MemoryMap) GetRecordsByColumn(inputName string, columnName string, limit uint) ([]record.Record, error) {
	if inputName != mm.config.Input {
		return nil, fmt.Errorf("This index does not handle the input '%v'.", inputName)
	}
	if columnName != mm.config.Column {
		return nil, fmt.Errorf("This index does not handle the column '%v'.", columnName)
	}

	indexedValues, foundIndexedValues := mm.index[columnName]
	if !foundIndexedValues {
		return make([]record.Record, 0), nil
	}

	length := int(limit)
	if length > len(indexedValues) {
		length = len(indexedValues)
	}

	records := make([]record.Record, length)
	for i := 0; i < length; i++ {
		indexedRecord, err := mm.input.Get(indexedValues[i])
		if err != nil {
			return nil, fmt.Errorf("Error retrieving indexed record: %w", err)
		}

		records[i] = indexedRecord
	}

	return records, nil
}

func (mm *MemoryMap) Close() error {
	return nil
}
