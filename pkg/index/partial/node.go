package partial

import (
	"rodb.io/pkg/record"
)

type TreeNodeOffset *int64

const TreeNodeSize int = 42 // TODO

type TreeNode struct {
	stream        *Stream
	offset        TreeNodeOffset
	value         []byte // TODO
	nextSiblingOffset   TreeNodeOffset
	firstChildOffset    TreeNodeOffset
	lastChildOffset     TreeNodeOffset
	firstPositionOffset PositionLinkedListOffset
	lastPositionOffset  PositionLinkedListOffset
}

func NewEmptyTreeNode(stream *Stream) (*TreeNode, error) {
	node := &TreeNode{
		stream:        stream,
		offset:        nil,
		value:         []byte{},
		nextSiblingOffset:   nil,
		firstChildOffset:    nil,
		lastChildOffset:     nil,
		firstPositionOffset: nil,
		lastPositionOffset:  nil,
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
	if offset == nil {
		return nil, nil
	}

	serialized, err := stream.Get(TreeNodeSize, *offset)
	if err != nil {
		return nil, err
	}

	node := &TreeNode{
		stream:        stream,
		offset:        offset,
	}

	// TODO unserialize in node

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

func (node *TreeNode) Save() error {
	serialized := []byte{} // TODO serialize node with size TreeNodeSize

	if node.offset == nil {
		newOffset, err := node.stream.Add(serialized)
		if err != nil {
			return err
		}
		node.offset = TreeNodeOffset(&newOffset)
	} else {
		err := node.stream.Replace(*node.offset, serialized)
		if err != nil {
			return err
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
func (node *TreeNode) AppendPositionIfNotExists(position record.Position) error {
	positionNode, err := NewPositionLinkedList(node.stream, position)
	if err != nil {
		return err
	}

	if node.firstPositionOffset == nil {
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
			newNode, err := NewEmptyTreeNode(node.stream)
			if err != nil {
				return err
			}

			newNode.value =          remainingBytes
			newNode.nextSiblingOffset =    nil
			newNode.firstChildOffset =     nil
			newNode.lastChildOffset =      nil
			newNode.firstPositionOffset =  positionList.offset
			newNode.lastPositionOffset =   positionList.offset
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

			newChild.value =         existingNode.value[commonBytes:]
			newChild.nextSiblingOffset =   nil
			newChild.firstChildOffset =    existingNode.firstChildOffset
			newChild.lastChildOffset =     existingNode.lastChildOffset
			newChild.firstPositionOffset = newChildFirstPosition.offset
			newChild.lastPositionOffset =  newChildLastPosition.offset
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

	return parentNode.FirstPosition()
}
