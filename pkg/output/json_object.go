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
	"strconv"
	"strings"
)

type JsonObject struct {
	config       *configModule.JsonObjectOutput
	inputs       inputModule.List
	input        inputModule.Input
	defaultIndex indexModule.Index
	indexes      indexModule.List
	paramParsers []parserModule.Parser
	endpoint     *regexp.Regexp
}

func NewJsonObject(
	config *configModule.JsonObjectOutput,
	inputs inputModule.List,
	defaultIndex indexModule.Index,
	indexes indexModule.List,
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

	jsonObject.endpoint = jsonObject.createEndpointRegexp()

	for _, relationship := range jsonObject.config.Relationships {
		err := checkRelationshipMatches(
			jsonObject.inputs,
			relationship,
			jsonObject.input,
		)
		if err != nil {
			return nil, err
		}
	}

	return jsonObject, nil
}

func (jsonObject *JsonObject) Name() string {
	return jsonObject.config.Name
}

func (jsonObject *JsonObject) Endpoint() *regexp.Regexp {
	return jsonObject.endpoint
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
	filtersPerIndex, err := jsonObject.getEndpointFiltersPerIndex(params)
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

func (jsonObject *JsonObject) endpointRegexpParamName(index int) string {
	return "param_" + strconv.Itoa(index)
}

func (jsonObject *JsonObject) createEndpointRegexp() *regexp.Regexp {
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

func (jsonObject *JsonObject) Close() error {
	return nil
}
