package output

import (
	"errors"
	"fmt"
	indexPackage "github.com/rodb-io/rodb/pkg/index"
	inputPackage "github.com/rodb-io/rodb/pkg/input"
	recordPackage "github.com/rodb-io/rodb/pkg/input/record"
	relationshipPackage "github.com/rodb-io/rodb/pkg/output/relationship"
)

func checkRelationshipMatches(
	inputs inputPackage.List,
	relationship *relationshipPackage.RelationshipConfig,
	parentInput inputPackage.Input,
) error {
	for _, relationship := range relationship.Relationships {
		relationshipInput, inputExists := inputs[relationship.Input]
		if !inputExists {
			return fmt.Errorf("Input '%v' not found in inputs list.", relationship.Input)
		}

		if err := checkRelationshipMatches(inputs, relationship, relationshipInput); err != nil {
			return err
		}
	}

	return nil
}

func getRelationshipFiltersPerIndex(
	data map[string]interface{},
	matchConfig []*relationshipPackage.RelationshipMatchConfig,
	relationshipName string,
) (map[string]map[string]interface{}, error) {
	filtersPerIndex := map[string]map[string]interface{}{}
	for _, match := range matchConfig {
		matchData, matchPropertyExists := data[match.ParentProperty]
		if !matchPropertyExists {
			return nil, errors.New("Parent property '" + match.ParentProperty + "' does not exists in relationship '" + relationshipName + "'.")
		}

		indexFilters, indexFiltersExist := filtersPerIndex[match.ChildIndex]
		if !indexFiltersExist {
			indexFilters = make(map[string]interface{})
			filtersPerIndex[match.ChildIndex] = indexFilters
		}

		indexFilters[match.ChildProperty] = matchData
	}

	return filtersPerIndex, nil
}

func getFilteredRecordPositionsPerIndex(
	defaultIndex indexPackage.Index,
	indexes indexPackage.List,
	input inputPackage.Input,
	filtersPerIndex map[string]map[string]interface{},
) ([]recordPackage.PositionIterator, error) {
	if len(filtersPerIndex) == 0 {
		iterator, err := defaultIndex.GetRecordPositions(
			input,
			map[string]interface{}{},
		)
		if err != nil {
			return nil, err
		}

		return []recordPackage.PositionIterator{iterator}, nil
	}

	iterators := make([]recordPackage.PositionIterator, 0, len(filtersPerIndex))
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
	relationships map[string]*relationshipPackage.RelationshipConfig,
	defaultIndex indexPackage.Index,
	indexes indexPackage.List,
	inputs inputPackage.List,
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

		relationshipRecordPositionsIterator := recordPackage.JoinPositionIterators(relationshipRecordPositionsPerFilter...)

		relationshipRecords := make(recordPackage.List, 0)
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
	position recordPackage.Position,
	relationships map[string]*relationshipPackage.RelationshipConfig,
	defaultIndex indexPackage.Index,
	indexes indexPackage.List,
	inputs inputPackage.List,
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
