package index

import (
	"errors"
	"fmt"
	"reflect"
	"github.com/rodb-io/rodb/pkg/input"
	"github.com/rodb-io/rodb/pkg/input/record"
	"github.com/rodb-io/rodb/pkg/util"
)

// Index for the values of a single property
type mapPropertyIndex = map[interface{}]record.PositionList

type Map struct {
	config *MapConfig
	input  input.Input
	index  map[string]mapPropertyIndex
}

func NewMap(
	config *MapConfig,
	inputs input.List,
) (*Map, error) {
	input, inputExists := inputs[config.Input]
	if !inputExists {
		return nil, fmt.Errorf("Input '%v' not found in inputs list.", config.Input)
	}

	mapIndex := &Map{
		config: config,
		input:  input,
	}

	if err := mapIndex.reindex(); err != nil {
		return nil, err
	}

	return mapIndex, nil
}

func (mapIndex *Map) Name() string {
	return mapIndex.config.Name
}

func (mapIndex *Map) reindex() error {
	index := make(map[string]mapPropertyIndex)
	for _, property := range mapIndex.config.Properties {
		index[property] = make(mapPropertyIndex)
	}

	updateProgress := util.TrackProgress(mapIndex.input, mapIndex.config.Logger)

	inputIterator, end, err := mapIndex.input.IterateAll()
	if err != nil {
		return err
	}
	defer func() {
		if err := end(); err != nil {
			mapIndex.config.Logger.Errorf("Error while closing the input iterator: %v", err)
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

		for _, property := range mapIndex.config.Properties {
			value, err := record.Get(property)
			if err != nil {
				return err
			}

			if value != nil {
				value = reflect.ValueOf(value).Interface()
			}

			if err := mapIndex.addValueToIndex(index, property, value, record.Position()); err != nil {
				return fmt.Errorf("Cannot index the property '%v': ", property)
			}
		}
	}

	mapIndex.index = index
	mapIndex.config.Logger.Infof("Successfully finished indexing")

	return nil
}

func (mapIndex *Map) addValueToIndex(
	index map[string]mapPropertyIndex,
	property string,
	value interface{},
	position record.Position,
) error {
	if valueArray, valueIsArray := value.([]interface{}); valueIsArray {
		for _, valueArrayValue := range valueArray {
			if err := mapIndex.addValueToIndex(index, property, valueArrayValue, position); err != nil {
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

func (mapIndex *Map) GetRecordPositions(
	input input.Input,
	filters map[string]interface{},
) (record.PositionIterator, error) {
	if input != mapIndex.input {
		return nil, fmt.Errorf("This index does not handle the input '%v'.", input.Name())
	}

	if len(filters) == 0 {
		return nil, fmt.Errorf("This index requires at least one filter.")
	}

	individualFiltersResults := make([]record.PositionIterator, 0, len(filters))
	for propertyName, filter := range filters {
		if !mapIndex.config.DoesHandleProperty(propertyName) {
			return nil, fmt.Errorf("This index does not handle the property '%v'.", propertyName)
		}

		indexedValues, foundIndexedValues := mapIndex.index[propertyName]
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

func (mapIndex *Map) Close() error {
	return nil
}
