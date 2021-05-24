package partial

import (
	"fmt"
	"io"
	"rodb.io/pkg/record"
)

type ReaderAtWriterAt interface{
	io.ReaderAt
	io.WriterAt
}

type TreeNodeOffset *int64

const TreeNodeSize int = 42 // TODO

type TreeNode struct {
	stream        ReaderAtWriterAt
	streamSize    *int64
	offset        TreeNodeOffset
	value         []byte // TODO
	nextSiblingOffset   TreeNodeOffset
	firstChildOffset    TreeNodeOffset
	lastChildOffset     TreeNodeOffset

	// TODO
	firstPosition *PositionLinkedList
	lastPosition  *PositionLinkedList
}

func NewEmptyTreeNode(
	stream ReaderAtWriterAt,
	streamSize *int64,
) (*TreeNode, error) {
	rootNode := &TreeNode{
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

	err := rootNode.Save()
	if err != nil {
		return nil, err
	}

	return rootNode, nil
}

func GetTreeNode(
	stream ReaderAtWriterAt,
	streamSize *int64,
	offset TreeNodeOffset,
) (*TreeNode, error) {
	if offset == nil {
		return nil, nil
	}

	serialized := make([]byte, 0, TreeNodeSize)
	size, err := stream.ReadAt(serialized, *offset)
	if err != nil {
		return nil, err
	}
	if size != TreeNodeSize {
		return nil, fmt.Errorf("Expected to read %v bytes, got %v", TreeNodeSize, size)
	}

	node := &TreeNode{
		stream:        stream,
		streamSize:    streamSize,
		offset:        offset,
	}

	// TODO unserialize in node

	return node, nil
}

func (node *TreeNode) NextSibling() (*TreeNode, error) {
	return GetTreeNode(node.stream, node.streamSize, node.nextSiblingOffset)
}

func (node *TreeNode) FirstChild() (*TreeNode, error) {
	return GetTreeNode(node.stream, node.streamSize, node.firstChildOffset)
}

func (node *TreeNode) LastChild() (*TreeNode, error) {
	return GetTreeNode(node.stream, node.streamSize, node.lastChildOffset)
}

func (node *TreeNode) Save() error {
	serialized := []byte{} // TODO serialize node

	if node.offset == nil {
		newOffset := *node.streamSize
		size, err := node.stream.WriteAt(serialized, newOffset)
		if err != nil {
			return err
		}
		if size != TreeNodeSize {
			return fmt.Errorf("Expected to write %v bytes, wrote %v", TreeNodeSize, size)
		}
		node.offset = TreeNodeOffset(&newOffset)
		*node.streamSize += int64(TreeNodeSize)
	} else {
		size, err := node.stream.WriteAt(serialized, *node.offset)
		if err != nil {
			return err
		}
		if size != TreeNodeSize {
			return fmt.Errorf("Expected to write %v bytes, wrote %v", TreeNodeSize, size)
		}
	}

	return nil
}

func (node *TreeNode) AppendChild(child *TreeNode) error {
	if node.firstChildOffset == nil {
		node.firstChildOffset = child.offset
		node.lastChildOffset = child.offset
		return node.Save()
	} else {
		lastChild, err := node.LastChild()
		if err != nil {
			return err
		}
		lastChild.nextSiblingOffset = child.offset
		lastChild.Save()

		node.lastChildOffset = child.offset
		return node.Save()
	}
}

// Finds a child that has a common prefix with the given array
func (node *TreeNode) FindChildByPrefix(
	value []byte,
) (
	foundNode *TreeNode,
	commonBytes int,
	err error,
) {
	child, err := node.FirstChild()
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

		child, err = child.NextSibling()
		if err != nil {
			return nil, 0, err
		}
	}

	return nil, 0, nil
}

// This only checks if the position already exists at the end of the list
func (node *TreeNode) AppendPositionIfNotExists(position record.Position) {
	positionNode := &PositionLinkedList{
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

func (node *TreeNode) AddSequence(bytes []byte, position record.Position) error {
	parentNode := node
	currentPosition := 0
	for {
		remainingBytes := bytes[currentPosition:]
		if len(remainingBytes) == 0 {
			return nil
		}

		existingNode, commonBytes, err := parentNode.FindChildByPrefix(remainingBytes)
		if err != nil {
			return err
		}

		// No node for the remaining string: adding a node with it
		if existingNode == nil {
			positionList := &PositionLinkedList{
				Position: position,
			}
			newNode, err := NewEmptyTreeNode(node.stream, node.streamSize)
			if err != nil {
				return err
			}

			newNode.value =          remainingBytes
			newNode.nextSiblingOffset =    nil
			newNode.firstChildOffset =     nil
			newNode.lastChildOffset =      nil
			newNode.firstPosition =  positionList
			newNode.lastPosition =   positionList
			err = newNode.Save()
			if err != nil {
				return err
			}

			parentNode.AppendChild(newNode)
			break
		}

		// Only matching a part of the node. We need to split it and continue
		// proceeding with the parent (which has a prefix matching)
		if commonBytes < len(existingNode.value) {
			newChildFirstPosition, newChildLastPosition := existingNode.firstPosition.Copy()
			newChild, err := NewEmptyTreeNode(node.stream, node.streamSize)
			if err != nil {
				return err
			}

			newChild.value =         existingNode.value[commonBytes:]
			newChild.nextSiblingOffset =   nil
			newChild.firstChildOffset =    existingNode.firstChildOffset
			newChild.lastChildOffset =     existingNode.lastChildOffset
			newChild.firstPosition = newChildFirstPosition
			newChild.lastPosition =  newChildLastPosition
			err = newChild.Save()
			if err != nil {
				return err
			}

			existingNode.firstChildOffset = newChild.offset
			existingNode.lastChildOffset = newChild.offset
			existingNode.value = existingNode.value[:commonBytes]
			err = existingNode.Save()
			if err != nil {
				return err
			}
		}

		existingNode.AppendPositionIfNotExists(position)
		parentNode = existingNode
		currentPosition += commonBytes
	}

	return nil
}

func (node *TreeNode) GetSequence(bytes []byte) (*PositionLinkedList, error) {
	parentNode := node
	currentPosition := 0
	for currentPosition < len(bytes) {
		currentNode, commonBytes, err := parentNode.FindChildByPrefix(bytes[currentPosition:])
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
