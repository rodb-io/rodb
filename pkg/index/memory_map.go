package index

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"rods/pkg/config"
	"rods/pkg/input"
	"rods/pkg/record"
	"rods/pkg/source"
	"time"
)

type memoryMapColumnValueIndex = record.PositionList
type memoryMapColumnIndex = map[interface{}]memoryMapColumnValueIndex
type memoryMapIndex = map[string]memoryMapColumnIndex

type MemoryMap struct {
	config       *config.MemoryMapIndex
	input        input.Input
	index        memoryMapIndex
	inputWatcher *source.Watcher
}

func NewMemoryMap(
	config *config.MemoryMapIndex,
	inputs input.List,
) (*MemoryMap, error) {
	input, inputExists := inputs[config.Input]
	if !inputExists {
		return nil, fmt.Errorf("Input '%v' not found in inputs list.", config.Input)
	}

	memoryMap := &MemoryMap{
		config: config,
		input:  input,
	}

	for _, columnName := range memoryMap.config.Columns {
		if !memoryMap.input.HasColumn(columnName) {
			return nil, errors.New("Input '" + memoryMap.config.Input + "' does not have a column named '" + columnName + "'.")
		}
	}

	memoryMap.inputWatcher = &source.Watcher{
		OnChange: func() {
			message := "The source has been modified by another process"
			if *memoryMap.config.DieOnInputChange {
				memoryMap.config.Logger.Fatalln(message + ". Quitting because it may have corrupted this index and 'dieOnInputChange' is 'true'.")
			} else {
				memoryMap.config.Logger.Warnln(message + ", but 'dieOnInputChange' is 'false'. This could have unpredictable consequences.")
			}
		},
	}
	input.Watch(memoryMap.inputWatcher)

	// Note: need to have the watcher setup so we can fail
	// if a file change happens during the indexing process
	err := memoryMap.Reindex()
	if err != nil {
		return nil, err
	}

	return memoryMap, nil
}

func (mm *MemoryMap) Reindex() error {
	index := make(memoryMapIndex)
	for _, column := range mm.config.Columns {
		index[column] = make(memoryMapColumnIndex)
	}

	totalSize, err := mm.input.Size()
	if err != nil {
		mm.config.Logger.Errorf("Cannot determine the total size of the input: '%+v'. The indexing progress will not be displayed.", err)
	} else if totalSize == 0 {
		mm.config.Logger.Infoln("The total size of the input is unknown. The indexing progress will not be displayed.")
	}

	nextProgress := time.Now()
	for result := range mm.input.IterateAll() {
		if result.Error != nil {
			return result.Error
		}

		if totalSize != 0 {
			if now := time.Now(); now.After(nextProgress) {
				progress := float64(result.Record.Position()) / float64(totalSize)
				mm.config.Logger.Infof("Indexing progress: %d%%", int(math.Floor(progress*100)))
				nextProgress = now.Add(time.Second)
			}
		}

		for _, column := range mm.config.Columns {
			value, err := result.Record.Get(column)
			if err != nil {
				return err
			}

			if value != nil {
				value = reflect.ValueOf(value).Interface()
			}

			columnIndex := index[column]
			valueIndexes, valueIndexesExists := columnIndex[value]
			if valueIndexesExists {
				columnIndex[value] = append(valueIndexes, result.Record.Position())
			} else {
				columnIndex[value] = record.PositionList{result.Record.Position()}
			}
		}
	}

	mm.index = index
	mm.config.Logger.Infof("Successfully finished indexing")

	return nil
}

// Get the record positions (if indexed) that matches all the given filters
// A limit of 0 means that there is no limit
func (mm *MemoryMap) GetRecordPositions(inputName string, filters map[string]interface{}, limit uint) (record.PositionList, error) {
	if !mm.config.DoesHandleInput(inputName) {
		return nil, fmt.Errorf("This index does not handle the input '%v'.", inputName)
	}
	if len(filters) == 0 {
		return nil, fmt.Errorf("This index requires at least one filter.")
	}

	individualFiltersResults := make([]memoryMapColumnValueIndex, 0, len(filters))
	for columnName, filter := range filters {
		if !mm.config.DoesHandleColumn(columnName) {
			return nil, fmt.Errorf("This index does not handle the column '%v'.", columnName)
		}

		indexedValues, foundIndexedValues := mm.index[columnName]
		if !foundIndexedValues {
			return make(record.PositionList, 0), nil
		}

		indexedResults, foundIndexedResults := indexedValues[filter]
		if !foundIndexedResults {
			return make(record.PositionList, 0), nil
		}

		individualFiltersResults = append(individualFiltersResults, indexedResults)
	}

	records := make(record.PositionList, 0)
	for i := 0; i < len(individualFiltersResults[0]); i++ {
		position := individualFiltersResults[0][i]

		matchesAllCriterias := true
		for j := 1; j < len(individualFiltersResults); j++ {
			matchesCurrentCriteria := false
			for _, currentPosition := range individualFiltersResults[j] {
				if currentPosition == position {
					matchesCurrentCriteria = true
					break
				}
			}

			if !matchesCurrentCriteria {
				matchesAllCriterias = false
				break
			}
		}

		if matchesAllCriterias {
			records = append(records, position)
			if limit != 0 && len(records) >= int(limit) {
				break
			}
		}
	}

	return records, nil
}

func (mm *MemoryMap) Close() error {
	return nil
}
