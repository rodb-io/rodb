package partial

import (
	"bytes"
	"encoding/binary"
	"rodb.io/pkg/record"
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
	node := &TreeNode{
		stream:              stream,
		offset:              0,
		valueOffset:         0,
		valueLength:         0,
		nextSiblingOffset:   0,
		firstChildOffset:    0,
		lastChildOffset:     0,
		firstPositionOffset: 0,
		lastPositionOffset:  0,
	}

	err := node.Save()
	if err != nil {
		return nil, err
	}

	return node, nil
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

	err = node.Unserialize(serialized)
	if err != nil {
		return nil, err
	}

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

func (node *TreeNode) Serialize() ([]byte, error) {
	var err error
	buffer := &bytes.Buffer{}

	if err = binary.Write(buffer, binary.BigEndian, node.valueOffset); err != nil {
		return nil, err
	}
	if err = binary.Write(buffer, binary.BigEndian, node.valueLength); err != nil {
		return nil, err
	}
	if err = binary.Write(buffer, binary.BigEndian, node.nextSiblingOffset); err != nil {
		return nil, err
	}
	if err = binary.Write(buffer, binary.BigEndian, node.firstChildOffset); err != nil {
		return nil, err
	}
	if err = binary.Write(buffer, binary.BigEndian, node.lastChildOffset); err != nil {
		return nil, err
	}
	if err = binary.Write(buffer, binary.BigEndian, node.firstPositionOffset); err != nil {
		return nil, err
	}
	if err = binary.Write(buffer, binary.BigEndian, node.lastPositionOffset); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (node *TreeNode) Unserialize(data []byte) error {
	var err error
	buffer := bytes.NewBuffer(data)

	if err = binary.Read(buffer, binary.BigEndian, &node.valueOffset); err != nil {
		return err
	}
	if err = binary.Read(buffer, binary.BigEndian, &node.valueLength); err != nil {
		return err
	}
	if err = binary.Read(buffer, binary.BigEndian, &node.nextSiblingOffset); err != nil {
		return err
	}
	if err = binary.Read(buffer, binary.BigEndian, &node.firstChildOffset); err != nil {
		return err
	}
	if err = binary.Read(buffer, binary.BigEndian, &node.lastChildOffset); err != nil {
		return err
	}
	if err = binary.Read(buffer, binary.BigEndian, &node.firstPositionOffset); err != nil {
		return err
	}
	if err = binary.Read(buffer, binary.BigEndian, &node.lastPositionOffset); err != nil {
		return err
	}

	return nil
}

func (node *TreeNode) Save() error {
	serialized, err := node.Serialize()
	if err != nil {
		return err
	}

	if node.offset == 0 {
		newOffset, err := node.stream.Add(serialized)
		if err != nil {
			return err
		}
		node.offset = TreeNodeOffset(newOffset)
	} else {
		err := node.stream.Replace(int64(node.offset), serialized)
		if err != nil {
			return err
		}
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
		err = nodeLastPosition.Save()
		if err != nil {
			return err
		}

		node.lastPositionOffset = positionNode.offset
		return node.Save()
	}

	return nil
}

func (node *TreeNode) AddSequence(bytes []byte, position record.Position) error {
	// TODO should not add this everytime
	bytesOffset, err := node.stream.Add(bytes)
	if err != nil {
		return err
	}

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

			newNode.valueOffset = TreeNodeValueOffset(bytesOffset + int64(currentPosition))
			newNode.valueLength = int64(len(remainingBytes))
			newNode.nextSiblingOffset = 0
			newNode.firstChildOffset = 0
			newNode.lastChildOffset = 0
			newNode.firstPositionOffset = positionList.offset
			newNode.lastPositionOffset = positionList.offset
			err = newNode.Save()
			if err != nil {
				return err
			}

			err = parentNode.AppendChild(newNode)
			if err != nil {
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
			err = newChild.Save()
			if err != nil {
				return err
			}

			existingNode.firstChildOffset = newChild.offset
			existingNode.lastChildOffset = newChild.offset
			existingNode.valueLength = commonBytes
			err = existingNode.Save()
			if err != nil {
				return err
			}
		}

		err = existingNode.AppendPositionIfNotExists(position)
		if err != nil {
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
