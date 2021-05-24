package partial

import (
	"rodb.io/pkg/record"
	"strconv"
	"strings"
	"testing"
)

func PartialIndexTreeNodeAppendChild(t *testing.T) {
	t.Run("no childs", func(t *testing.T) {
		node := &TreeNode{
			value: []byte("A"),
		}
		newChild := &TreeNode{
			value: []byte("B"),
		}
		node.AppendChild(newChild)

		if node.firstChild != newChild {
			t.Errorf("Expected to have B as first child, got %+v", node.firstChild)
		}
		if node.lastChild != newChild {
			t.Errorf("Expected to have B as last child, got %+v", node.lastChild)
		}
	})
	t.Run("already have child", func(t *testing.T) {
		child := &TreeNode{
			value: []byte("B"),
		}
		node := &TreeNode{
			value:      []byte("A"),
			firstChild: child,
			lastChild:  child,
		}
		newChild := &TreeNode{
			value: []byte("C"),
		}
		node.AppendChild(newChild)

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

func PartialIndexTreeNodeFindChildByPrefix(t *testing.T) {
	t.Run("no childs", func(t *testing.T) {
		node := &TreeNode{
			value:      []byte("A"),
			firstChild: nil,
			lastChild:  nil,
		}

		got, _ := node.FindChildByPrefix([]byte("B"))
		if got != nil {
			t.Errorf("Expected to get nil, got %+v", got)
		}
	})
	t.Run("found exact", func(t *testing.T) {
		secondChild := &TreeNode{
			value: []byte("CDE"),
		}
		child := &TreeNode{
			value:       []byte("B"),
			nextSibling: secondChild,
		}
		node := &TreeNode{
			value:      []byte("A"),
			firstChild: child,
			lastChild:  secondChild,
		}

		gotNode, gotLength := node.FindChildByPrefix([]byte("CDE"))
		if gotNode != secondChild {
			t.Errorf("Expected to get node C, got %+v", gotNode)
		}
		if expectLength := len(secondChild.value); gotLength != expectLength {
			t.Errorf("Expected to get length %v, got %v", expectLength, gotLength)
		}
	})
	t.Run("found partial", func(t *testing.T) {
		secondChild := &TreeNode{
			value: []byte("CDE"),
		}
		child := &TreeNode{
			value:       []byte("B"),
			nextSibling: secondChild,
		}
		node := &TreeNode{
			value:      []byte("A"),
			firstChild: child,
			lastChild:  secondChild,
		}

		gotNode, gotLength := node.FindChildByPrefix([]byte("CDX"))
		if gotNode != secondChild {
			t.Errorf("Expected to get node C, got %+v", gotNode)
		}
		if expectLength := 2; gotLength != expectLength {
			t.Errorf("Expected to get length %v, got %v", expectLength, gotLength)
		}
	})
	t.Run("not found", func(t *testing.T) {
		child := &TreeNode{
			value: []byte("B"),
		}
		node := &TreeNode{
			value:      []byte("A"),
			firstChild: child,
			lastChild:  child,
		}

		got, _ := node.FindChildByPrefix([]byte("C"))
		if got != nil {
			t.Errorf("Expected to get nil, got %+v", got)
		}
	})
}

func PartialIndexTreeNodeAppendPosition(t *testing.T) {
	t.Run("no positions", func(t *testing.T) {
		node := &TreeNode{}
		node.AppendPositionIfNotExists(1)

		if node.firstPosition.Position != 1 {
			t.Errorf("Expected to have 1 as first position, got %+v", node.firstPosition)
		}
		if node.lastPosition.Position != 1 {
			t.Errorf("Expected to have 1 as last position, got %+v", node.lastPosition)
		}
	})
	t.Run("already have another position", func(t *testing.T) {
		position := &PositionLinkedList{
			Position: 1,
		}
		node := &TreeNode{
			firstPosition: position,
			lastPosition:  position,
		}
		node.AppendPositionIfNotExists(2)

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
		position := &PositionLinkedList{
			Position: 1,
		}
		node := &TreeNode{
			firstPosition: position,
			lastPosition:  position,
		}
		node.AppendPositionIfNotExists(1)

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
		root := &TreeNode{}

		root.AddSequence([]byte("FOO"), 1)

		if string(root.firstChild.value) != "FOO" {
			t.Errorf("Expected to have FOO as first node of root, got %v", root.firstChild.value)
		}
		if root.firstChild.firstPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in first position of root, got %v", root.firstChild.firstPosition)
		}
		if root.firstChild.lastPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in last position of root, got %v", root.firstChild.lastPosition)
		}
	})
	t.Run("new suffix to existing node", func(t *testing.T) {
		root := &TreeNode{}
		root.AddSequence([]byte("FOO"), 1)

		root.AddSequence([]byte("FOOT"), 2)

		if root.firstChild.firstPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in first position of root, got %v", root.firstChild.firstPosition)
		}
		if root.firstChild.lastPosition.Position != 2 {
			t.Errorf("Expected to have position 1 in last position of root, got %v", root.firstChild.lastPosition)
		}

		if string(root.firstChild.firstChild.value) != "T" {
			t.Errorf("Expected to have value 'T' as first child of FOO, got %v", string(root.firstChild.firstChild.value))
		}
		if root.firstChild.firstChild.firstPosition.Position != 2 {
			t.Errorf("Expected to have position 1 in first position of T, got %v", root.firstChild.firstChild.firstPosition)
		}
		if root.firstChild.firstChild.lastPosition.Position != 2 {
			t.Errorf("Expected to have position 1 in last position of T, got %v", root.firstChild.firstChild.lastPosition)
		}
	})
	t.Run("splitting existing node", func(t *testing.T) {
		root := &TreeNode{}
		root.AddSequence([]byte("FOO"), 1)
		root.AddSequence([]byte("FORMAT"), 2)

		if string(root.firstChild.value) != "FO" {
			t.Errorf("Expected to have value 'FO' as first child of root, got %v", string(root.firstChild.value))
		}
		if root.firstChild.firstPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in first position of root, got %v", root.firstChild.firstPosition)
		}
		if root.firstChild.lastPosition.Position != 2 {
			t.Errorf("Expected to have position 1 in last position of root, got %v", root.firstChild.lastPosition)
		}

		if string(root.firstChild.firstChild.value) != "O" {
			t.Errorf("Expected to have value 'O' as first child of 'FO', got %v", string(root.firstChild.firstChild.value))
		}
		if root.firstChild.firstChild.firstPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in first position of T, got %v", root.firstChild.firstChild.firstPosition)
		}
		if root.firstChild.firstChild.lastPosition.Position != 1 {
			t.Errorf("Expected to have position 1 in last position of T, got %v", root.firstChild.firstChild.lastPosition)
		}

		if string(root.firstChild.lastChild.value) != "RMAT" {
			t.Errorf("Expected to have value 'RMAT' as first child of 'FO', got %v", string(root.firstChild.lastChild.value))
		}
		if root.firstChild.lastChild.firstPosition.Position != 2 {
			t.Errorf("Expected to have position 2 in first position of 'RMAT', got %v", root.firstChild.lastChild.firstPosition)
		}
		if root.firstChild.lastChild.lastPosition.Position != 2 {
			t.Errorf("Expected to have position 2 in last position of 'RMAT', got %v", root.firstChild.lastChild.lastPosition)
		}
	})
}

func PartialIndexTreeNodeGetSequence(t *testing.T) {
	root := &TreeNode{}
	root.AddSequence([]byte("FOO"), 1)
	root.AddSequence([]byte("FOOT"), 2)
	root.AddSequence([]byte("TOP"), 3)

	checkList := func(t *testing.T, list *PositionLinkedList, positions []record.Position) {
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
		checkList(t, root.GetSequence([]byte("FOO")), []record.Position{1, 2})
	})
	t.Run("not found", func(t *testing.T) {
		checkList(t, root.GetSequence([]byte("BAR")), []record.Position{})
	})
	t.Run("single", func(t *testing.T) {
		checkList(t, root.GetSequence([]byte("TOP")), []record.Position{3})
	})
	t.Run("multiple middle", func(t *testing.T) {
		checkList(t, root.GetSequence([]byte("OO")), []record.Position{1, 2})
	})
	t.Run("suffix and prefix", func(t *testing.T) {
		checkList(t, root.GetSequence([]byte("T")), []record.Position{2, 3})
	})
	t.Run("partially matches the end", func(t *testing.T) {
		checkList(t, root.GetSequence([]byte("FO")), []record.Position{1, 2})
	})
}
