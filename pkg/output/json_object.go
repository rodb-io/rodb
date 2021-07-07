package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	indexPackage "rodb.io/pkg/index"
	inputPackage "rodb.io/pkg/input"
	recordPackage "rodb.io/pkg/input/record"
	parserPackage "rodb.io/pkg/parser"
)

type JsonObject struct {
	config       *JsonObjectConfig
	inputs       inputPackage.List
	input        inputPackage.Input
	defaultIndex indexPackage.Index
	indexes      indexPackage.List
	paramParsers map[string]parserPackage.Parser
}

func NewJsonObject(
	config *JsonObjectConfig,
	inputs inputPackage.List,
	defaultIndex indexPackage.Index,
	indexes indexPackage.List,
	parsers parserPackage.List,
) (*JsonObject, error) {
	paramParsers := make(map[string]parserPackage.Parser)
	for paramName, param := range config.Parameters {
		parser, parserExists := parsers[param.Parser]
		if !parserExists {
			return nil, errors.New("Parser '" + param.Parser + "' does not exist")
		}
		paramParsers[paramName] = parser
	}

	input, ok := inputs[config.Input]
	if !ok {
		return nil, fmt.Errorf("There is no input named '%v'", config.Input)
	}

	jsonObject := &JsonObject{
		config:       config,
		inputs:       inputs,
		input:        input,
		defaultIndex: defaultIndex,
		indexes:      indexes,
		paramParsers: paramParsers,
	}

	for _, relationship := range jsonObject.config.Relationships {
		if err := checkRelationshipMatches(jsonObject.inputs, relationship, jsonObject.input); err != nil {
			return nil, err
		}
	}

	return jsonObject, nil
}

func (jsonObject *JsonObject) Name() string {
	return jsonObject.config.Name
}

func (jsonObject *JsonObject) ExpectedPayloadType() *string {
	return nil
}

func (jsonObject *JsonObject) ResponseType() string {
	return "application/json"
}

func (jsonObject *JsonObject) Handle(
	params map[string]string,
	payload []byte,
	sendError func(err error) error,
	sendSucces func() io.Writer,
) error {
	filtersPerIndex, err := jsonObject.getRouteFiltersPerIndex(params)
	if err != nil {
		return sendError(err)
	}

	positionsPerIndex, err := getFilteredRecordPositionsPerIndex(
		jsonObject.defaultIndex,
		jsonObject.indexes,
		jsonObject.input,
		filtersPerIndex,
	)
	if err != nil {
		return sendError(err)
	}

	nextPosition := recordPackage.JoinPositionIterators(positionsPerIndex...)

	position, err := nextPosition()
	if err != nil {
		return sendError(err)
	}

	if position == nil {
		return sendError(recordPackage.RecordNotFoundError)
	}

	data, err := getDataFromPosition(
		*position,
		jsonObject.config.Relationships,
		jsonObject.defaultIndex,
		jsonObject.indexes,
		jsonObject.inputs,
		jsonObject.config.Input,
	)
	if err != nil {
		return sendError(err)
	}

	return json.NewEncoder(sendSucces()).Encode(data)
}

func (jsonObject *JsonObject) getRouteFiltersPerIndex(params map[string]string) (map[string]map[string]interface{}, error) {
	filtersPerIndex := map[string]map[string]interface{}{}
	for paramName, param := range jsonObject.config.Parameters {
		indexFilters, indexFiltersExists := filtersPerIndex[param.Index]
		if !indexFiltersExists {
			indexFilters = make(map[string]interface{})
			filtersPerIndex[param.Index] = indexFilters
		}

		paramValue, err := jsonObject.paramParsers[paramName].Parse(params[paramName])
		if err != nil {
			return nil, err
		}
		indexFilters[param.Property] = paramValue
	}

	return filtersPerIndex, nil
}

func (jsonObject *JsonObject) HasParameter(paramName string) bool {
	_, paramExists := jsonObject.config.Parameters[paramName]
	return paramExists
}

func (jsonObject *JsonObject) GetParameterParser(paramName string) (parserPackage.Parser, error) {
	parser, parserExists := jsonObject.paramParsers[paramName]
	if !parserExists {
		return nil, errors.New("Parameter '" + paramName + "' does not exist")
	}

	return parser, nil
}

func (jsonObject *JsonObject) Close() error {
	return nil
}
