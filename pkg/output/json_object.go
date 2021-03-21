package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	configModule "rods/pkg/config"
	indexModule "rods/pkg/index"
	inputModule "rods/pkg/input"
	parserModule "rods/pkg/parser"
	recordModule "rods/pkg/record"
	serviceModule "rods/pkg/service"
	"strconv"
	"strings"
)

type JsonObject struct {
	config       *configModule.JsonObjectOutput
	inputs       inputModule.List
	input        inputModule.Input
	indexes      indexModule.List
	services     []serviceModule.Service
	paramParsers []parserModule.Parser
	route        *serviceModule.Route
}

func NewJsonObject(
	config *configModule.JsonObjectOutput,
	inputs inputModule.List,
	indexes indexModule.List,
	services serviceModule.List,
	parsers parserModule.List,
) (*JsonObject, error) {
	paramParsers := make([]parserModule.Parser, len(config.Parameters))
	for i, param := range config.Parameters {
		parser, parserExists := parsers[param.Parser]
		if !parserExists {
			return nil, errors.New("Parser '" + param.Parser + "' does not exist")
		}
		paramParsers[i] = parser
	}

	outputServices := make([]serviceModule.Service, len(config.Services))
	for i, serviceName := range config.Services {
		service, serviceExists := services[serviceName]
		if !serviceExists {
			return nil, fmt.Errorf("Service '%v' not found in services list.", serviceName)
		}

		outputServices[i] = service
	}

	input, inputExists := inputs[config.Input]
	if !inputExists {
		return nil, fmt.Errorf("Input '%v' not found in inputs list.", config.Input)
	}

	jsonObject := &JsonObject{
		config:       config,
		inputs:       inputs,
		input:        input,
		indexes:      indexes,
		services:     outputServices,
		paramParsers: paramParsers,
	}

	for _, relationship := range jsonObject.config.Relationships {
		err := jsonObject.checkRelationshipMatches(
			relationship,
			jsonObject.config.Input,
		)
		if err != nil {
			return nil, err
		}
	}

	route := &serviceModule.Route{
		Endpoint:            jsonObject.endpointRegexp(),
		ExpectedPayloadType: nil,
		ResponseType:        "application/json",
		Handler:             jsonObject.getHandler(),
	}

	jsonObject.route = route

	for _, service := range jsonObject.services {
		service.AddRoute(route)
	}

	return jsonObject, nil
}

func (jsonObject *JsonObject) checkRelationshipMatches(
	relationship *configModule.Relationship,
	parentInputName string,
) error {
	input, inputExists := jsonObject.inputs[parentInputName]
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
		err := jsonObject.checkRelationshipMatches(
			relationship,
			relationship.Input,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (jsonObject *JsonObject) getHandler() serviceModule.RouteHandler {
	return func(
		params map[string]string,
		payload []byte,
		sendError func(err error) error,
		sendSucces func() io.Writer,
	) error {
		filtersPerIndex, err := jsonObject.getEndpointFiltersPerIndex(params)
		if err != nil {
			return sendError(err)
		}

		limit := uint(0)
		if len(filtersPerIndex) == 1 {
			limit = 1
		}

		positionsPerIndex, err := jsonObject.getFilteredRecordPositionsPerIndex(limit, filtersPerIndex)
		if err != nil {
			return sendError(err)
		}

		positions := recordModule.JoinPositionLists(1, positionsPerIndex...)
		if len(positions) == 0 {
			return sendError(serviceModule.RecordNotFoundError)
		}

		record, err := jsonObject.input.Get(positions[0])
		if err != nil {
			return sendError(err)
		}

		data, err := record.All()
		if err != nil {
			return sendError(err)
		}

		data, err = jsonObject.loadRelationships(data, jsonObject.config.Relationships)
		if err != nil {
			return sendError(err)
		}

		return json.NewEncoder(sendSucces()).Encode(data)
	}
}

func (jsonObject *JsonObject) endpointRegexpParamName(index int) string {
	return "param_" + strconv.Itoa(index)
}

func (jsonObject *JsonObject) endpointRegexp() *regexp.Regexp {
	parts := strings.Split(jsonObject.config.Endpoint, "?")

	endpoint := parts[0]
	for partIndex := 1; partIndex < len(parts); partIndex++ {
		paramIndex := partIndex - 1

		if paramIndex >= len(jsonObject.paramParsers) {
			endpoint = endpoint + "(.*)" + parts[partIndex]
		} else {
			paramPattern := jsonObject.paramParsers[paramIndex].GetRegexpPattern()
			paramName := jsonObject.endpointRegexpParamName(paramIndex)
			endpoint = endpoint + "(?P<" + paramName + ">" + paramPattern + ")" + parts[partIndex]
		}
	}

	return regexp.MustCompile("^" + endpoint + "$")
}

func (jsonObject *JsonObject) getEndpointFiltersPerIndex(params map[string]string) (map[string]map[string]interface{}, error) {
	filtersPerIndex := map[string]map[string]interface{}{}
	for i, param := range jsonObject.config.Parameters {
		indexFilters, indexFiltersExists := filtersPerIndex[param.Index]
		if !indexFiltersExists {
			indexFilters = make(map[string]interface{})
			filtersPerIndex[param.Index] = indexFilters
		}

		paramName := jsonObject.endpointRegexpParamName(i)
		paramValue, err := jsonObject.paramParsers[i].Parse(params[paramName])
		if err != nil {
			return nil, err
		}
		indexFilters[param.Column] = paramValue
	}

	return filtersPerIndex, nil
}

func (jsonObject *JsonObject) getRelationshipFiltersPerIndex(
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

func (jsonObject *JsonObject) getFilteredRecordPositionsPerIndex(
	limit uint,
	filtersPerIndex map[string]map[string]interface{},
) ([]recordModule.PositionList, error) {
	positionsPerIndex := make([]recordModule.PositionList, 0, len(filtersPerIndex))
	for indexName, filters := range filtersPerIndex {
		index, indexExists := jsonObject.indexes[indexName]
		if !indexExists {
			return nil, fmt.Errorf("Index '%v' not found in indexes list.", indexName)
		}

		positionsForThisIndex, err := index.GetRecordPositions(jsonObject.config.Input, filters, limit)
		if err != nil {
			return nil, err
		}

		positionsPerIndex = append(positionsPerIndex, positionsForThisIndex)
	}

	return positionsPerIndex, nil
}

func (jsonObject *JsonObject) loadRelationships(
	data map[string]interface{},
	relationships map[string]*configModule.Relationship,
) (map[string]interface{}, error) {
	for relationshipName, relationshipConfig := range relationships {
		filtersPerIndex, err := jsonObject.getRelationshipFiltersPerIndex(
			data,
			relationshipConfig.Match,
			relationshipName,
		)
		if err != nil {
			return nil, err
		}

		relationshipRecordPositionsPerFilter, err := jsonObject.getFilteredRecordPositionsPerIndex(0, filtersPerIndex)
		if err != nil {
			return nil, err
		}

		var limit uint = 1
		if relationshipConfig.IsArray {
			limit = relationshipConfig.Limit
		}

		input, inputExists := jsonObject.inputs[relationshipConfig.Input]
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

			relationshipData, err = jsonObject.loadRelationships(
				relationshipData,
				relationshipConfig.Relationships,
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

func (jsonObject *JsonObject) Close() error {
	for _, service := range jsonObject.services {
		service.DeleteRoute(jsonObject.route)
	}

	return nil
}
