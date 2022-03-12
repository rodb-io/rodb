package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	indexPackage "github.com/rodb-io/rodb/pkg/index"
	inputPackage "github.com/rodb-io/rodb/pkg/input"
	recordPackage "github.com/rodb-io/rodb/pkg/input/record"
	parserPackage "github.com/rodb-io/rodb/pkg/parser"
	"strconv"
)

type JsonArray struct {
	config       *JsonArrayConfig
	inputs       inputPackage.List
	input        inputPackage.Input
	defaultIndex indexPackage.Index
	indexes      indexPackage.List
	parsers      parserPackage.List
}

func NewJsonArray(
	config *JsonArrayConfig,
	inputs inputPackage.List,
	defaultIndex indexPackage.Index,
	indexes indexPackage.List,
	parsers parserPackage.List,
) (*JsonArray, error) {
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
	}

	for _, relationship := range jsonArray.config.Relationships {
		if err := checkRelationshipMatches(jsonArray.inputs, relationship, jsonArray.input); err != nil {
			return nil, err
		}
	}

	return jsonArray, nil
}

func (jsonArray *JsonArray) Name() string {
	return jsonArray.config.Name
}

func (jsonArray *JsonArray) ExpectedPayloadType() *string {
	return nil
}

func (jsonArray *JsonArray) ResponseType() string {
	return "application/json"
}

func (jsonArray *JsonArray) Handle(
	params map[string]string,
	payload []byte,
	sendError func(err error) error,
	sendSucces func() io.Writer,
) error {
	limit, err := jsonArray.getLimit(params)
	if err != nil {
		return sendError(err)
	}

	offset, err := jsonArray.getOffset(params)
	if err != nil {
		return sendError(err)
	}

	filtersPerIndex, err := jsonArray.getFiltersPerIndex(params)
	if err != nil {
		return sendError(err)
	}

	positionsPerIndex, err := getFilteredRecordPositionsPerIndex(
		jsonArray.defaultIndex,
		jsonArray.indexes,
		jsonArray.input,
		filtersPerIndex,
	)
	if err != nil {
		return sendError(err)
	}

	nextPosition := recordPackage.JoinPositionIterators(positionsPerIndex...)

	// Skipping rows depending on the offset
	for i := uint(0); i < offset; i++ {
		value, err := nextPosition()
		if err != nil {
			return sendError(err)
		}
		if value == nil {
			break
		}
	}

	rowsData := make([]interface{}, 0)
	for len(rowsData) < int(limit) {
		position, err := nextPosition()
		if err != nil {
			return sendError(err)
		}
		if position == nil {
			break
		}

		rowData, err := getDataFromPosition(
			*position,
			jsonArray.config.Relationships,
			jsonArray.defaultIndex,
			jsonArray.indexes,
			jsonArray.inputs,
			jsonArray.config.Input,
		)
		if err != nil {
			return sendError(err)
		}

		rowsData = append(rowsData, rowData)
	}

	return json.NewEncoder(sendSucces()).Encode(rowsData)
}

func (jsonArray *JsonArray) getLimit(params map[string]string) (uint, error) {
	limit := jsonArray.config.Limit.Default
	if limitParam, limitParamExists := params[jsonArray.config.Limit.Parameter]; limitParamExists {
		limitAsInt, err := strconv.Atoi(limitParam)
		if err != nil {
			return 0, err
		}
		if limitAsInt <= 0 {
			return 0, errors.New("The '" + jsonArray.config.Limit.Parameter + "' parameter must be a positive and non-zero number.")
		}
		limit = uint(limitAsInt)
	}
	if limit > jsonArray.config.Limit.Max {
		limit = jsonArray.config.Limit.Max
	}

	return limit, nil
}

func (jsonArray *JsonArray) getOffset(params map[string]string) (uint, error) {
	offset := uint(0)
	if offsetParam, offsetParamExists := params[jsonArray.config.Offset.Parameter]; offsetParamExists {
		offsetAsInt, err := strconv.Atoi(offsetParam)
		if err != nil {
			return 0, err
		}
		if offsetAsInt < 0 {
			return 0, errors.New("The '" + jsonArray.config.Offset.Parameter + "' parameter cannot be negative.")
		}
		offset = uint(offsetAsInt)
	}

	return offset, nil
}

func (jsonArray *JsonArray) getFiltersPerIndex(params map[string]string) (map[string]map[string]interface{}, error) {
	filtersPerIndex := make(map[string]map[string]interface{})
	for paramName, paramConfig := range jsonArray.config.Parameters {
		paramValue, paramExists := params[paramName]
		if !paramExists {
			continue
		}

		parser, parserExists := jsonArray.parsers[paramConfig.Parser]
		if !parserExists {
			return nil, errors.New("Parser '" + paramConfig.Parser + "' does not exist")
		}

		parsedParamValue, err := parser.Parse(paramValue)
		if err != nil {
			return nil, err
		}

		indexFilters, indexFiltersExists := filtersPerIndex[paramConfig.Index]
		if !indexFiltersExists {
			indexFilters = make(map[string]interface{})
			filtersPerIndex[paramConfig.Index] = indexFilters
		}

		indexFilters[paramConfig.Property] = parsedParamValue
	}

	return filtersPerIndex, nil
}

func (jsonArray *JsonArray) HasParameter(paramName string) bool {
	_, paramExists := jsonArray.config.Parameters[paramName]
	return paramExists
}

func (jsonArray *JsonArray) GetParameterParser(paramName string) (parserPackage.Parser, error) {
	parameter, parameterExists := jsonArray.config.Parameters[paramName]
	if !parameterExists {
		return nil, errors.New("Parameter '" + paramName + "' does not exist")
	}

	parser, parserExists := jsonArray.parsers[parameter.Parser]
	if !parserExists {
		return nil, errors.New("Parser '" + parameter.Parser + "' does not exist")
	}

	return parser, nil
}

func (jsonArray *JsonArray) Close() error {
	return nil
}
