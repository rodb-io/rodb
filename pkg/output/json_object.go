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
	serviceModule "rods/pkg/service"
	"strconv"
	"strings"
)

type JsonObject struct {
	config       *configModule.JsonObjectOutput
	inputs       inputModule.List
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

	jsonObject := &JsonObject{
		config:       config,
		inputs:       inputs,
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
	relationship *configModule.JsonObjectOutputRelationship,
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
		filters, err := jsonObject.getEndpointFilters(params)
		if err != nil {
			return sendError(err)
		}

		index, indexExists := jsonObject.indexes[jsonObject.config.Index]
		if !indexExists {
			return sendError(fmt.Errorf("Index '%v' not found in indexes list.", jsonObject.config.Index))
		}

		records, err := index.GetRecords(jsonObject.config.Input, filters, 1)
		if err != nil {
			return sendError(err)
		}

		if len(records) == 0 {
			return sendError(serviceModule.RecordNotFoundError)
		}

		data, err := records[0].All()
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

func (jsonObject *JsonObject) getEndpointFilters(params map[string]string) (map[string]interface{}, error) {
	filters := map[string]interface{}{}
	for i, param := range jsonObject.config.Parameters {
		paramName := jsonObject.endpointRegexpParamName(i)
		paramValue, err := jsonObject.paramParsers[i].Parse(params[paramName])
		if err != nil {
			return nil, err
		}
		filters[param.Column] = paramValue
	}

	return filters, nil
}

func (jsonObject *JsonObject) loadRelationships(
	data map[string]interface{},
	relationships map[string]*configModule.JsonObjectOutputRelationship,
) (map[string]interface{}, error) {
	for relationshipName, relationshipConfig := range relationships {
		filters := map[string]interface{}{}
		for _, match := range relationshipConfig.Match {
			matchData, matchColumnExists := data[match.ParentColumn]
			if !matchColumnExists {
				return nil, errors.New("Parent column '" + match.ParentColumn + "' does not exists in relationship '" + relationshipName + "'.")
			}

			filters[match.ChildColumn] = matchData
		}

		index, indexExists := jsonObject.indexes[relationshipConfig.Index]
		if !indexExists {
			return nil, fmt.Errorf("Index '%v' not found in indexes list.", relationshipConfig.Index)
		}

		var limit uint = 1
		if relationshipConfig.IsArray {
			limit = relationshipConfig.Limit
		}

		relationshipRecords, err := index.GetRecords(
			relationshipConfig.Input,
			filters,
			limit,
		)
		if err != nil {
			return nil, err
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
