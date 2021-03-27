package output

import (
	"errors"
	"fmt"
	configModule "rods/pkg/config"
	indexModule "rods/pkg/index"
	inputModule "rods/pkg/input"
	recordModule "rods/pkg/record"
)

func checkRelationshipMatches(
	inputs inputModule.List,
	relationship *configModule.Relationship,
	parentInput inputModule.Input,
) error {
	for _, sort := range relationship.Sort {
		if !parentInput.HasColumn(sort.Column) {
			return fmt.Errorf("Input '%v' does not have a column called '%v'.", parentInput.Name(), sort.Column)
		}
	}

	for _, match := range relationship.Match {
		if !parentInput.HasColumn(match.ParentColumn) {
			return fmt.Errorf("Input '%v' does not have a column called '%v'.", parentInput.Name(), match.ParentColumn)
		}
	}

	for _, relationship := range relationship.Relationships {
		relationshipInput, inputExists := inputs[relationship.Input]
		if !inputExists {
			return fmt.Errorf("Input '%v' not found in inputs list.", relationship.Input)
		}

		err := checkRelationshipMatches(
			inputs,
			relationship,
			relationshipInput,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func getRelationshipFiltersPerIndex(
	data map[string]interface{},
	matchConfig []*configModule.RelationshipMatch,
	relationshipName string,
) (map[string]map[string]interface{}, error) {
	filtersPerIndex := map[string]map[string]interface{}{}
	for _, match := range matchConfig {
		matchData, matchColumnExists := data[match.ParentColumn]
		if !matchColumnExists {
			return nil, errors.New("Parent column '" + match.ParentColumn + "' does not exists in relationship '" + relationshipName + "'.")
		}

		indexFilters, indexFiltersExist := filtersPerIndex[match.ChildIndex]
		if !indexFiltersExist {
			indexFilters = make(map[string]interface{})
			filtersPerIndex[match.ChildIndex] = indexFilters
		}

		indexFilters[match.ChildColumn] = matchData
	}

	return filtersPerIndex, nil
}

func getFilteredRecordPositionsPerIndex(
	defaultIndex indexModule.Index,
	indexes indexModule.List,
	input inputModule.Input,
	filtersPerIndex map[string]map[string]interface{},
) ([]recordModule.PositionIterator, error) {
	if len(filtersPerIndex) == 0 {
		iterator, err := defaultIndex.GetRecordPositions(
			input,
			map[string]interface{}{},
		)
		if err != nil {
			return nil, err
		}

		return []recordModule.PositionIterator{iterator}, nil
	}

	iterators := make([]recordModule.PositionIterator, 0, len(filtersPerIndex))
	for indexName, filters := range filtersPerIndex {
		index, indexExists := indexes[indexName]
		if !indexExists {
			return nil, fmt.Errorf("Index '%v' not found in indexes list.", indexName)
		}

		iteratorForThisIndex, err := index.GetRecordPositions(input, filters)
		if err != nil {
			return nil, err
		}

		iterators = append(iterators, iteratorForThisIndex)
	}

	return iterators, nil
}

func loadRelationships(
	data map[string]interface{},
	relationships map[string]*configModule.Relationship,
	defaultIndex indexModule.Index,
	indexes indexModule.List,
	inputs inputModule.List,
	rootInput string,
) (map[string]interface{}, error) {
	for relationshipName, relationshipConfig := range relationships {
		filtersPerIndex, err := getRelationshipFiltersPerIndex(
			data,
			relationshipConfig.Match,
			relationshipName,
		)
		if err != nil {
			return nil, err
		}

		input, ok := inputs[relationshipConfig.Input]
		if !ok {
			return nil, fmt.Errorf("There is no input named '%v'", relationshipConfig.Input)
		}

		relationshipRecordPositionsPerFilter, err := getFilteredRecordPositionsPerIndex(
			defaultIndex,
			indexes,
			input,
			filtersPerIndex,
		)
		if err != nil {
			return nil, err
		}

		input, inputExists := inputs[relationshipConfig.Input]
		if !inputExists {
			return nil, fmt.Errorf("Input '%v' not found in inputs list.", relationshipConfig.Input)
		}

		relationshipRecordPositionsIterator := recordModule.JoinPositionIterators(relationshipRecordPositionsPerFilter...)

		relationshipRecords := make(recordModule.List, 0)
		for {
			relationshipRecordPosition, err := relationshipRecordPositionsIterator()
			if err != nil {
				return nil, err
			}
			if relationshipRecordPosition == nil {
				break
			}

			relationshipRecord, err := input.Get(*relationshipRecordPosition)
			if err != nil {
				return nil, err
			}

			relationshipRecords = append(relationshipRecords, relationshipRecord)
		}

		if len(relationshipConfig.Sort) > 0 {
			relationshipRecords = relationshipRecords.Sort(relationshipConfig.Sort)
		}

		count := 0
		if relationshipConfig.IsArray {
			count = int(relationshipConfig.Limit)
		} else {
			count = 1
		}
		if count == 0 {
			count = len(relationshipRecords)
		} else if len(relationshipRecords) < count {
			count = len(relationshipRecords)
		}

		relationshipItems := make([]map[string]interface{}, 0, count)
		for _, relationshipRecord := range relationshipRecords {
			relationshipData, err := relationshipRecord.All()
			if err != nil {
				return nil, err
			}

			relationshipData, err = loadRelationships(
				relationshipData,
				relationshipConfig.Relationships,
				defaultIndex,
				indexes,
				inputs,
				relationshipConfig.Input,
			)
			if err != nil {
				return nil, err
			}

			relationshipItems = append(relationshipItems, relationshipData)
			if len(relationshipItems) >= count {
				break
			}
		}

		if relationshipConfig.IsArray {
			data[relationshipName] = relationshipItems
		} else {
			if len(relationshipItems) == 0 {
				data[relationshipName] = nil
			} else {
				data[relationshipName] = relationshipItems[0]
			}
		}
	}

	return data, nil
}

func getDataFromPosition(
	position recordModule.Position,
	relationships map[string]*configModule.Relationship,
	defaultIndex indexModule.Index,
	indexes indexModule.List,
	inputs inputModule.List,
	rootInput string,
) (map[string]interface{}, error) {
	input, inputExists := inputs[rootInput]
	if !inputExists {
		return nil, fmt.Errorf("Input '%v' not found in inputs list.", rootInput)
	}

	record, err := input.Get(position)
	if err != nil {
		return nil, err
	}

	data, err := record.All()
	if err != nil {
		return nil, err
	}

	data, err = loadRelationships(data, relationships, defaultIndex, indexes, inputs, rootInput)
	if err != nil {
		return nil, err
	}

	return data, nil
}
