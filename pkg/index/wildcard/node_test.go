package wildcard

import (
	"fmt"
	"github.com/rodb-io/rodb/pkg/input/record"
	"strconv"
	"strings"
	"testing"
)

func createTestNode(t *testing.T, stream *Stream, value []byte) *TreeNode {
	node, err := NewEmptyTreeNode(stream)
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	valueOffset, err := stream.Add(value)
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	node.valueOffset = TreeNodeValueOffset(valueOffset)
	node.valueLength = int64(len(value))

	if err := node.Save(); err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	return node
}

func TestGetTreeNode(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		stream := createTestStream(t)
		offset, err := stream.Add([]byte{
			0, 0, 0, 0, 0, 0, 0, 0x01,
			0, 0, 0, 0, 0, 0, 0, 0x02,
			0, 0, 0, 0, 0, 0, 0, 0x03,
			0, 0, 0, 0, 0, 0, 0, 0x04,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0xFF,
			0, 0, 0, 0, 0, 0, 0x4, 0xD2,
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		node, err := GetTreeNode(stream, TreeNodeOffset(offset))
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := TreeNodeValueOffset(1), node.valueOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(2), node.valueLength; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(3), node.nextSiblingOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(4), node.firstChildOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(0), node.lastChildOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := PositionLinkedListOffset(255), node.firstPositionOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := PositionLinkedListOffset(1234), node.lastPositionOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestNewEmptyTreeNode(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		stream := createTestStream(t)
		initialSize := stream.streamSize

		node, err := NewEmptyTreeNode(stream)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := initialSize, int64(node.offset); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		expectBytes := []byte{
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
		}
		gotBytes, err := stream.Get(initialSize, 56)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := fmt.Sprintf("%x", expectBytes), fmt.Sprintf("%x", gotBytes); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestTreeNodeSerialize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		node := TreeNode{
			valueOffset:         1,
			valueLength:         2,
			nextSiblingOffset:   3,
			firstChildOffset:    4,
			lastChildOffset:     0,
			firstPositionOffset: 255,
			lastPositionOffset:  1234,
		}

		expect := []byte{
			0, 0, 0, 0, 0, 0, 0, 0x01,
			0, 0, 0, 0, 0, 0, 0, 0x02,
			0, 0, 0, 0, 0, 0, 0, 0x03,
			0, 0, 0, 0, 0, 0, 0, 0x04,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0xFF,
			0, 0, 0, 0, 0, 0, 0x4, 0xD2,
		}
		if expect, got := fmt.Sprintf("%x", expect), fmt.Sprintf("%x", node.Serialize()); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestTreeNodeUnserialize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		node := TreeNode{}
		node.Unserialize([]byte{
			0, 0, 0, 0, 0, 0, 0, 0x01,
			0, 0, 0, 0, 0, 0, 0, 0x02,
			0, 0, 0, 0, 0, 0, 0, 0x03,
			0, 0, 0, 0, 0, 0, 0, 0x04,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0xFF,
			0, 0, 0, 0, 0, 0, 0x4, 0xD2,
		})
		if expect, got := TreeNodeValueOffset(1), node.valueOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(2), node.valueLength; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(3), node.nextSiblingOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(4), node.firstChildOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(0), node.lastChildOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := PositionLinkedListOffset(255), node.firstPositionOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := PositionLinkedListOffset(1234), node.lastPositionOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
	t.Run("from serialize", func(t *testing.T) {
		list1 := TreeNode{
			valueOffset:         1,
			valueLength:         2,
			nextSiblingOffset:   3,
			firstChildOffset:    4,
			lastChildOffset:     0,
			firstPositionOffset: 255,
			lastPositionOffset:  1234,
		}

		list2 := TreeNode{}
		list2.Unserialize(list1.Serialize())

		if expect, got := TreeNodeValueOffset(1), list2.valueOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(2), list2.valueLength; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(3), list2.nextSiblingOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(4), list2.firstChildOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(0), list2.lastChildOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := PositionLinkedListOffset(255), list2.firstPositionOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := PositionLinkedListOffset(1234), list2.lastPositionOffset; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestTreeNodeSave(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		stream := createTestStream(t)
		offset, err := stream.Add([]byte{
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		node := TreeNode{
			stream:              stream,
			offset:              TreeNodeOffset(offset),
			valueOffset:         1,
			valueLength:         2,
			nextSiblingOffset:   3,
			firstChildOffset:    4,
			lastChildOffset:     0,
			firstPositionOffset: 255,
			lastPositionOffset:  1234,
		}
		if err := node.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := offset, int64(node.offset); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		expectBytes := []byte{
			0, 0, 0, 0, 0, 0, 0, 0x01,
			0, 0, 0, 0, 0, 0, 0, 0x02,
			0, 0, 0, 0, 0, 0, 0, 0x03,
			0, 0, 0, 0, 0, 0, 0, 0x04,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0xFF,
			0, 0, 0, 0, 0, 0, 0x4, 0xD2,
		}
		gotBytes, err := stream.Get(offset, 56)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := fmt.Sprintf("%x", expectBytes), fmt.Sprintf("%x", gotBytes); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestTreeNodeAppendChild(t *testing.T) {
	t.Run("no childs", func(t *testing.T) {
		stream := createTestStream(t)
		node := createTestNode(t, stream, []byte("A"))
		newChild := createTestNode(t, stream, []byte("B"))

		if err := node.AppendChild(newChild); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if node.firstChildOffset != newChild.offset {
			t.Fatalf("Expected to have B as first child offset, got %+v", node.firstChildOffset)
		}
		if node.lastChildOffset != newChild.offset {
			t.Fatalf("Expected to have B as last child offset, got %+v", node.lastChildOffset)
		}
	})
	t.Run("already have child", func(t *testing.T) {
		stream := createTestStream(t)
		child := createTestNode(t, stream, []byte("B"))
		if err := child.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		newChild := createTestNode(t, stream, []byte("C"))
		if err := child.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		node := createTestNode(t, stream, []byte("A"))
		node.firstChildOffset = child.offset
		node.lastChildOffset = child.offset
		if err := node.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if err := node.AppendChild(newChild); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if node.firstChildOffset != child.offset {
			t.Fatalf("Expected to have B as first child offset, got %+v", node.firstChildOffset)
		}
		if node.lastChildOffset != newChild.offset {
			t.Fatalf("Expected to have C as last child offset, got %+v", node.lastChildOffset)
		}

		child, err := node.FirstChild()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if child.nextSiblingOffset != newChild.offset {
			t.Fatalf("Expected to have C's offset as sibling of B, got %+v", child.nextSiblingOffset)
		}
	})
}

func TestTreeNodeFindChildByPrefix(t *testing.T) {
	t.Run("no childs", func(t *testing.T) {
		stream := createTestStream(t)
		node := createTestNode(t, stream, []byte("A"))
		node.firstChildOffset = 0
		node.lastChildOffset = 0

		got, _, err := node.FindChildByPrefix([]byte("B"))
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if got != nil {
			t.Fatalf("Expected to get nil, got %+v", got)
		}
	})
	t.Run("found exact", func(t *testing.T) {
		stream := createTestStream(t)
		secondChild := createTestNode(t, stream, []byte("CDE"))
		child := createTestNode(t, stream, []byte("B"))
		child.nextSiblingOffset = secondChild.offset
		if err := child.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		node := createTestNode(t, stream, []byte("A"))
		node.firstChildOffset = child.offset
		node.lastChildOffset = secondChild.offset
		if err := node.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		gotNode, gotLength, err := node.FindChildByPrefix([]byte("CDE"))
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if gotNode.offset != secondChild.offset {
			t.Fatalf("Expected to get node C, got %+v", gotNode)
		}

		if expectLength := int64(3); gotLength != expectLength {
			t.Fatalf("Expected to get length %v, got %v", expectLength, gotLength)
		}
	})
	t.Run("found wildcard", func(t *testing.T) {
		stream := createTestStream(t)
		secondChild := createTestNode(t, stream, []byte("CDE"))
		child := createTestNode(t, stream, []byte("B"))
		child.nextSiblingOffset = secondChild.offset
		if err := child.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		node := createTestNode(t, stream, []byte("A"))
		node.firstChildOffset = child.offset
		node.lastChildOffset = secondChild.offset
		if err := node.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		gotNode, gotLength, err := node.FindChildByPrefix([]byte("CDX"))
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if gotNode.offset != secondChild.offset {
			t.Fatalf("Expected to get node C, got %+v", gotNode)
		}
		if expectLength := int64(2); gotLength != expectLength {
			t.Fatalf("Expected to get length %v, got %v", expectLength, gotLength)
		}
	})
	t.Run("not found", func(t *testing.T) {
		stream := createTestStream(t)
		child := createTestNode(t, stream, []byte("B"))
		node := createTestNode(t, stream, []byte("A"))
		node.firstChildOffset = child.offset
		node.lastChildOffset = child.offset
		if err := node.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		got, _, err := node.FindChildByPrefix([]byte("C"))
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if got != nil {
			t.Fatalf("Expected to get nil, got %+v", got)
		}
	})
}

func TestTreeNodeAppendPositionIfNotExists(t *testing.T) {
	t.Run("no positions", func(t *testing.T) {
		stream := createTestStream(t)
		node := createTestNode(t, stream, []byte(""))
		if err := node.AppendPositionIfNotExists(1); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		firstPosition, err := node.FirstPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if firstPosition.Position != 1 {
			t.Fatalf("Expected to have 1 as first position, got %+v", firstPosition)
		}

		lastPosition, err := node.LastPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if lastPosition.Position != 1 {
			t.Fatalf("Expected to have 1 as last position, got %+v", lastPosition)
		}
	})
	t.Run("already have another position", func(t *testing.T) {
		stream := createTestStream(t)
		position, err := NewPositionLinkedList(stream, 1)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		node := createTestNode(t, stream, []byte(""))
		node.firstPositionOffset = position.offset
		node.lastPositionOffset = position.offset
		if err := node.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if err := node.AppendPositionIfNotExists(2); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		firstPosition, err := node.FirstPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if firstPosition.Position != 1 {
			t.Fatalf("Expected to have B as first position, got %+v", firstPosition)
		}

		lastPosition, err := node.LastPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if lastPosition.Position != 2 {
			t.Fatalf("Expected to have 2 as last position, got %+v", lastPosition)
		}

		nextPosition, err := firstPosition.NextPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if nextPosition.Position != 2 {
			t.Fatalf("Expected to have position 2 next to 1, got %+v", nextPosition)
		}
	})
	t.Run("already have the same position", func(t *testing.T) {
		stream := createTestStream(t)
		position, err := NewPositionLinkedList(stream, 1)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		node := createTestNode(t, stream, []byte(""))
		node.firstPositionOffset = position.offset
		node.lastPositionOffset = position.offset
		if err := node.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if err := node.AppendPositionIfNotExists(1); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		lastPosition, err := node.LastPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if lastPosition.Position != position.Position {
			t.Fatalf("Expected to have 1 as first position, got %+v", lastPosition)
		}

		nextPosition, err := position.NextPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if nextPosition != nil {
			t.Fatalf("Expected not to have a position next to 1, got %+v", nextPosition)
		}
	})
}

func TestTreeNodeAddSingleSequence(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		stream := createTestStream(t)
		root := createTestNode(t, stream, []byte(""))

		if err := root.AddSingleSequence([]byte("FOO"), 1); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		firstChild, err := root.FirstChild()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		value, err := firstChild.Value()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if string(value) != "FOO" {
			t.Fatalf("Expected to have FOO as first node of root, got %v", value)
		}

		firstPosition, err := firstChild.FirstPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if firstPosition.Position != 1 {
			t.Fatalf("Expected to have position 1 in first position of root, got %v", firstPosition)
		}

		lastPosition, err := firstChild.LastPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if lastPosition.Position != 1 {
			t.Fatalf("Expected to have position 1 in last position of root, got %v", lastPosition)
		}
	})
	t.Run("new suffix to existing node", func(t *testing.T) {
		stream := createTestStream(t)
		root := createTestNode(t, stream, []byte(""))
		if err := root.AddSingleSequence([]byte("FOO"), 1); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if err := root.AddSingleSequence([]byte("FOOT"), 2); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		firstChild, err := root.FirstChild()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		firstPosition, err := firstChild.FirstPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if firstPosition.Position != 1 {
			t.Fatalf("Expected to have position 1 in first position of root, got %v", firstPosition)
		}

		lastPosition, err := firstChild.LastPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if lastPosition.Position != 2 {
			t.Fatalf("Expected to have position 1 in last position of root, got %v", lastPosition)
		}

		firstChildOfFirstChild, err := firstChild.FirstChild()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		value, err := firstChildOfFirstChild.Value()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if string(value) != "T" {
			t.Fatalf("Expected to have value 'T' as first child of FOO, got %v", string(value))
		}

		firstPosition, err = firstChildOfFirstChild.FirstPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if firstPosition.Position != 2 {
			t.Fatalf("Expected to have position 1 in first position of T, got %v", firstPosition)
		}

		lastPosition, err = firstChildOfFirstChild.LastPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if lastPosition.Position != 2 {
			t.Fatalf("Expected to have position 1 in last position of T, got %v", lastPosition)
		}
	})
	t.Run("splitting existing node", func(t *testing.T) {
		stream := createTestStream(t)
		root := createTestNode(t, stream, []byte(""))
		if err := root.AddSingleSequence([]byte("FOO"), 1); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if err := root.AddSingleSequence([]byte("FORMAT"), 2); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		firstChild, err := root.FirstChild()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		value, err := firstChild.Value()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if string(value) != "FO" {
			t.Fatalf("Expected to have value 'FO' as first child of root, got %v", string(value))
		}

		firstPosition, err := firstChild.FirstPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if firstPosition.Position != 1 {
			t.Fatalf("Expected to have position 1 in first position of root, got %v", firstPosition)
		}

		lastPosition, err := firstChild.LastPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if lastPosition.Position != 2 {
			t.Fatalf("Expected to have position 1 in last position of root, got %v", lastPosition)
		}

		firstChildOfFirstChild, err := firstChild.FirstChild()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		value, err = firstChildOfFirstChild.Value()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if string(value) != "O" {
			t.Fatalf("Expected to have value 'O' as first child of 'FO', got %v", string(value))
		}

		firstPosition, err = firstChildOfFirstChild.FirstPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if firstPosition.Position != 1 {
			t.Fatalf("Expected to have position 1 in first position of T, got %v", firstPosition)
		}

		lastPosition, err = firstChildOfFirstChild.LastPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if lastPosition.Position != 1 {
			t.Fatalf("Expected to have position 1 in last position of T, got %v", lastPosition)
		}

		lastChildOfFirstChild, err := firstChild.LastChild()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		value, err = lastChildOfFirstChild.Value()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if string(value) != "RMAT" {
			t.Fatalf("Expected to have value 'RMAT' as first child of 'FO', got %v", string(value))
		}

		firstPosition, err = lastChildOfFirstChild.FirstPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if firstPosition.Position != 2 {
			t.Fatalf("Expected to have position 2 in first position of 'RMAT', got %v", firstPosition)
		}

		lastPosition, err = lastChildOfFirstChild.LastPosition()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if lastPosition.Position != 2 {
			t.Fatalf("Expected to have position 2 in last position of 'RMAT', got %v", lastPosition)
		}
	})
}

func TestTreeNodeGetSequence(t *testing.T) {
	stream := createTestStream(t)
	root, err := NewEmptyTreeNode(stream)
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}
	if err := root.Save(); err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	addSequences := func(t *testing.T, bytes []byte, position record.Position) {
		for i := 0; i < len(bytes); i++ {
			if err := root.AddSingleSequence(bytes[i:], position); err != nil {
				t.Fatalf("Unexpected error: '%+v'", err)
			}
		}
	}
	addSequences(t, []byte("FOO"), 1)
	addSequences(t, []byte("FOOT"), 2)
	addSequences(t, []byte("TOP"), 3)

	checkList := func(t *testing.T, node *PositionLinkedList, positions []record.Position) {
		stringPositions := []string{}
		for _, position := range positions {
			stringPositions = append(stringPositions, strconv.Itoa(int(position)))
		}
		expect := strings.Join(stringPositions, ",")

		got := ""
		iterator := node.Iterate()
		for {
			position, err := iterator()
			if err != nil {
				t.Fatalf("Unexpected error: '%+v'", err)
			}
			if position == nil {
				break
			}

			if got != "" {
				got += ","
			}
			got += strconv.Itoa(int(*position))
		}

		if got != expect {
			t.Fatalf("Expected to get positions %v, got %v", expect, got)
		}
	}

	getSequence := func(node *TreeNode, sequence []byte) *PositionLinkedList {
		list, err := node.GetSequence(sequence)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		return list
	}

	t.Run("multiple prefix", func(t *testing.T) {
		checkList(t, getSequence(root, []byte("FOO")), []record.Position{1, 2})
	})
	t.Run("not found", func(t *testing.T) {
		checkList(t, getSequence(root, []byte("BAR")), []record.Position{})
	})
	t.Run("single", func(t *testing.T) {
		checkList(t, getSequence(root, []byte("TOP")), []record.Position{3})
	})
	t.Run("multiple middle", func(t *testing.T) {
		checkList(t, getSequence(root, []byte("OO")), []record.Position{1, 2})
	})
	t.Run("suffix and prefix", func(t *testing.T) {
		checkList(t, getSequence(root, []byte("T")), []record.Position{2, 3})
	})
	t.Run("only matches the end", func(t *testing.T) {
		checkList(t, getSequence(root, []byte("FO")), []record.Position{1, 2})
	})
}
