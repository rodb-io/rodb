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
)

type JsonObject struct {
	config         *configModule.JsonObjectOutput
	inputs         inputModule.List
	input          inputModule.Input
	defaultIndex   indexModule.Index
	indexes        indexModule.List
	paramParsers   map[string]parserModule.Parser
	endpoint       *regexp.Regexp
	endpointParams []string
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

	jsonObject.endpoint, jsonObject.endpointParams = jsonObject.createEndpointRegexp()

	for _, endpointParam := range jsonObject.endpointParams {
		_, endpointParamExists := jsonObject.config.Parameters[endpointParam]
		if !endpointParamExists {
			return nil, fmt.Errorf("The parameter '%v' set in the endpoint does not exists in the parameters list.", endpointParam)
		}
	}

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

// Returns a regular expression to match a string, and the list of param names
// (matching the sub-expressions of the regexp)
func (jsonObject *JsonObject) createEndpointRegexp() (*regexp.Regexp, []string) {
	paramRegexp := regexp.MustCompile("{([^}]+)}")

	paramMatches := paramRegexp.FindAllStringSubmatch(jsonObject.config.Endpoint, -1)
	params := make([]string, len(paramMatches))
	for i, paramMatch := range paramMatches {
		params[i] = paramMatch[1]
	}

	parts := paramRegexp.Split(jsonObject.config.Endpoint, -1)
	endpoint := parts[0]
	for partIndex := 1; partIndex < len(parts); partIndex++ {
		paramIndex := partIndex - 1
		paramName := params[paramIndex]

		if paramIndex >= len(jsonObject.paramParsers) {
			endpoint = endpoint + "(.*)" + parts[partIndex]
		} else {
			paramPattern := jsonObject.paramParsers[paramName].GetRegexpPattern()
			regexpParamName := jsonObject.endpointRegexpParamName(paramIndex)
			endpoint = endpoint + "(?P<" + regexpParamName + ">" + paramPattern + ")" + parts[partIndex]
		}
	}

	return regexp.MustCompile("^" + endpoint + "$"), params
}

func (jsonObject *JsonObject) getEndpointFiltersPerIndex(params map[string]string) (map[string]map[string]interface{}, error) {
	filtersPerIndex := map[string]map[string]interface{}{}
	for paramIndex, paramName := range jsonObject.endpointParams {
		param := jsonObject.config.Parameters[paramName]

		indexFilters, indexFiltersExists := filtersPerIndex[param.Index]
		if !indexFiltersExists {
			indexFilters = make(map[string]interface{})
			filtersPerIndex[param.Index] = indexFilters
		}

		regexpParamName := jsonObject.endpointRegexpParamName(paramIndex)
		paramValue, err := jsonObject.paramParsers[paramName].Parse(params[regexpParamName])
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
