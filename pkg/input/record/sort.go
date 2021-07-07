package record

import (
	"rodb.io/pkg/parser"
	"sort"
)

func (records List) Sort(config []*SortConfig) List {
	sorter := &recordListSorter{records, config}
	sort.Sort(sorter)
	return sorter.records
}

// Implementing sort.Interface
type recordListSorter struct {
	records List
	config  []*SortConfig
}

func (sorter *recordListSorter) Len() int {
	return len(sorter.records)
}
func (sorter *recordListSorter) Less(i int, j int) bool {
	iRecord, jRecord := sorter.records[i], sorter.records[j]

	for _, sort := range sorter.config {
		iValue, err := iRecord.Get(sort.Property)
		if err != nil {
			sort.Logger.Errorf("Unhandlable error during sort operation on property '%v': %v", sort.Property, err)
		}

		jValue, err := jRecord.Get(sort.Property)
		if err != nil {
			sort.Logger.Errorf("Unhandlable error during sort operation on property '%v': %v", sort.Property, err)
		}

		result, err := parser.Compare(iValue, jValue)
		if err != nil {
			sort.Logger.Errorf("Unhandlable error during sort operation on property '%v': %v", sort.Property, err)
		}

		if result == nil {
			// Equal for this criteria: go to the next one
			continue
		}

		if sort.IsAscending() {
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
