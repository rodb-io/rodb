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

	output := &JsonObject{
		config:       config,
		index:        index,
		service:      service,
		paramParsers: paramParsers,
		logger:       log,
	}

	route := &serviceModule.Route{
		Endpoint:            output.endpointRegexp(),
		ExpectedPayloadType: nil,
		ResponseType:        "application/json",
		Handler:             output.getHandler(),
	}

	service.AddRoute(route)

	return output, nil
}

func (output *JsonObject) getHandler() func(params map[string]string, payload []byte) ([]byte, error) {
	return func(params map[string]string, payload []byte) ([]byte, error) {
		filters, err := output.getEndpointFilters(params)
		if err != nil {
			return nil, err
		}

		records, err := output.index.GetRecords(output.config.Input, filters, 1)
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

func (output *JsonObject) endpointRegexpParamName(index int) string {
	return "param_" + strconv.Itoa(index)
}

func (output *JsonObject) endpointRegexp() *regexp.Regexp {
	endpoint := output.config.Endpoint
	for i := range output.config.Parameters {
		paramPattern := output.paramParsers[i].GetRegexpPattern()
		paramName := output.endpointRegexpParamName(i)
		endpoint = strings.Replace(endpoint, "?", "(?P<"+paramName+">"+paramPattern+")", 1)
	}

	return regexp.MustCompile("^" + endpoint + "$")
}

func (output *JsonObject) getEndpointFilters(params map[string]string) (map[string]interface{}, error) {
	filters := map[string]interface{}{}
	for i, param := range output.config.Parameters {
		paramName := output.endpointRegexpParamName(i)
		paramValue, err := output.paramParsers[i].Parse(params[paramName])
		if err != nil {
			return nil, err
		}
		filters[param.Column] = paramValue
	}

	return filters, nil
}

func (output *JsonObject) Close() error {
	output.service.DeleteRoute(output.route)
	return nil
}
