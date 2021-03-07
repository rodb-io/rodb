package output

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"regexp"
	configModule "rods/pkg/config"
	indexModule "rods/pkg/index"
	parserModule "rods/pkg/parser"
	serviceModule "rods/pkg/service"
	"strconv"
	"strings"
)

type JsonObject struct {
	config       *configModule.JsonObjectOutput
	index        indexModule.Index
	service      serviceModule.Service
	paramParsers []parserModule.Parser
	logger       *logrus.Logger
	route        *serviceModule.Route
}

func NewJsonObject(
	config *configModule.JsonObjectOutput,
	index indexModule.Index,
	service serviceModule.Service,
	parsers parserModule.List,
	log *logrus.Logger,
) (*JsonObject, error) {
	paramParsers := make([]parserModule.Parser, len(config.Parameters))
	for i, param := range config.Parameters {
		parser, parserExists := parsers[param.Parser]
		if !parserExists {
			return nil, errors.New("Parser '" + param.Parser + "' does not exist")
		}
		paramParsers[i] = parser
	}

	jsonObject := &JsonObject{
		config:       config,
		index:        index,
		service:      service,
		paramParsers: paramParsers,
		logger:       log,
	}

	route := &serviceModule.Route{
		Endpoint:            jsonObject.endpointRegexp(),
		ExpectedPayloadType: nil,
		ResponseType:        "application/json",
		Handler:             jsonObject.getHandler(),
	}

	jsonObject.route = route
	service.AddRoute(route)

	return jsonObject, nil
}

func (jsonObject *JsonObject) getHandler() func(params map[string]string, payload []byte) ([]byte, error) {
	return func(params map[string]string, payload []byte) ([]byte, error) {
		filters, err := jsonObject.getEndpointFilters(params)
		if err != nil {
			return nil, err
		}

		records, err := jsonObject.index.GetRecords(jsonObject.config.Input, filters, 1)
		if err != nil {
			return nil, err
		}

		if len(records) == 0 {
			return nil, serviceModule.RecordNotFoundError
		}

		data, err := records[0].All()
		if err != nil {
			return nil, err
		}

		return json.Marshal(data)
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

func (jsonObject *JsonObject) Close() error {
	jsonObject.service.DeleteRoute(jsonObject.route)
	return nil
}
