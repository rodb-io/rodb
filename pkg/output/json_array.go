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
)

type JsonArray struct {
	config       *configModule.JsonArrayOutput
	inputs       inputModule.List
	input        inputModule.Input
	defaultIndex indexModule.Index
	indexes      indexModule.List
	parsers      parserModule.List
	services     []serviceModule.Service
	route        *serviceModule.Route
}

func NewJsonArray(
	config *configModule.JsonArrayOutput,
	inputs inputModule.List,
	defaultIndex indexModule.Index,
	indexes indexModule.List,
	services serviceModule.List,
	parsers parserModule.List,
) (*JsonArray, error) {
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

	jsonArray := &JsonArray{
		config:       config,
		inputs:       inputs,
		input:        input,
		defaultIndex: defaultIndex,
		indexes:      indexes,
		parsers:      parsers,
		services:     outputServices,
	}

	for _, relationship := range jsonArray.config.Relationships {
		err := checkRelationshipMatches(
			jsonArray.inputs,
			relationship,
			jsonArray.config.Input,
		)
		if err != nil {
			return nil, err
		}
	}

	route := &serviceModule.Route{
		Endpoint:            regexp.MustCompile("^" + regexp.QuoteMeta(config.Endpoint) + "$"),
		ExpectedPayloadType: nil,
		ResponseType:        "application/json",
		Handler:             jsonArray.getHandler(),
	}

	jsonArray.route = route

	for _, service := range jsonArray.services {
		service.AddRoute(route)
	}

	return jsonArray, nil
}

func (jsonArray *JsonArray) Name() string {
	return jsonArray.config.Name
}

func (jsonArray *JsonArray) getHandler() serviceModule.RouteHandler {
	return func(
		params map[string]string,
		payload []byte,
		sendError func(err error) error,
		sendSucces func() io.Writer,
	) error {
		limit, err := jsonArray.getLimit(params)
		if err != nil {
			return sendError(err)
		}

		filtersPerIndex, err := jsonArray.getFiltersPerIndex(params)
		if err != nil {
			return sendError(err)
		}

		positionsPerIndex, err := getFilteredRecordPositionsPerIndex(
			jsonArray.indexes["default"],
			jsonArray.indexes,
			jsonArray.input,
			0,
			filtersPerIndex,
		)
		if err != nil {
			return sendError(err)
		}

		positions := recordModule.JoinPositionLists(limit, positionsPerIndex...)

		rowsData := make([]interface{}, len(positions))
		for i, position := range positions {
			rowsData[i], err = getDataFromPosition(
				position,
				jsonArray.config.Relationships,
				jsonArray.defaultIndex,
				jsonArray.indexes,
				jsonArray.inputs,
				jsonArray.config.Input,
			)
			if err != nil {
				return sendError(err)
			}
		}

		return json.NewEncoder(sendSucces()).Encode(rowsData)
	}
}

func (jsonArray *JsonArray) getLimit(params map[string]string) (uint, error) {
	limit := jsonArray.config.Limit.Default
	if limitParam, limitParamExists := params[jsonArray.config.Limit.Parameter]; limitParamExists {
		limitAsInt, err := strconv.Atoi(limitParam)
		if err != nil {
			return 0, err
		}
		limit = uint(limitAsInt)
	}
	if limit > jsonArray.config.Limit.Max {
		limit = jsonArray.config.Limit.Max
	}

	return limit, nil
}

func (jsonArray *JsonArray) getFiltersPerIndex(params map[string]string) (map[string]map[string]interface{}, error) {
	filtersPerIndex := make(map[string]map[string]interface{})
	for searchName, searchConfig := range jsonArray.config.Search {
		paramValue, paramExists := params[searchName]
		if !paramExists {
			continue
		}

		parser, parserExists := jsonArray.parsers[searchConfig.Parser]
		if !parserExists {
			return nil, errors.New("Parser '" + searchConfig.Parser + "' does not exist")
		}

		parsedParamValue, err := parser.Parse(paramValue)
		if err != nil {
			return nil, err
		}

		indexFilters, indexFiltersExists := filtersPerIndex[searchConfig.Index]
		if !indexFiltersExists {
			indexFilters = make(map[string]interface{})
			filtersPerIndex[searchConfig.Index] = indexFilters
		}

		indexFilters[searchConfig.Column] = parsedParamValue
	}

	return filtersPerIndex, nil
}

func (jsonArray *JsonArray) Close() error {
	for _, service := range jsonArray.services {
		service.DeleteRoute(jsonArray.route)
	}

	return nil
}
