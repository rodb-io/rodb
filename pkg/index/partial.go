package index

import (
	"fmt"
	"os"
	"reflect"
	"rodb.io/pkg/config"
	"rodb.io/pkg/input"
	"rodb.io/pkg/record"
	"rodb.io/pkg/util"
	"strings"
)

type Partial struct {
	config *config.PartialIndex
	input  input.Input
	index  map[string]*partialIndexTreeNode
}

func NewPartial(
	config *config.PartialIndex,
	inputs input.List,
) (*Partial, error) {
	input, inputExists := inputs[config.Input]
	if !inputExists {
		return nil, fmt.Errorf("Input '%v' not found in inputs list.", config.Input)
	}

	partialIndex := &Partial{
		config: config,
		input:  input,
	}

	err := partialIndex.Reindex()
	if err != nil {
		return nil, err
	}

	return partialIndex, nil
}

func (partialIndex *Partial) Name() string {
	return partialIndex.config.Name
}

func (partialIndex *Partial) Reindex() error {
	indexFile, err := os.Create("/temp/test-index")
	if err != nil {
		return err
	}
	var indexFileSize int64 = 0

	index := make(map[string]*partialIndexTreeNode)
	for _, property := range partialIndex.config.Properties {
		index[property] = createEmptyPartialIndexTreeNode(indexFile, &indexFileSize)
	}

	updateProgress := util.TrackProgress(partialIndex.input, partialIndex.config.Logger)

	inputIterator, end, err := partialIndex.input.IterateAll()
	if err != nil {
		return err
	}
	defer func() {
		err := end()
		if err != nil {
			partialIndex.config.Logger.Errorf("Error while closing the input iterator: %v", err)
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

		for _, property := range partialIndex.config.Properties {
			value, err := record.Get(property)
			if err != nil {
				return err
			}

			if value != nil {
				value = reflect.ValueOf(value).Interface()
			}

			err = partialIndex.addValueToIndex(
				index,
				property,
				value,
				record.Position(),
			)
			if err != nil {
				return fmt.Errorf("Cannot index the property '%v': ", property)
			}
		}
	}

	partialIndex.index = index
	partialIndex.config.Logger.Infof("Successfully finished indexing")

	return nil
}

func (partialIndex *Partial) addValueToIndex(
	index map[string]*partialIndexTreeNode,
	property string,
	value interface{},
	position record.Position,
) error {
	if valueArray, valueIsArray := value.([]interface{}); valueIsArray {
		for _, valueArrayValue := range valueArray {
			err := partialIndex.addValueToIndex(index, property, valueArrayValue, position)
			if err != nil {
				return err
			}
		}
	}

	stringValue, valueIsString := value.(string)
	if !valueIsString {
		return fmt.Errorf("Cannot index the value '%v' from property '%v' because it is not a string.", value, property)
	}

	if partialIndex.config.ShouldIgnoreCase() {
		stringValue = strings.ToLower(stringValue)
	}

	root := index[property]
	bytes := []byte(stringValue)
	for i := 0; i < len(bytes); i++ {
		root.addSequence(bytes[i:], position)
	}

	return nil
}

func (partialIndex *Partial) GetRecordPositions(
	input input.Input,
	filters map[string]interface{},
) (record.PositionIterator, error) {
	if input != partialIndex.input {
		return nil, fmt.Errorf("This index does not handle the input '%v'.", input.Name())
	}

	if len(filters) == 0 {
		return nil, fmt.Errorf("This index requires at least one filter.")
	}

	individualFiltersResults := make([]record.PositionIterator, 0, len(filters))
	for propertyName, filter := range filters {
		if !partialIndex.config.DoesHandleProperty(propertyName) {
			return nil, fmt.Errorf("This index does not handle the property '%v'.", propertyName)
		}

		indexedTree, foundIndexedValues := partialIndex.index[propertyName]
		if !foundIndexedValues {
			return record.EmptyIterator, nil
		}

		stringFilter, filterIsString := filter.(string)
		if !filterIsString {
			return nil, fmt.Errorf("Cannot filter the value '%v' from property '%v' because it is not a string.", filter, propertyName)
		}

		if partialIndex.config.ShouldIgnoreCase() {
			stringFilter = strings.ToLower(stringFilter)
		}

		indexedResults := indexedTree.getSequence([]byte(stringFilter))
		if indexedResults == nil {
			return record.EmptyIterator, nil
		}

		individualFiltersResults = append(individualFiltersResults, indexedResults.Iterate())
	}

	return record.JoinPositionIterators(individualFiltersResults...), nil
}

func (partialIndex *Partial) Close() error {
	return nil
}