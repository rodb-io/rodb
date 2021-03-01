package index

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
	"rods/pkg/config"
	"rods/pkg/input"
	"rods/pkg/record"
)

type memoryMapColumnValueIndex = []record.Position
type memoryMapColumnIndex = map[interface{}]memoryMapColumnValueIndex
type memoryMapIndex = map[string]memoryMapColumnIndex

type MemoryMap struct {
	config *config.MemoryMapIndex
	input  input.Input
	logger *logrus.Logger
	index  memoryMapIndex
}

func NewMemoryMap(
	config *config.MemoryMapIndex,
	input input.Input,
	log *logrus.Logger,
) (*MemoryMap, error) {
	memoryMap := &MemoryMap{
		config: config,
		input:  input,
		logger: log,
	}

	memoryMap.index = make(memoryMapIndex)
	for _, column := range memoryMap.config.Columns {
		memoryMap.index[column] = make(memoryMapColumnIndex)
	}

	for result := range memoryMap.input.IterateAll() {
		if result.Error != nil {
			return nil, result.Error
		}

		for _, column := range memoryMap.config.Columns {
			value, err := result.Record.Get(column)
			if err != nil {
				return nil, err
			}

			if value != nil {
				value = reflect.ValueOf(value).Elem().Interface()
			}

			columnIndex := memoryMap.index[column]
			valueIndexes, valueIndexesExists := columnIndex[value]
			if valueIndexesExists {
				columnIndex[value] = append(valueIndexes, result.Record.Position())
			} else {
				columnIndex[value] = []record.Position{result.Record.Position()}
			}
		}
	}

	return memoryMap, nil
}

// Get the records from the given input (if indexed) and that matchesall the given filters
// A limit of 0 means that there is no limit
func (mm *MemoryMap) GetRecords(inputName string, filters map[string]interface{}, limit uint) ([]record.Record, error) {
	if inputName != mm.config.Input {
		return nil, fmt.Errorf("This index does not handle the input '%v'.", inputName)
	}
	if len(filters) == 0 {
		return nil, fmt.Errorf("This index requires at least one filter.")
	}

	individualFiltersResults := make([]memoryMapColumnValueIndex, 0, len(filters))
	for columnName, filter := range filters {
		isHandled := false
		for _, handledColumn := range mm.config.Columns {
			if columnName == handledColumn {
				isHandled = true
				break
			}
		}

		if !isHandled {
			return nil, fmt.Errorf("This index does not handle the column '%v'.", columnName)
		}

		indexedValues, foundIndexedValues := mm.index[columnName]
		if !foundIndexedValues {
			return make([]record.Record, 0), nil
		}

		indexedResults, foundIndexedResults := indexedValues[filter]
		if !foundIndexedResults {
			return make([]record.Record, 0), nil
		}

		individualFiltersResults = append(individualFiltersResults, indexedResults)
	}

	records := make([]record.Record, 0)
	for i := 0; i < len(individualFiltersResults[0]); i++ {
		position := individualFiltersResults[0][i]

		matchesAllCriterias := true
		for j := 1; j < len(individualFiltersResults); j++ {
			matchesCurrentCriteria := false
			for _, currentPosition := range individualFiltersResults[j] {
				if currentPosition == position {
					matchesCurrentCriteria = true
					break
				}
			}

			if !matchesCurrentCriteria {
				matchesAllCriterias = false
				break
			}
		}

		if matchesAllCriterias {
			indexedRecord, err := mm.input.Get(position)
			if err != nil {
				return nil, fmt.Errorf("Error retrieving indexed record: %w", err)
			}

			records = append(records, indexedRecord)
			if limit != 0 && len(records) >= int(limit) {
				break
			}
		}
	}

	return records, nil
}

func (mm *MemoryMap) Close() error {
	return nil
}
