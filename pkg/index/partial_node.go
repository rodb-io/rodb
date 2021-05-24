package index

import (
	"fmt"
	"io"
	"rodb.io/pkg/record"
)

type readerAtWriterAt interface{
	io.ReaderAt
	io.WriterAt
}

type partialIndexTreeNodeOffset *int64

const serializedPartialIndexTreeNodeSize int = 42 // TODO

type partialIndexTreeNode struct {
	stream        readerAtWriterAt
	streamSize    *int64
	offset        partialIndexTreeNodeOffset
	value         []byte // TODO
	nextSiblingOffset   partialIndexTreeNodeOffset
	firstChildOffset    partialIndexTreeNodeOffset
	lastChildOffset     partialIndexTreeNodeOffset

	// TODO
	firstPosition *record.PositionLinkedList
	lastPosition  *record.PositionLinkedList
}

func createEmptyPartialIndexTreeNode(
	stream readerAtWriterAt,
	streamSize *int64,
) (*partialIndexTreeNode, error) {
	rootNode := &partialIndexTreeNode{
		stream:        stream,
		streamSize:    streamSize,
		offset:        nil,
		value:         []byte{},
		nextSiblingOffset:   nil,
		firstChildOffset:    nil,
		lastChildOffset:     nil,
		firstPosition: nil,
		lastPosition:  nil,
	}

	err := rootNode.save()
	if err != nil {
		return nil, err
	}

	return rootNode, nil
}

func getPartialIndexTreeNode(
	stream readerAtWriterAt,
	streamSize *int64,
	offset partialIndexTreeNodeOffset,
) (*partialIndexTreeNode, error) {
	if offset == nil {
		return nil, nil
	}

	serialized := make([]byte, 0, serializedPartialIndexTreeNodeSize)
	size, err := stream.ReadAt(serialized, *offset)
	if err != nil {
		return nil, err
	}
	if size != serializedPartialIndexTreeNodeSize {
		return nil, fmt.Errorf("Expected to read %v bytes, got %v", serializedPartialIndexTreeNodeSize, size)
	}

	node := &partialIndexTreeNode{
		stream:        stream,
		streamSize:    streamSize,
		offset:        offset,
	}

	// TODO unserialize in node

	return node, nil
}

func (node *partialIndexTreeNode) nextSibling() (*partialIndexTreeNode, error) {
	return getPartialIndexTreeNode(node.stream, node.streamSize, node.nextSiblingOffset)
}

func (node *partialIndexTreeNode) firstChild() (*partialIndexTreeNode, error) {
	return getPartialIndexTreeNode(node.stream, node.streamSize, node.firstChildOffset)
}

func (node *partialIndexTreeNode) lastChild() (*partialIndexTreeNode, error) {
	return getPartialIndexTreeNode(node.stream, node.streamSize, node.lastChildOffset)
}

func (node *partialIndexTreeNode) save() error {
	serialized := []byte{} // TODO serialize node

	if node.offset == nil {
		newOffset := *node.streamSize
		size, err := node.stream.WriteAt(serialized, newOffset)
		if err != nil {
			return err
		}
		if size != serializedPartialIndexTreeNodeSize {
			return fmt.Errorf("Expected to write %v bytes, wrote %v", serializedPartialIndexTreeNodeSize, size)
		}
		node.offset = partialIndexTreeNodeOffset(&newOffset)
		*node.streamSize += int64(serializedPartialIndexTreeNodeSize)
	} else {
		size, err := node.stream.WriteAt(serialized, *node.offset)
		if err != nil {
			return err
		}
		if size != serializedPartialIndexTreeNodeSize {
			return fmt.Errorf("Expected to write %v bytes, wrote %v", serializedPartialIndexTreeNodeSize, size)
		}
	}

	return nil
}

func (node *partialIndexTreeNode) appendChild(child *partialIndexTreeNode) error {
	if node.firstChildOffset == nil {
		node.firstChildOffset = child.offset
		node.lastChildOffset = child.offset
		return node.save()
	} else {
		lastChild, err := node.lastChild()
		if err != nil {
			return err
		}
		lastChild.nextSiblingOffset = child.offset
		lastChild.save()

		node.lastChildOffset = child.offset
		return node.save()
	}
}

