package index

import (
	"rodb.io/pkg/record"
	"testing"
)

func PartialIndexTreeNodeAppendChild(t *testing.T) {
	t.Run("no childs", func(t *testing.T) {
		node := &partialIndexTrieNode{
			value: 'A',
		}
		newChild := &partialIndexTrieNode{
			value: 'B',
		}
		node.appendChild(newChild)

		if node.firstChild != newChild {
			t.Errorf("Expected to have B as first child, got %+v", node.firstChild)
		}
		if node.lastChild != newChild {
			t.Errorf("Expected to have B as last child, got %+v", node.lastChild)
		}
	})
	t.Run("already have child", func(t *testing.T) {
		child := &partialIndexTrieNode{
			value: 'B',
		}
		node := &partialIndexTrieNode{
			value:      'A',
			firstChild: child,
			lastChild:  child,
		}
		newChild := &partialIndexTrieNode{
			value: 'C',
		}
		node.appendChild(newChild)

		if node.firstChild != child {
			t.Errorf("Expected to have B as first child, got %+v", node.firstChild)
		}
		if node.lastChild != newChild {
			t.Errorf("Expected to have C as last child, got %+v", node.lastChild)
		}
		if child.nextSibling != newChild {
			t.Errorf("Expected to have C as sibling of B, got %+v", child.nextSibling)
		}
	})
}

func PartialIndexTreeNodeFindChildByValue(t *testing.T) {
	t.Run("no childs", func(t *testing.T) {
		node := &partialIndexTrieNode{
			value:      'A',
			firstChild: nil,
			lastChild:  nil,
		}

		got := node.findChildByValue('B')
		if got != nil {
			t.Errorf("Expected to get nil, got %+v", got)
		}
	})
	t.Run("found", func(t *testing.T) {
		secondChild := &partialIndexTrieNode{
			value: 'C',
		}
		child := &partialIndexTrieNode{
			value:       'B',
			nextSibling: secondChild,
		}
		node := &partialIndexTrieNode{
			value:      'A',
			firstChild: child,
			lastChild:  secondChild,
		}

		got := node.findChildByValue('C')
		if got != secondChild {
			t.Errorf("Expected to get node C, got %+v", got)
		}
	})
	t.Run("not found", func(t *testing.T) {
		child := &partialIndexTrieNode{
			value: 'B',
		}
		node := &partialIndexTrieNode{
			value:      'A',
			firstChild: child,
			lastChild:  child,
		}

		got := node.findChildByValue('C')
		if got != nil {
			t.Errorf("Expected to get nil, got %+v", got)
		}
	})
}

func PartialIndexTreeNodeAppendPosition(t *testing.T) {
	t.Run("no positions", func(t *testing.T) {
		node := &partialIndexTrieNode{}
		node.appendPositionIfNotExists(1)

		if node.firstPosition.Position != 1 {
			t.Errorf("Expected to have 1 as first position, got %+v", node.firstPosition)
		}
		if node.lastPosition.Position != 1 {
			t.Errorf("Expected to have 1 as last position, got %+v", node.lastPosition)
		}
	})
	t.Run("already have another position", func(t *testing.T) {
		position := &record.PositionLinkedList{
			Position: 1,
		}
		node := &partialIndexTrieNode{
			firstPosition: position,
			lastPosition:  position,
		}
		node.appendPositionIfNotExists(2)

		if node.firstPosition.Position != 1 {
			t.Errorf("Expected to have B as first position, got %+v", node.firstPosition)
		}
		if node.lastPosition.Position != 2 {
			t.Errorf("Expected to have 2 as last position, got %+v", node.lastPosition)
		}
		if position.NextPosition.Position != 2 {
			t.Errorf("Expected to have position 2 next to 1, got %+v", position.NextPosition)
		}
	})
	t.Run("already have the same position", func(t *testing.T) {
		position := &record.PositionLinkedList{
			Position: 1,
		}
		node := &partialIndexTrieNode{
			firstPosition: position,
			lastPosition:  position,
		}
		node.appendPositionIfNotExists(1)

		if node.lastPosition != position {
			t.Errorf("Expected to have 1 as first position, got %+v", node.lastPosition)
		}
		if position.NextPosition != nil {
			t.Errorf("Expected not to have a position next to 1, got %+v", position.NextPosition)
		}
	})
}
