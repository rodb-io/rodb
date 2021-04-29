package index

import (
	"fmt"
	"reflect"
	"rodb.io/pkg/config"
	"rodb.io/pkg/input"
	"rodb.io/pkg/record"
	"rodb.io/pkg/util"
)

type MemoryPartial struct {
	config *config.MemoryPartialIndex
	input  input.Input
	index  map[string]*partialIndexTrieNode
}

func NewMemoryPartial(
	config *config.MemoryPartialIndex,
	inputs input.List,
) (*MemoryPartial, error) {
	input, inputExists := inputs[config.Input]
	if !inputExists {
		return nil, fmt.Errorf("Input '%v' not found in inputs list.", config.Input)
	}

	memoryPartial := &MemoryPartial{
		config: config,
		input:  input,
	}

	err := memoryPartial.Reindex()
	if err != nil {
		return nil, err
	}

	return memoryPartial, nil
}

func (mp *MemoryPartial) Name() string {
	return mp.config.Name
}

func (mp *MemoryPartial) Reindex() error {
	index := make(map[string]*partialIndexTrieNode)
	for _, property := range mp.config.Properties {
		index[property] = &partialIndexTrieNode{
			value:         rune(0),
			nextSibling:   nil,
			firstChild:    nil,
			lastChild:     nil,
			firstPosition: nil,
			lastPosition:  nil,
		}
	}

	updateProgress := util.TrackProgress(mp.input, mp.config.Logger)

	inputIterator, end, err := mp.input.IterateAll()
	if err != nil {
		return err
	}
	defer func() {
		err := end()
		if err != nil {
			mp.config.Logger.Errorf("Error while closing the input iterator: %v", err)
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

		for _, property := range mp.config.Properties {
			value, err := record.Get(property)
			if err != nil {
				return err
			}

			if value != nil {
				value = reflect.ValueOf(value).Interface()
			}

			err = mp.addValueToIndex(index, property, value, record.Position())
			if err != nil {
				return fmt.Errorf("Cannot index the property '%v': ", property)
			}
		}
	}

	mp.index = index
	mp.config.Logger.Infof("Successfully finished indexing")

	return nil
}

func (mp *MemoryPartial) addValueToIndex(
	index map[string]*partialIndexTrieNode,
	property string,
	value interface{},
	position record.Position,
) error {
	if valueArray, valueIsArray := value.([]interface{}); valueIsArray {
		for _, valueArrayValue := range valueArray {
			err := mp.addValueToIndex(index, property, valueArrayValue, position)
			if err != nil {
				return err
			}
		}
	}

	stringValue, valueIsString := value.(string)
	if !valueIsString {
		return fmt.Errorf("Cannot index the value '%v' from property '%v' because it is not a string.", value, property)
	}

	root := index[property]
	runes := []rune(stringValue)
	for i := 0; i < len(runes)-1; i++ {
		root.addSequence(runes[i:], position)
	}

	return nil
}

func (mp *MemoryPartial) GetRecordPositions(
	input input.Input,
	filters map[string]interface{},
) (record.PositionIterator, error) {
	if input != mp.input {
		return nil, fmt.Errorf("This index does not handle the input '%v'.", input.Name())
	}

	if len(filters) == 0 {
		return nil, fmt.Errorf("This index requires at least one filter.")
	}

	individualFiltersResults := make([]record.PositionIterator, 0, len(filters))
	for propertyName, filter := range filters {
		if !mp.config.DoesHandleProperty(propertyName) {
			return nil, fmt.Errorf("This index does not handle the property '%v'.", propertyName)
		}

		indexedTree, foundIndexedValues := mp.index[propertyName]
		if !foundIndexedValues {
			return record.EmptyIterator, nil
		}

		stringFilter, filterIsString := filter.(string)
		if !filterIsString {
			return nil, fmt.Errorf("Cannot filter the value '%v' from property '%v' because it is not a string.", filter, propertyName)
		}
		filterRunes := []rune(stringFilter)

		indexedResults := indexedTree.getSequence(filterRunes)
		if indexedResults == nil {
			return record.EmptyIterator, nil
		}

		individualFiltersResults = append(individualFiltersResults, indexedResults.Iterate())
	}

	return record.JoinPositionIterators(individualFiltersResults...), nil
}

func (mp *MemoryPartial) Close() error {
	return nil
}
