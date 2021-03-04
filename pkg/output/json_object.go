package output

import (
	"github.com/sirupsen/logrus"
	"regexp"
	"errors"
	configModule "rods/pkg/config"
	indexModule "rods/pkg/index"
	serviceModule "rods/pkg/service"
	parserModule "rods/pkg/parser"
	"strconv"
	"strings"
)

type JsonObject struct {
	config  *configModule.JsonObjectOutput
	index   indexModule.Index
	service serviceModule.Service
	paramParsers []parserModule.Parser
	logger  *logrus.Logger
	route   *serviceModule.Route
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
		config:  config,
		index:   index,
		service: service,
		paramParsers: paramParsers,
		logger:  log,
	}

	paramName := func (i int) string {
		return "param_" + strconv.Itoa(i)
	}

	// TODO move that in a function
	endpoint := output.config.Endpoint
	for i := range output.config.Parameters {
		paramPattern := output.paramParsers[i].GetRegexpPattern()
		endpoint = strings.Replace(endpoint, "?", "(?P<" + paramName(i) + ">" + paramPattern + ")", 1)
	}

	route := &serviceModule.Route{
		Endpoint:            regexp.MustCompile("^" + endpoint + "$"),
		ExpectedPayloadType: nil,
		ResponseType:        "application/json",
		Handler: func(params map[string]string, payload []byte) ([]byte, error) {
			filters := map[string]interface{} {}
			for i, param := range output.config.Parameters {
				paramValue, err := output.paramParsers[i].Parse(params[paramName(i)])
				if err != nil {
					return nil, err
				}
				filters[param.Column] = paramValue
			}

			output.index.GetRecords(output.config.Input, filters, 1)

			// TODO split this handler function properly (like the http service?)

			return nil, nil
		},
	}

	// TODO use the parsers in the csv input (and move the code)
	// TODO test the parsers and json object

	service.AddRoute(route)

	return output, nil
}

func (output *JsonObject) Close() error {
	output.service.DeleteRoute(output.route)
	return nil
}