// Finds a child that has a common prefix with the given array
func (node *partialIndexTreeNode) findChildByPrefix(
	value []byte,
) (
	foundNode *partialIndexTreeNode,
	commonBytes int,
	err error,
) {
	child, err := node.firstChild()
	if err != nil {
		return nil, 0, err
	}

	for child != nil {
		commonBytes := 0
		for byteIndex := 0; byteIndex < len(child.value) && byteIndex < len(value); byteIndex++ {
			if child.value[byteIndex] == value[byteIndex] {
				commonBytes++
			} else {
				break
			}
		}

		if commonBytes > 0 {
			return child, commonBytes, nil
		}

		child, err = child.nextSibling()
		if err != nil {
			return nil, 0, err
		}
	}

	return nil, 0, nil
}

// This only checks if the position already exists at the end of the list
func (node *partialIndexTreeNode) appendPositionIfNotExists(position record.Position) {
	positionNode := &record.PositionLinkedList{
		Position:     position,
		NextPosition: nil,
	}

	if node.firstPosition == nil {
		node.firstPosition = positionNode
		node.lastPosition = positionNode
	} else if node.lastPosition.Position != position {
		node.lastPosition.NextPosition = positionNode
		node.lastPosition = positionNode
	}
}

func (node *partialIndexTreeNode) addSequence(bytes []byte, position record.Position) error {
	parentNode := node
	currentPosition := 0
	for {
		remainingBytes := bytes[currentPosition:]
		if len(remainingBytes) == 0 {
			return nil
		}

		existingNode, commonBytes, err := parentNode.findChildByPrefix(remainingBytes)
		if err != nil {
			return err
		}

		// No node for the remaining string: adding a node with it
		if existingNode == nil {
			positionList := &record.PositionLinkedList{
				Position: position,
			}
			newNode, err := createEmptyPartialIndexTreeNode(node.stream, node.streamSize)
			if err != nil {
				return err
			}

			newNode.value =          remainingBytes
			newNode.nextSiblingOffset =    nil
			newNode.firstChildOffset =     nil
			newNode.lastChildOffset =      nil
			newNode.firstPosition =  positionList
			newNode.lastPosition =   positionList
			err = newNode.save()
			if err != nil {
				return err
			}

			parentNode.appendChild(newNode)
			break
		}

		// Only matching a part of the node. We need to split it and continue
		// proceeding with the parent (which has a prefix matching)
		if commonBytes < len(existingNode.value) {
			newChildFirstPosition, newChildLastPosition := existingNode.firstPosition.Copy()
			newChild, err := createEmptyPartialIndexTreeNode(node.stream, node.streamSize)
			if err != nil {
				return err
			}

			newChild.value =         existingNode.value[commonBytes:]
			newChild.nextSiblingOffset =   nil
			newChild.firstChildOffset =    existingNode.firstChildOffset
			newChild.lastChildOffset =     existingNode.lastChildOffset
			newChild.firstPosition = newChildFirstPosition
			newChild.lastPosition =  newChildLastPosition
			err = newChild.save()
			if err != nil {
				return err
			}

			existingNode.firstChildOffset = newChild.offset
			existingNode.lastChildOffset = newChild.offset
			existingNode.value = existingNode.value[:commonBytes]
			err = existingNode.save()
			if err != nil {
				return err
			}
		}

		existingNode.appendPositionIfNotExists(position)
		parentNode = existingNode
		currentPosition += commonBytes
	}

	return nil
}

func (node *partialIndexTreeNode) getSequence(bytes []byte) (*record.PositionLinkedList, error) {
	parentNode := node
	currentPosition := 0
	for currentPosition < len(bytes) {
		currentNode, commonBytes, err := parentNode.findChildByPrefix(bytes[currentPosition:])
		if err != nil {
			return nil, err
		}
		if currentNode == nil {
			return nil, nil
		}
		if commonBytes < len(currentNode.value) && commonBytes+currentPosition < len(bytes) {
			return nil, nil
		}

		currentPosition += len(currentNode.value)
		parentNode = currentNode
	}

	return parentNode.firstPosition, nil
}
