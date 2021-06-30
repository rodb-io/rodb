package index

import (
	"fmt"
	"os"
	"reflect"
	"rodb.io/pkg/config"
	"rodb.io/pkg/index/partial"
	"rodb.io/pkg/input"
	"rodb.io/pkg/record"
	"rodb.io/pkg/util"
	"strings"
)

type Partial struct {
	config *config.PartialIndex
	input  input.Input
	index  map[string]*partial.TreeNode
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

	_, err := os.Stat(partialIndex.config.Path)
	if os.IsNotExist(err) {
		err = partialIndex.createIndex()
		if err != nil {
			return nil, fmt.Errorf("Error while creating the index: %w", err)
		}
	} else if err != nil {
		return nil, err
	} else {
		err = partialIndex.loadIndex()
		if err != nil {
			return nil, fmt.Errorf("Error while loading the index: %w", err)
		}
	}

	return partialIndex, nil
}

func (partialIndex *Partial) Name() string {
	return partialIndex.config.Name
}

func (partialIndex *Partial) createIndex() error {
	indexFile, err := os.Create(partialIndex.config.Path)
	if err != nil {
		return err
	}

	indexStream := partial.NewStream(indexFile, 0)

	metadata, err := partial.NewMetadata(indexStream, partial.MetadataInput{
		Input:          partialIndex.input,
		IgnoreCase:     *partialIndex.config.IgnoreCase,
		RootNodesCount: len(partialIndex.config.Properties),
	})
	if err != nil {
		return err
	}

	index := make(map[string]*partial.TreeNode)
	for propertyIndex, property := range partialIndex.config.Properties {
		index[property], err = partial.NewEmptyTreeNode(indexStream)
		if err != nil {
			return err
		}

		metadata.SetRootNode(propertyIndex, index[property])
	}

	err = metadata.Save()
	if err != nil {
		return err
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

	metadata.SetCompleted(true)
	err = metadata.Save()
	if err != nil {
		return err
	}

	partialIndex.index = index

	indexStat, err := indexFile.Stat()
	if err != nil {
		return err
	}

	partialIndex.config.Logger.WithField("indexSize", indexStat.Size()).Infof("Successfully finished indexing")

	return nil
}

func (partialIndex *Partial) loadIndex() error {
	indexFile, err := os.Open(partialIndex.config.Path)
	if err != nil {
		return err
	}

	indexFileStat, err := indexFile.Stat()
	if err != nil {
		return err
	}

	indexStream := partial.NewStream(indexFile, indexFileStat.Size())

	metadata, err := partial.LoadMetadata(indexStream)
	if err != nil {
		return err
	}

	err = metadata.AssertValid(partial.MetadataInput{
		Input:          partialIndex.input,
		IgnoreCase:     *partialIndex.config.IgnoreCase,
		RootNodesCount: len(partialIndex.config.Properties),
	})
	if err != nil {
		return err
	}

	index := make(map[string]*partial.TreeNode)
	for propertyIndex, property := range partialIndex.config.Properties {
		index[property], err = metadata.GetRootNode(propertyIndex)
		if err != nil {
			return err
		}
	}

	partialIndex.index = index

	partialIndex.config.Logger.Infof("Successfully loaded index")

	return nil
}

func (partialIndex *Partial) addValueToIndex(
	index map[string]*partial.TreeNode,
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
	err := root.AddAllSequences(bytes, position)
	if err != nil {
		return err
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

		indexedResults, err := indexedTree.GetSequence([]byte(stringFilter))
		if err != nil {
			return nil, err
		}
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
