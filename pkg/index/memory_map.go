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

type memoryMapPropertyValueIndex = record.PositionList
type memoryMapPropertyIndex = map[interface{}]memoryMapPropertyValueIndex
type memoryMapIndex = map[string]memoryMapPropertyIndex

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
	for _, property := range mm.config.Properties {
		index[property] = make(memoryMapPropertyIndex)
	}

	var addValueToIndex func(property string, value interface{}, position record.Position) error
	addValueToIndex = func(property string, value interface{}, position record.Position) error {
		if valueArray, valueIsArray := value.([]interface{}); valueIsArray {
			for _, valueArrayValue := range valueArray {
				err := addValueToIndex(property, valueArrayValue, position)
				if err != nil {
					return err
				}
			}
		}

		if _, valueIsMap := value.(map[string]interface{}); valueIsMap {
			return errors.New("Indexing objects is not supported")
		}

		propertyIndex := index[property]
		valueIndexes, valueIndexesExists := propertyIndex[value]
		if valueIndexesExists {
			propertyIndex[value] = append(valueIndexes, position)
		} else {
			propertyIndex[value] = record.PositionList{position}
		}

		return nil
	}

	totalSize, err := mm.input.Size()
	if err != nil {
		mm.config.Logger.Errorf("Cannot determine the total size of the input: '%+v'. The indexing progress will not be displayed.", err)
	} else if totalSize == 0 {
		mm.config.Logger.Infoln("The total size of the input is unknown. The indexing progress will not be displayed.")
	}

	nextProgress := time.Now()
	inputIterator, end, err := mm.input.IterateAll()
	if err != nil {
		return err
	}
	defer func() {
		err := end()
		if err != nil {
			mm.config.Logger.Errorf("Error while closing input iterator: %v", err)
		}
	}()

	for {
		record, err := inputIterator()
		if err != nil {
			return err
		}
		if record == nil {
			break
		}

		if totalSize != 0 {
			if now := time.Now(); now.After(nextProgress) {
				progress := float64(record.Position()) / float64(totalSize)
				mm.config.Logger.Infof("Indexing progress: %d%%", int(math.Floor(progress*100)))
				nextProgress = now.Add(time.Second)
			}
		}

		for _, property := range mm.config.Properties {
			value, err := record.Get(property)
			if err != nil {
				return err
			}

			if value != nil {
				value = reflect.ValueOf(value).Interface()
			}

			err = addValueToIndex(property, value, record.Position())
			if err != nil {
				return fmt.Errorf("Cannot index the property '%v': ", property)
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

	individualFiltersResults := make([]memoryMapPropertyValueIndex, 0, len(filters))
	for propertyName, filter := range filters {
		if !mm.config.DoesHandleProperty(propertyName) {
			return nil, fmt.Errorf("This index does not handle the property '%v'.", propertyName)
		}

		indexedValues, foundIndexedValues := mm.index[propertyName]
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
