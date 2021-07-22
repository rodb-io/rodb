package index

import (
	"fmt"
	"os"
	"reflect"
	wildcardPackage "rodb.io/pkg/index/wildcard"
	"rodb.io/pkg/input"
	"rodb.io/pkg/input/record"
	"rodb.io/pkg/util"
	"strings"
)

type Wildcard struct {
	config *WildcardConfig
	input  input.Input
	index  map[string]*wildcardPackage.TreeNode
}

func NewWildcard(
	config *WildcardConfig,
	inputs input.List,
) (*Wildcard, error) {
	input, inputExists := inputs[config.Input]
	if !inputExists {
		return nil, fmt.Errorf("Input '%v' not found in inputs list.", config.Input)
	}

	wildcard := &Wildcard{
		config: config,
		input:  input,
	}

	_, err := os.Stat(wildcard.config.Path)
	if os.IsNotExist(err) {
		if err := wildcard.createIndex(); err != nil {
			return nil, fmt.Errorf("Error while creating the index: %w", err)
		}
	} else if err != nil {
		return nil, err
	} else {
		if err := wildcard.loadIndex(); err != nil {
			return nil, fmt.Errorf("Error while loading the index: %w", err)
		}
	}

	return wildcard, nil
}

func (wildcard *Wildcard) Name() string {
	return wildcard.config.Name
}

func (wildcard *Wildcard) createIndex() error {
	indexFile, err := os.Create(wildcard.config.Path)
	if err != nil {
		return err
	}

	indexStream := wildcardPackage.NewStream(indexFile, 0)

	metadata, err := wildcardPackage.NewMetadata(indexStream, wildcardPackage.MetadataInput{
		Input:          wildcard.input,
		IgnoreCase:     *wildcard.config.IgnoreCase,
		RootNodesCount: len(wildcard.config.Properties),
	})
	if err != nil {
		return err
	}

	index := make(map[string]*wildcardPackage.TreeNode)
	for propertyIndex, property := range wildcard.config.Properties {
		index[property], err = wildcardPackage.NewEmptyTreeNode(indexStream)
		if err != nil {
			return err
		}
		if err := index[property].Save(); err != nil {
			return err
		}

		metadata.SetRootNode(propertyIndex, index[property])
	}

	if err := metadata.Save(); err != nil {
		return err
	}

	updateProgress := util.TrackProgress(wildcard.input, wildcard.config.Logger)

	inputIterator, end, err := wildcard.input.IterateAll()
	if err != nil {
		return err
	}
	defer func() {
		if err := end(); err != nil {
			wildcard.config.Logger.Errorf("Error while closing the input iterator: %v", err)
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

		for _, property := range wildcard.config.Properties {
			value, err := record.Get(property)
			if err != nil {
				return err
			}

			if value != nil {
				value = reflect.ValueOf(value).Interface()
			}

			if err := wildcard.addValueToIndex(index, property, value, record.Position()); err != nil {
				return fmt.Errorf("Cannot index the property '%v': ", property)
			}
		}
	}

	metadata.SetCompleted(true)
	if err := metadata.Save(); err != nil {
		return err
	}

	if err := indexStream.Flush(); err != nil {
		return err
	}

	wildcard.index = index

	indexStat, err := indexFile.Stat()
	if err != nil {
		return err
	}

	wildcard.config.Logger.WithField("indexSize", indexStat.Size()).Infof("Successfully finished indexing")

	return nil
}

func (wildcard *Wildcard) loadIndex() error {
	indexFile, err := os.Open(wildcard.config.Path)
	if err != nil {
		return err
	}

	indexFileStat, err := indexFile.Stat()
	if err != nil {
		return err
	}

	indexStream := wildcardPackage.NewStream(indexFile, indexFileStat.Size())

	metadata, err := wildcardPackage.LoadMetadata(indexStream)
	if err != nil {
		return err
	}

	input := wildcardPackage.MetadataInput{
		Input:          wildcard.input,
		IgnoreCase:     *wildcard.config.IgnoreCase,
		RootNodesCount: len(wildcard.config.Properties),
	}
	if err := metadata.AssertValid(input); err != nil {
		return err
	}

	index := make(map[string]*wildcardPackage.TreeNode)
	for propertyIndex, property := range wildcard.config.Properties {
		index[property], err = metadata.GetRootNode(propertyIndex)
		if err != nil {
			return err
		}
	}

	wildcard.index = index

	wildcard.config.Logger.Infof("Successfully loaded index")

	return nil
}

func (wildcard *Wildcard) addValueToIndex(
	index map[string]*wildcardPackage.TreeNode,
	property string,
	value interface{},
	position record.Position,
) error {
	if valueArray, valueIsArray := value.([]interface{}); valueIsArray {
		for _, valueArrayValue := range valueArray {
			if err := wildcard.addValueToIndex(index, property, valueArrayValue, position); err != nil {
				return err
			}
		}
	}

	stringValue, valueIsString := value.(string)
	if !valueIsString {
		return fmt.Errorf("Cannot index the value '%v' from property '%v' because it is not a string.", value, property)
	}

	if wildcard.config.ShouldIgnoreCase() {
		stringValue = strings.ToLower(stringValue)
	}

	root := index[property]
	bytes := []byte(stringValue)
	if err := root.AddAllSequences(bytes, position); err != nil {
		return err
	}

	return nil
}

func (wildcard *Wildcard) GetRecordPositions(
	input input.Input,
	filters map[string]interface{},
) (record.PositionIterator, error) {
	if input != wildcard.input {
		return nil, fmt.Errorf("This index does not handle the input '%v'.", input.Name())
	}

	if len(filters) == 0 {
		return nil, fmt.Errorf("This index requires at least one filter.")
	}

	individualFiltersResults := make([]record.PositionIterator, 0, len(filters))
	for propertyName, filter := range filters {
		if !wildcard.config.DoesHandleProperty(propertyName) {
			return nil, fmt.Errorf("This index does not handle the property '%v'.", propertyName)
		}

		indexedTree, foundIndexedValues := wildcard.index[propertyName]
		if !foundIndexedValues {
			return record.EmptyIterator, nil
		}

		stringFilter, filterIsString := filter.(string)
		if !filterIsString {
			return nil, fmt.Errorf("Cannot filter the value '%v' from property '%v' because it is not a string.", filter, propertyName)
		}

		if wildcard.config.ShouldIgnoreCase() {
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

func (wildcard *Wildcard) Close() error {
	return nil
}
