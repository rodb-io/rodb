package index

import (
	"fmt"
	"reflect"
	"rodb.io/pkg/config"
	"rodb.io/pkg/input"
	"rodb.io/pkg/record"
	"rodb.io/pkg/util"
)

type memoryPartialTrieNode struct {
	value rune
	nextSibling *memoryPartialTrieNode
	firstChild *memoryPartialTrieNode
	lastChild *memoryPartialTrieNode
	firstPosition *record.PositionLinkedList
	lastPosition *record.PositionLinkedList
}

func (node *memoryPartialTrieNode) appendChild(child *memoryPartialTrieNode) {
	if node.firstChild == nil {
		node.firstChild = child
		node.lastChild = child
	} else {
		node.lastChild.nextSibling = child
		node.lastChild = child
	}
}

func (node *memoryPartialTrieNode) findChildByValue(value rune) *memoryPartialTrieNode {
	for child := node.firstChild; child != nil; child = child.nextSibling {
		if child.value == value {
			return child
		}
	}

	return nil
}

func (node *memoryPartialTrieNode) appendPositionIfNotExists(position record.Position) {
	positionNode := &record.PositionLinkedList{
		Position: position,
	}

	if node.firstPosition == nil {
		node.firstPosition = positionNode
		node.lastPosition = positionNode
	} else {
		node.lastPosition.NextPosition = positionNode
		node.lastPosition = positionNode
	}
}

type MemoryPartial struct {
	config *config.MemoryPartialIndex
	input  input.Input
	index  map[string]*memoryPartialTrieNode
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
	index := make(map[string]*memoryPartialTrieNode)
	for _, property := range mp.config.Properties {
		index[property] = &memoryPartialTrieNode{}
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
	index map[string]*memoryPartialTrieNode,
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
	previousNode := root

	runes := []rune(stringValue)
	for _, currentRune := range runes {
		if existingNode := previousNode.findChildByValue(currentRune); existingNode == nil {
			newNode := &memoryPartialTrieNode{
				value: currentRune,
			}
			previousNode.appendChild(newNode)

			if existingRootNode := root.findChildByValue(currentRune); existingRootNode == nil {
				existingRootNode.appendChild(newNode)
			} else {
				existingRootNode.appendPositionIfNotExists(position)
			}

			previousNode = newNode
		} else {
			existingNode.appendPositionIfNotExists(position)
			previousNode = existingNode
		}
	}

	return nil
}

func (mp *MemoryPartial) GetRecordPositions(
	input input.Input,
	filters map[string]interface{},
) (record.PositionIterator, error) {
	return nil, nil
}

func (mp *MemoryPartial) Close() error {
	return nil
}
