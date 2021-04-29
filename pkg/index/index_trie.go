package index

import (
	"rodb.io/pkg/record"
)

type partialIndexTrieNode struct {
	value         rune
	nextSibling   *partialIndexTrieNode
	firstChild    *partialIndexTrieNode
	lastChild     *partialIndexTrieNode
	firstPosition *record.PositionLinkedList
	lastPosition  *record.PositionLinkedList
}

func (node *partialIndexTrieNode) appendChild(child *partialIndexTrieNode) {
	if node.firstChild == nil {
		node.firstChild = child
		node.lastChild = child
	} else {
		node.lastChild.nextSibling = child
		node.lastChild = child
	}
}

func (node *partialIndexTrieNode) findChildByValue(value rune) *partialIndexTrieNode {
	for child := node.firstChild; child != nil; child = child.nextSibling {
		if child.value == value {
			return child
		}
	}

	return nil
}

// This only checks if the position already exists at the end of the list
func (node *partialIndexTrieNode) appendPositionIfNotExists(position record.Position) {
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

func (node *partialIndexTrieNode) addSequence(runes []rune, position record.Position) {
	parentNode := node
	for _, currentRune := range runes {
		if existingNode := parentNode.findChildByValue(currentRune); existingNode == nil {
			positionList := &record.PositionLinkedList{
				Position: position,
			}
			newNode := &partialIndexTrieNode{
				value:         currentRune,
				nextSibling:   nil,
				firstChild:    nil,
				lastChild:     nil,
				firstPosition: positionList,
				lastPosition:  positionList,
			}
			parentNode.appendChild(newNode)
			parentNode = newNode
		} else {
			existingNode.appendPositionIfNotExists(position)
			parentNode = existingNode
		}
	}
}

func (node *partialIndexTrieNode) getSequence(runes []rune) *record.PositionLinkedList {
	parentNode := node
	for _, currentRune := range runes {
		currentNode := parentNode.findChildByValue(currentRune)
		if currentNode == nil {
			return nil
		}

		parentNode = currentNode
	}

	return parentNode.firstPosition
}
