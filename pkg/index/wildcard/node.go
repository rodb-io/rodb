package wildcard

import (
	"encoding/binary"
	"rodb.io/pkg/input/record"
)

type TreeNodeOffset int64

type TreeNodeValueOffset int64

const TreeNodeSize int = 56

type TreeNode struct {
	stream              *Stream
	offset              TreeNodeOffset
	valueOffset         TreeNodeValueOffset
	valueLength         int64
	nextSiblingOffset   TreeNodeOffset
	firstChildOffset    TreeNodeOffset
	lastChildOffset     TreeNodeOffset
	firstPositionOffset PositionLinkedListOffset
	lastPositionOffset  PositionLinkedListOffset
}

func NewEmptyTreeNode(stream *Stream) (*TreeNode, error) {
	offset, err := stream.Add(make([]byte, TreeNodeSize))
	if err != nil {
		return nil, err
	}

	return &TreeNode{
		stream:              stream,
		offset:              TreeNodeOffset(offset),
		valueOffset:         0,
		valueLength:         0,
		nextSiblingOffset:   0,
		firstChildOffset:    0,
		lastChildOffset:     0,
		firstPositionOffset: 0,
		lastPositionOffset:  0,
	}, nil
}

func GetTreeNode(
	stream *Stream,
	offset TreeNodeOffset,
) (*TreeNode, error) {
	if offset == 0 {
		return nil, nil
	}

	serialized, err := stream.Get(int64(offset), TreeNodeSize)
	if err != nil {
		return nil, err
	}

	node := &TreeNode{
		stream: stream,
		offset: offset,
	}

	node.Unserialize(serialized)

	return node, nil
}

func (node *TreeNode) NextSibling() (*TreeNode, error) {
	return GetTreeNode(node.stream, node.nextSiblingOffset)
}

func (node *TreeNode) FirstChild() (*TreeNode, error) {
	return GetTreeNode(node.stream, node.firstChildOffset)
}

func (node *TreeNode) LastChild() (*TreeNode, error) {
	return GetTreeNode(node.stream, node.lastChildOffset)
}

func (node *TreeNode) FirstPosition() (*PositionLinkedList, error) {
	return GetPositionLinkedList(node.stream, node.firstPositionOffset)
}

func (node *TreeNode) LastPosition() (*PositionLinkedList, error) {
	return GetPositionLinkedList(node.stream, node.lastPositionOffset)
}

func (node *TreeNode) Value() ([]byte, error) {
	return node.stream.Get(int64(node.valueOffset), int(node.valueLength))
}

func (node *TreeNode) Serialize() []byte {
	buffer := make([]byte, TreeNodeSize)

	binary.BigEndian.PutUint64(buffer[0:8], uint64(node.valueOffset))
	binary.BigEndian.PutUint64(buffer[8:16], uint64(node.valueLength))
	binary.BigEndian.PutUint64(buffer[16:24], uint64(node.nextSiblingOffset))
	binary.BigEndian.PutUint64(buffer[24:32], uint64(node.firstChildOffset))
	binary.BigEndian.PutUint64(buffer[32:40], uint64(node.lastChildOffset))
	binary.BigEndian.PutUint64(buffer[40:48], uint64(node.firstPositionOffset))
	binary.BigEndian.PutUint64(buffer[48:56], uint64(node.lastPositionOffset))

	return buffer
}

func (node *TreeNode) Unserialize(data []byte) {
	node.valueOffset = TreeNodeValueOffset(binary.BigEndian.Uint64(data[0:8]))
	node.valueLength = int64(binary.BigEndian.Uint64(data[8:16]))
	node.nextSiblingOffset = TreeNodeOffset(binary.BigEndian.Uint64(data[16:24]))
	node.firstChildOffset = TreeNodeOffset(binary.BigEndian.Uint64(data[24:32]))
	node.lastChildOffset = TreeNodeOffset(binary.BigEndian.Uint64(data[32:40]))
	node.firstPositionOffset = PositionLinkedListOffset(binary.BigEndian.Uint64(data[40:48]))
	node.lastPositionOffset = PositionLinkedListOffset(binary.BigEndian.Uint64(data[48:56]))
}

func (node *TreeNode) Save() error {
	if err := node.stream.Replace(int64(node.offset), node.Serialize()); err != nil {
		return err
	}

	return nil
}

func (node *TreeNode) AppendChild(child *TreeNode) error {
	if node.firstChildOffset == 0 {
		node.firstChildOffset = child.offset
		node.lastChildOffset = child.offset
		return node.Save()
	} else {
		lastChild, err := node.LastChild()
		if err != nil {
			return err
		}
		lastChild.nextSiblingOffset = child.offset
		if err := lastChild.Save(); err != nil {
			return err
		}

		node.lastChildOffset = child.offset
		return node.Save()
	}
}

