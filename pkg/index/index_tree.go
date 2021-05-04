package index

import (
	"rodb.io/pkg/record"
)

type partialIndexTreeNode struct {
	value         []byte
	nextSibling   *partialIndexTreeNode
	firstChild    *partialIndexTreeNode
	lastChild     *partialIndexTreeNode
	firstPosition *record.PositionLinkedList
	lastPosition  *record.PositionLinkedList
}

func (node *partialIndexTreeNode) appendChild(child *partialIndexTreeNode) {
	if node.firstChild == nil {
		node.firstChild = child
		node.lastChild = child
	} else {
		node.lastChild.nextSibling = child
		node.lastChild = child
	}
}

// Finds a child that has a common prefix with the given array
func (node *partialIndexTreeNode) findChildByPrefix(
	value []byte,
) (
	foundNode *partialIndexTreeNode,
	commonBytes int,
) {
	for child := node.firstChild; child != nil; child = child.nextSibling {
		commonBytes := 0
		for byteIndex := 0; byteIndex < len(child.value) && byteIndex < len(value); byteIndex++ {
			if child.value[byteIndex] == value[byteIndex] {
				commonBytes++
			} else {
				break
			}
		}

		if commonBytes > 0 {
			return child, commonBytes
		}
	}

	return nil, 0
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

func (node *partialIndexTreeNode) addSequence(bytes []byte, position record.Position) {
	parentNode := node
	currentPosition := 0
	for {
		remainingBytes := bytes[currentPosition:]
		if len(remainingBytes) == 0 {
			return
		}

		existingNode, commonBytes := parentNode.findChildByPrefix(remainingBytes)

		// No node for the remaining string: adding a node with it
		if existingNode == nil {
			positionList := &record.PositionLinkedList{
				Position: position,
			}
			newNode := &partialIndexTreeNode{
				value:         remainingBytes,
				nextSibling:   nil,
				firstChild:    nil,
				lastChild:     nil,
				firstPosition: positionList,
				lastPosition:  positionList,
			}
			parentNode.appendChild(newNode)
			break
		}

		// Only matching a part of the node. We need to split it and continue
		// proceeding with the parent (which has a prefix matching)
		if commonBytes < len(existingNode.value) {
			newChildFirstPosition, newChildLastPosition := existingNode.firstPosition.Copy()
			newChild := &partialIndexTreeNode{
				value:         existingNode.value[commonBytes:],
				nextSibling:   nil,
				firstChild:    existingNode.firstChild,
				lastChild:     existingNode.lastChild,
				firstPosition: newChildFirstPosition,
				lastPosition:  newChildLastPosition,
			}

			existingNode.firstChild = newChild
			existingNode.lastChild = newChild
			existingNode.value = existingNode.value[:commonBytes]
		}

		existingNode.appendPositionIfNotExists(position)
		parentNode = existingNode
		currentPosition += commonBytes
	}
}

func (node *partialIndexTreeNode) getSequence(bytes []byte) *record.PositionLinkedList {
	parentNode := node
	currentPosition := 0
	for currentPosition < len(bytes) {
		currentNode, commonBytes := parentNode.findChildByPrefix(bytes[currentPosition:])
		if currentNode == nil {
			return nil
		}
		if commonBytes < len(currentNode.value) && commonBytes+currentPosition < len(bytes) {
			return nil
		}

		currentPosition += len(currentNode.value)
		parentNode = currentNode
	}

	return parentNode.firstPosition
}
