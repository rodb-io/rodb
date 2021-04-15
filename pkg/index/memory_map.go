package index

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"rodb.io/pkg/config"
	"rodb.io/pkg/input"
	"rodb.io/pkg/record"
	"time"
)

type memoryMapColumnValueIndex = record.PositionList
type memoryMapColumnIndex = map[interface{}]memoryMapColumnValueIndex
type memoryMapIndex = map[string]memoryMapColumnIndex

type MemoryMap struct {
	config *config.MemoryMapIndex
	input  input.Input
	index  memoryMapIndex
}

func NewMemoryMap(
	config *config.MemoryMapIndex,
	inputs input.List,
) (*MemoryMap, error) {
	input, inputExists := inputs[config.Input]
	if !inputExists {
		return nil, fmt.Errorf("Input '%v' not found in inputs list.", config.Input)
	}

	memoryMap := &MemoryMap{
		config: config,
		input:  input,
	}

	for _, columnName := range memoryMap.config.Columns {
		if !memoryMap.input.HasColumn(columnName) {
			return nil, errors.New("Input '" + memoryMap.config.Input + "' does not have a column named '" + columnName + "'.")
		}
		parser := memoryMap.input.GetColumnParser(columnName)
		if !parser.Primitive() {
			return nil, errors.New("Column '" + columnName + "' defined in index '" + memoryMap.Name() + "' cannot be used because it's not a primitive type.")
		}
	}

	err := memoryMap.Reindex()
	if err != nil {
		return nil, err
	}

	return memoryMap, nil
}

func (mm *MemoryMap) Name() string {
	return mm.config.Name
}

func (mm *MemoryMap) Reindex() error {
	index := make(memoryMapIndex)
	for _, column := range mm.config.Columns {
		index[column] = make(memoryMapColumnIndex)
	}

	totalSize, err := mm.input.Size()
	if err != nil {
		mm.config.Logger.Errorf("Cannot determine the total size of the input: '%+v'. The indexing progress will not be displayed.", err)
	} else if totalSize == 0 {
		mm.config.Logger.Infoln("The total size of the input is unknown. The indexing progress will not be displayed.")
	}

	nextProgress := time.Now()
	inputIterator := mm.input.IterateAll()
	for result := range inputIterator {
		if result.Error != nil {
			return result.Error
		}

		if totalSize != 0 {
			if now := time.Now(); now.After(nextProgress) {
				progress := float64(result.Record.Position()) / float64(totalSize)
				mm.config.Logger.Infof("Indexing progress: %d%%", int(math.Floor(progress*100)))
				nextProgress = now.Add(time.Second)
			}
		}

		for _, column := range mm.config.Columns {
			value, err := result.Record.Get(column)
			if err != nil {
				return err
			}

			if value != nil {
				value = reflect.ValueOf(value).Interface()
			}

			columnIndex := index[column]
			valueIndexes, valueIndexesExists := columnIndex[value]
			if valueIndexesExists {
				columnIndex[value] = append(valueIndexes, result.Record.Position())
			} else {
				columnIndex[value] = record.PositionList{result.Record.Position()}
			}
		}
	}

	mm.index = index
	mm.config.Logger.Infof("Successfully finished indexing")

	return nil
}

// Get the record positions (if indexed) that matches all the given filters
// A limit of 0 means that there is no limit
func (mm *MemoryMap) GetRecordPositions(
	input input.Input,
	filters map[string]interface{},
) (record.PositionIterator, error) {
	if input != mm.input {
		return nil, fmt.Errorf("This index does not handle the input '%v'.", input.Name())
	}

	if len(filters) == 0 {
		return nil, fmt.Errorf("This index requires at least one filter.")
	}

	individualFiltersResults := make([]memoryMapColumnValueIndex, 0, len(filters))
	for columnName, filter := range filters {
		if !mm.config.DoesHandleColumn(columnName) {
			return nil, fmt.Errorf("This index does not handle the column '%v'.", columnName)
		}

		indexedValues, foundIndexedValues := mm.index[columnName]
		if !foundIndexedValues {
			return nil, nil
		}

		indexedResults, foundIndexedResults := indexedValues[filter]
		if !foundIndexedResults {
			return nil, nil
		}

		individualFiltersResults = append(individualFiltersResults, indexedResults)
	}

	var i int = 0
	return func() (*record.Position, error) {
		for i < len(individualFiltersResults[0]) {
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

			i++
			if matchesAllCriterias {
				return &position, nil
			}
		}

		return nil, nil
	}, nil
}

func (mm *MemoryMap) Close() error {
	return nil
}
