package index

import (
	"rodb.io/pkg/record"
	"strconv"
	"strings"
	"testing"
)

func PartialIndexTreeNodeAppendChild(t *testing.T) {
	t.Run("no childs", func(t *testing.T) {
		node := &partialIndexTreeNode{
			value: 'A',
		}
		newChild := &partialIndexTreeNode{
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
		child := &partialIndexTreeNode{
			value: 'B',
		}
		node := &partialIndexTreeNode{
			value:      'A',
			firstChild: child,
			lastChild:  child,
		}
		newChild := &partialIndexTreeNode{
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
		node := &partialIndexTreeNode{
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
		secondChild := &partialIndexTreeNode{
			value: 'C',
		}
		child := &partialIndexTreeNode{
			value:       'B',
			nextSibling: secondChild,
		}
		node := &partialIndexTreeNode{
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
		child := &partialIndexTreeNode{
			value: 'B',
		}
		node := &partialIndexTreeNode{
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
		node := &partialIndexTreeNode{}
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
		node := &partialIndexTreeNode{
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
		node := &partialIndexTreeNode{
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

func PartialIndexTreeNodeAddSequence(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		root := &partialIndexTreeNode{}

		root.addSequence([]byte("FOO"), 1)

		if root.firstChild.value != 'F' {
			t.Errorf("Expected to have F as first node of root, got %v", root.firstChild.value)
		}
		if root.firstChild.firstPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in first position of root, got %v", root.firstChild.firstPosition)
		}
		if root.firstChild.lastPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in last position of root, got %v", root.firstChild.lastPosition)
		}

		if root.firstChild.firstChild.value != 'O' {
			t.Errorf("Expected to have O as first node F, got %v", root.firstChild.firstChild.value)
		}
		if root.firstChild.firstChild.firstPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in first position of O, got %v", root.firstChild.firstChild.firstPosition)
		}
		if root.firstChild.firstChild.lastPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in last position of O, got %v", root.firstChild.firstChild.lastPosition)
		}

		if root.firstChild.firstChild.firstChild.value != 'F' {
			t.Errorf("Expected to have O as first node of O, got %v", root.firstChild.firstChild.firstChild.value)
		}
		if root.firstChild.firstChild.firstChild.firstPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in first position of O, got %v", root.firstChild.firstChild.firstChild.firstPosition)
		}
		if root.firstChild.firstChild.firstChild.lastPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in last position of O, got %v", root.firstChild.firstChild.firstChild.lastPosition)
		}
	})
	t.Run("new suffix to existing tree", func(t *testing.T) {
		root := &partialIndexTreeNode{}
		root.addSequence([]byte("FOO"), 1)

		root.addSequence([]byte("FOOT"), 2)

		if root.firstChild.firstPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in first position of root, got %v", root.firstChild.firstPosition)
		}
		if root.firstChild.lastPosition.Position != 2 {
			t.Errorf("Expected to have position 1 in last position of root, got %v", root.firstChild.lastPosition)
		}

		if root.firstChild.firstChild.firstPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in first position of O, got %v", root.firstChild.firstChild.firstPosition)
		}
		if root.firstChild.firstChild.lastPosition.Position != 2 {
			t.Errorf("Expected to have position 1 in last position of O, got %v", root.firstChild.firstChild.lastPosition)
		}

		if root.firstChild.firstChild.firstChild.firstPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in first position of O, got %v", root.firstChild.firstChild.firstChild.firstPosition)
		}
		if root.firstChild.firstChild.firstChild.lastPosition.Position != 2 {
			t.Errorf("Expected to have position 1 in last position of O, got %v", root.firstChild.firstChild.firstChild.lastPosition)
		}

		if root.firstChild.firstChild.firstChild.firstChild.value != 'T' {
			t.Errorf("Expected to have T as first node of O, got %v", root.firstChild.firstChild.firstChild.firstChild.value)
		}
		if root.firstChild.firstChild.firstChild.firstChild.firstPosition.Position != 2 {
			t.Errorf("Expected to have position 2 in first position of T, got %v", root.firstChild.firstChild.firstChild.firstChild.firstPosition)
		}
		if root.firstChild.firstChild.firstChild.firstChild.lastPosition.Position != 2 {
			t.Errorf("Expected to have position 2 in last position of T, got %v", root.firstChild.firstChild.firstChild.firstChild.lastPosition)
		}
	})
}

func PartialIndexTreeNodeGetSequence(t *testing.T) {
	root := &partialIndexTreeNode{}
	root.addSequence([]byte("FOO"), 1)
	root.addSequence([]byte("FOOT"), 2)
	root.addSequence([]byte("TOP"), 3)

	checkList := func(t *testing.T, list *record.PositionLinkedList, positions []record.Position) {
		stringPositions := []string{}
		for _, position := range positions {
			stringPositions = append(stringPositions, strconv.Itoa(int(position)))
		}
		expect := strings.Join(stringPositions, ",")

		got := ""
		iterator := list.Iterate()
		for {
			position, _ := iterator()
			if position == nil {
				break
			}

			if got != "" {
				got += ","
			}
			got += strconv.Itoa(int(*position))
		}

		if got != expect {
			t.Errorf("Expected to get positions %v, got %v", expect, got)
		}
	}

	t.Run("multiple prefix", func(t *testing.T) {
		checkList(t, root.getSequence([]byte("FOO")), []record.Position{1, 2})
	})
	t.Run("not found", func(t *testing.T) {
		checkList(t, root.getSequence([]byte("BAR")), []record.Position{})
	})
	t.Run("single", func(t *testing.T) {
		checkList(t, root.getSequence([]byte("TOP")), []record.Position{3})
	})
	t.Run("multiple middle", func(t *testing.T) {
		checkList(t, root.getSequence([]byte("OO")), []record.Position{1, 2})
	})
	t.Run("suffix and prefix", func(t *testing.T) {
		checkList(t, root.getSequence([]byte("T")), []record.Position{2, 3})
	})
}
