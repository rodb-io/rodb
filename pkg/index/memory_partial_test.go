package index

import (
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/config"
	"rodb.io/pkg/input"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/record"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func stringifyMemoryPartialTrie(root *partialIndexTrieNode) string {
	positionsToString := func(positions *record.PositionLinkedList) string {
		result := ""
		for currentPosition := positions; currentPosition != nil; currentPosition = currentPosition.NextPosition {
			if currentPosition != positions {
				result += ","
			}
			result += strconv.Itoa(int(currentPosition.Position))
		}

		return result
	}

	results := make([]string, 0)

	var stringify func(prefix string, node *partialIndexTrieNode)
	stringify = func(prefix string, node *partialIndexTrieNode) {
		for child := node.firstChild; child != nil; child = child.nextSibling {
			positions := positionsToString(child.firstPosition)
			currentValue := prefix + string(child.value)
			results = append(results, currentValue+"="+positions)
			stringify(currentValue, child)
		}
	}

	stringify("", root)
	sort.Strings(results)

	return strings.Join(results, "\n")
}

func TestMemoryPartial(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		index, err := NewMemoryPartial(
			&config.MemoryPartialIndex{
				Properties: []string{"col"},
				Input:      "input",
				Logger:     logrus.NewEntry(logrus.StandardLogger()),
			},
			input.List{
				"input": input.NewMock(parser.NewMock(), []record.Record{
					record.NewStringPropertiesMock(map[string]string{
						"col": "BANANA",
					}, 1),
					record.NewStringPropertiesMock(map[string]string{
						"col": "BANANO",
					}, 2),
					record.NewStringPropertiesMock(map[string]string{
						"col": "PLANT",
					}, 3),
				}),
			},
		)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expect := strings.Join([]string{
			"A=1,2,3",
			"AN=1,2,3",
			"ANA=1,2",
			"ANAN=1,2",
			"ANANA=1",
			"ANANO=2",
			"ANO=2",
			"ANT=3",
			"B=1,2",
			"BA=1,2",
			"BAN=1,2",
			"BANA=1,2",
			"BANAN=1,2",
			"BANANA=1",
			"BANANO=2",
			"L=3",
			"LA=3",
			"LAN=3",
			"LANT=3",
			"N=1,2,3",
			"NA=1,2",
			"NAN=1,2",
			"NANA=1",
			"NANO=2",
			"NO=2",
			"NT=3",
			"P=3",
			"PL=3",
			"PLA=3",
			"PLAN=3",
			"PLANT=3",
		}, "\n")
		got := stringifyMemoryPartialTrie(index.index["col"])
		if got != expect {
			t.Errorf("Unexpected list of results. Expected :\n=====\n%v\n=====\nbut got:\n=====\n%v\n", expect, got)
		}
	})
}
