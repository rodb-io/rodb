package record

import (
	"rods/pkg/config"
	"rods/pkg/parser"
	"sort"
)

type Position = int64

type PositionList []Position

type List []Record

type Record interface {
	All() (map[string]interface{}, error)
	Get(field string) (interface{}, error)
	Position() Position
}

func (records List) Sort(config []*config.Sort) List {
	sorter := &recordListSorter{records, config}
	sort.Sort(sorter)
	return sorter.records
}

// Implementing sort.Interface
type recordListSorter struct {
	records List
	config  []*config.Sort
}

func (sorter *recordListSorter) Len() int {
	return len(sorter.records)
}
func (sorter *recordListSorter) Less(i int, j int) bool {
	iRecord, jRecord := sorter.records[i], sorter.records[j]

	for _, sort := range sorter.config {
		iValue, err := iRecord.Get(sort.Column)
		if err != nil {
			sort.Logger.Errorf("Unhandlable error during sort operation: %v", err)
		}

		jValue, err := jRecord.Get(sort.Column)
		if err != nil {
			sort.Logger.Errorf("Unhandlable error during sort operation: %v", err)
		}

		result, err := parser.Compare(iValue, jValue)
		if err != nil {
			sort.Logger.Errorf("Unhandlable error during sort operation: %v", err)
		}

		if result == nil {
			// Equal for this criteria: go to the next one
			continue
		}

		if *sort.Ascending == true {
			return *result
		} else {
			return !*result
		}
	}

	// If we end-up here, both are equal
	return false
}
func (sorter *recordListSorter) Swap(i int, j int) {
	iRecord, jRecord := sorter.records[i], sorter.records[j]
	sorter.records[i], sorter.records[j] = jRecord, iRecord
}
