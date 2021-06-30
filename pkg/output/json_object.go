package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	configModule "rodb.io/pkg/config"
	indexModule "rodb.io/pkg/index"
	inputModule "rodb.io/pkg/input"
	parserModule "rodb.io/pkg/parser"
	recordModule "rodb.io/pkg/record"
)

type JsonObject struct {
	config       *configModule.JsonObjectOutput
	inputs       inputModule.List
	input        inputModule.Input
	defaultIndex indexModule.Index
	indexes      indexModule.List
	paramParsers map[string]parserModule.Parser
}

func NewJsonObject(
	config *configModule.JsonObjectOutput,
	inputs inputModule.List,
	defaultIndex indexModule.Index,
	indexes indexModule.List,
	parsers parserModule.List,
) (*JsonObject, error) {
	paramParsers := make(map[string]parserModule.Parser)
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

	nextPosition := recordModule.JoinPositionIterators(positionsPerIndex...)

	position, err := nextPosition()
	if err != nil {
		return sendError(err)
	}

	if position == nil {
		return sendError(recordModule.RecordNotFoundError)
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

func (jsonObject *JsonObject) GetParameterParser(paramName string) (parserModule.Parser, error) {
	parser, parserExists := jsonObject.paramParsers[paramName]
	if !parserExists {
		return nil, errors.New("Parameter '" + paramName + "' does not exist")
	}

	return parser, nil
}

func (jsonObject *JsonObject) Close() error {
	return nil
}
