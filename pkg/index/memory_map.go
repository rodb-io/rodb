package index

import (
	"errors"
	"fmt"
	"reflect"
	"rodb.io/pkg/config"
	"rodb.io/pkg/input"
	"rodb.io/pkg/record"
	"rodb.io/pkg/util"
)

// Index for the values of a single property
type memoryMapPropertyIndex = map[interface{}]record.PositionList

type MemoryMap struct {
	config *config.MemoryMapIndex
	input  input.Input
	index  map[string]memoryMapPropertyIndex
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
	index := make(map[string]memoryMapPropertyIndex)
	for _, property := range mm.config.Properties {
		index[property] = make(memoryMapPropertyIndex)
	}

	updateProgress := util.TrackProgress(mm.input, mm.config.Logger)

	inputIterator, end, err := mm.input.IterateAll()
	if err != nil {
		return err
	}
	defer func() {
		err := end()
		if err != nil {
			mm.config.Logger.Errorf("Error while closing the input iterator: %v", err)
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

		updateProgress(record.Position())

		for _, property := range mm.config.Properties {
			value, err := record.Get(property)
			if err != nil {
				return err
			}

			if value != nil {
				value = reflect.ValueOf(value).Interface()
			}

			err = mm.addValueToIndex(index, property, value, record.Position())
			if err != nil {
				return fmt.Errorf("Cannot index the property '%v': ", property)
			}
		}
	}

	mm.index = index
	mm.config.Logger.Infof("Successfully finished indexing")

	return nil
}

func (mm *MemoryMap) addValueToIndex(
	index map[string]memoryMapPropertyIndex,
	property string,
	value interface{},
	position record.Position,
) error {
	if valueArray, valueIsArray := value.([]interface{}); valueIsArray {
		for _, valueArrayValue := range valueArray {
			err := mm.addValueToIndex(index, property, valueArrayValue, position)
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

	individualFiltersResults := make([]record.PositionIterator, 0, len(filters))
	for propertyName, filter := range filters {
		if !mm.config.DoesHandleProperty(propertyName) {
			return nil, fmt.Errorf("This index does not handle the property '%v'.", propertyName)
		}

		indexedValues, foundIndexedValues := mm.index[propertyName]
		if !foundIndexedValues {
			return record.EmptyIterator, nil
		}

		indexedResults, foundIndexedResults := indexedValues[filter]
		if !foundIndexedResults {
			return record.EmptyIterator, nil
		}

		individualFiltersResults = append(individualFiltersResults, indexedResults.Iterate())
	}

	return record.JoinPositionIterators(individualFiltersResults...), nil
}

func (mm *MemoryMap) Close() error {
	return nil
}
