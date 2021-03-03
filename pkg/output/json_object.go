package output

import (
	"github.com/sirupsen/logrus"
	"errors"
	"regexp"
	configModule "rods/pkg/config"
	indexModule "rods/pkg/index"
	serviceModule "rods/pkg/service"
	"strconv"
	"strings"
)

type JsonObject struct {
	config  *configModule.JsonObjectOutput
	index   indexModule.Index
	service serviceModule.Service
	logger  *logrus.Logger
	route   *serviceModule.Route
}

func NewJsonObject(
	config *configModule.JsonObjectOutput,
	index indexModule.Index,
	service serviceModule.Service,
	log *logrus.Logger,
) (*JsonObject, error) {
	output := &JsonObject{
		config:  config,
		index:   index,
		service: service,
		logger:  log,
	}

	paramName := func (i int) string {
		return "param_" + strconv.Itoa(i)
	}

	// TODO move that in a function
	endpoint := output.config.Endpoint
	for i, param := range output.config.Parameters {
		paramPattern := param.TypeDefinition().GetRegexpPattern()
		endpoint = strings.Replace(endpoint, "?", "(?P<" + paramName(i) + ">" + paramPattern + ")", 1)
	}

	route := &serviceModule.Route{
		Endpoint:            regexp.MustCompile("^" + endpoint + "$"),
		ExpectedPayloadType: nil,
		ResponseType:        "application/json",
		Handler: func(params map[string]string, payload []byte) ([]byte, error) {
			filters := map[string]interface{} {}
			for i, param := range output.config.Parameters {
				paramValue, err := param.TypeDefinition().Parse(params[paramName(i)])
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

	// TODO Make the type handle the parts implemented in the record
	// TODO make the boolean type configurable like for the csv columns?

	service.AddRoute(route)

	return output, nil
}

func (output *JsonObject) Close() error {
	output.service.DeleteRoute(output.route)
	return nil
}