// Finds a child that has a common prefix with the given array
func (node *TreeNode) FindChildByPrefix(
	value []byte,
) (
	foundNode *TreeNode,
	commonBytes int64,
	err error,
) {
	child, err := node.FirstChild()
	if err != nil {
		return nil, 0, err
	}

	for child != nil {
		childValue, err := child.Value()
		if err != nil {
			return nil, 0, err
		}

		commonBytes := int64(0)
		for byteIndex := 0; byteIndex < len(childValue) && byteIndex < len(value); byteIndex++ {
			if childValue[byteIndex] == value[byteIndex] {
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
func (node *TreeNode) AppendPositionIfNotExists(position record.Position) error {
	positionNode, err := NewPositionLinkedList(node.stream, position)
	if err != nil {
		return err
	}

	if node.firstPositionOffset == 0 {
		node.firstPositionOffset = positionNode.offset
		node.lastPositionOffset = positionNode.offset
		return node.Save()
	}

	nodeLastPosition, err := node.LastPosition()
	if err != nil {
		return err
	}
	if nodeLastPosition.Position != position {
		nodeLastPosition.nextPositionOffset = positionNode.offset
		if err := nodeLastPosition.Save(); err != nil {
			return err
		}

		node.lastPositionOffset = positionNode.offset
		return node.Save()
	}

	return nil
}

func (node *TreeNode) AddSingleSequence(bytes []byte, position record.Position) error {
	bytesOffset, err := node.stream.Add(bytes)
	if err != nil {
		return err
	}

	return node.addSequence(bytes, bytesOffset, position)
}

func (node *TreeNode) AddAllSequences(bytes []byte, position record.Position) error {
	bytesOffset, err := node.stream.Add(bytes)
	if err != nil {
		return err
	}

	for i := 0; i < len(bytes); i++ {
		if err := node.addSequence(bytes[i:], bytesOffset+int64(i), position); err != nil {
			return err
		}
	}

	return nil
}

func (node *TreeNode) addSequence(
	bytes []byte,
	bytesOffsetInStream int64,
	position record.Position,
) error {
	parentNode := node
	currentPosition := int64(0)
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
			positionList, err := NewPositionLinkedList(node.stream, position)
			if err != nil {
				return err
			}

			newNode, err := NewEmptyTreeNode(node.stream)
			if err != nil {
				return err
			}

			newNode.valueOffset = TreeNodeValueOffset(bytesOffsetInStream + int64(currentPosition))
			newNode.valueLength = int64(len(remainingBytes))
			newNode.nextSiblingOffset = 0
			newNode.firstChildOffset = 0
			newNode.lastChildOffset = 0
			newNode.firstPositionOffset = positionList.offset
			newNode.lastPositionOffset = positionList.offset
			if err := newNode.Save(); err != nil {
				return err
			}

			if err := parentNode.AppendChild(newNode); err != nil {
				return err
			}
			break
		}

		// Only matching a part of the node. We need to split it and continue
		// proceeding with the parent (which has a prefix matching)
		if commonBytes < existingNode.valueLength {
			existingNodeFirstPosition, err := existingNode.FirstPosition()
			if err != nil {
				return err
			}

			newChildFirstPosition, newChildLastPosition, err := existingNodeFirstPosition.Copy()
			if err != nil {
				return err
			}
			newChild, err := NewEmptyTreeNode(node.stream)
			if err != nil {
				return err
			}

			newChild.valueOffset = TreeNodeValueOffset(int64(existingNode.valueOffset) + int64(commonBytes))
			newChild.valueLength = existingNode.valueLength - commonBytes
			newChild.nextSiblingOffset = 0
			newChild.firstChildOffset = existingNode.firstChildOffset
			newChild.lastChildOffset = existingNode.lastChildOffset
			newChild.firstPositionOffset = newChildFirstPosition.offset
			newChild.lastPositionOffset = newChildLastPosition.offset
			if err := newChild.Save(); err != nil {
				return err
			}

			existingNode.firstChildOffset = newChild.offset
			existingNode.lastChildOffset = newChild.offset
			existingNode.valueLength = commonBytes
			if err := existingNode.Save(); err != nil {
				return err
			}
		}

		if err := existingNode.AppendPositionIfNotExists(position); err != nil {
			return err
		}
		parentNode = existingNode
		currentPosition += commonBytes
	}

	return nil
}

func (node *TreeNode) GetSequence(bytes []byte) (*PositionLinkedList, error) {
	parentNode := node
	currentPosition := int64(0)
	for currentPosition < int64(len(bytes)) {
		currentNode, commonBytes, err := parentNode.FindChildByPrefix(bytes[currentPosition:])
		if err != nil {
			return nil, err
		}
		if currentNode == nil {
			return nil, nil
		}
		if commonBytes < currentNode.valueLength && commonBytes+currentPosition < int64(len(bytes)) {
			return nil, nil
		}

		currentPosition += currentNode.valueLength
		parentNode = currentNode
	}

	return parentNode.FirstPosition()
}
