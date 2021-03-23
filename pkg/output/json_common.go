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
	parentInputName string,
) error {
	input, inputExists := inputs[parentInputName]
	if !inputExists {
		return fmt.Errorf("Input '%v' not found in inputs list.", parentInputName)
	}

	for _, sort := range relationship.Sort {
		if !input.HasColumn(sort.Column) {
			return fmt.Errorf("Input '%v' does not have a column called '%v'.", parentInputName, sort.Column)
		}
	}

	for _, match := range relationship.Match {
		if !input.HasColumn(match.ParentColumn) {
			return fmt.Errorf("Input '%v' does not have a column called '%v'.", parentInputName, match.ParentColumn)
		}
	}

	for _, relationship := range relationship.Relationships {
		err := checkRelationshipMatches(
			inputs,
			relationship,
			relationship.Input,
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
	indexes indexModule.List,
	inputName string,
	limit uint,
	filtersPerIndex map[string]map[string]interface{},
) ([]recordModule.PositionList, error) {
	if len(filtersPerIndex) == 0 {
		index, indexExists := indexes["default"]
		if !indexExists {
			return nil, fmt.Errorf("Index 'default' not found in indexes list.")
		}

		records, err := index.GetRecordPositions(
			inputName,
			map[string]interface{}{},
			limit,
		)
		if err != nil {
			return nil, err
		}

		return []recordModule.PositionList{records}, nil
	}

	positionsPerIndex := make([]recordModule.PositionList, 0, len(filtersPerIndex))
	for indexName, filters := range filtersPerIndex {
		index, indexExists := indexes[indexName]
		if !indexExists {
			return nil, fmt.Errorf("Index '%v' not found in indexes list.", indexName)
		}

		positionsForThisIndex, err := index.GetRecordPositions(inputName, filters, limit)
		if err != nil {
			return nil, err
		}

		positionsPerIndex = append(positionsPerIndex, positionsForThisIndex)
	}

	return positionsPerIndex, nil
}

func loadRelationships(
	data map[string]interface{},
	relationships map[string]*configModule.Relationship,
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

		relationshipRecordPositionsPerFilter, err := getFilteredRecordPositionsPerIndex(indexes, relationshipConfig.Input, 0, filtersPerIndex)
		if err != nil {
			return nil, err
		}

		var limit uint = 1
		if relationshipConfig.IsArray {
			limit = relationshipConfig.Limit
		}

		input, inputExists := inputs[relationshipConfig.Input]
		if !inputExists {
			return nil, fmt.Errorf("Input '%v' not found in inputs list.", relationshipConfig.Input)
		}

		relationshipRecordPositions := recordModule.JoinPositionLists(limit, relationshipRecordPositionsPerFilter...)

		relationshipRecords := make(recordModule.List, len(relationshipRecordPositions))
		for i, position := range relationshipRecordPositions {
			relationshipRecords[i], err = input.Get(position)
			if err != nil {
				return nil, err
			}
		}

		if len(relationshipConfig.Sort) > 0 {
			relationshipRecords = relationshipRecords.Sort(relationshipConfig.Sort)
		}

		relationshipItems := make([]map[string]interface{}, 0, len(relationshipRecords))
		for _, relationshipRecord := range relationshipRecords {
			relationshipData, err := relationshipRecord.All()
			if err != nil {
				return nil, err
			}

			relationshipData, err = loadRelationships(
				relationshipData,
				relationshipConfig.Relationships,
				indexes,
				inputs,
				relationshipConfig.Input,
			)
			if err != nil {
				return nil, err
			}

			relationshipItems = append(relationshipItems, relationshipData)
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

	data, err = loadRelationships(data, relationships, indexes, inputs, rootInput)
	if err != nil {
		return nil, err
	}

	return data, nil
}
