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
		var paramPattern string
		switch param.Type {
			case configModule.String:
				paramPattern = ".*"
			case configModule.Integer:
				paramPattern = "[-]?[0-9]+"
			case configModule.Float:
				paramPattern = "[-]?[0-9]+([.][0-9]+)?"
			case configModule.Boolean:
				paramPattern = "(true|false|1|0|TRUE|FALSE)"
			default:
				return nil, errors.New("Unknown param type '" + string(param.Type) + "'")
		}
		endpoint = strings.Replace(endpoint, "?", "(?P<" + paramName(i) + ">" + paramPattern + ")", 1)
	}

	// TODO commonize handling of the types?

	route := &serviceModule.Route{
		Endpoint:            regexp.MustCompile("^" + endpoint + "$"),
		ExpectedPayloadType: nil,
		ResponseType:        "application/json",
		Handler: func(params map[string]string, payload []byte) ([]byte, error) {
			filters := map[string]interface{} {}
			for i, param := range output.config.Parameters {
				switch param.Type {
					case configModule.String:
						filters[param.Column] = params[paramName(i)]
					case configModule.Integer:
						intParam, err := strconv.Atoi(params[paramName(i)])
						if err != nil {
							return nil, err
						}
						filters[param.Column] = intParam
					case configModule.Float:
						intParam, err := strconv.ParseFloat(params[paramName(i)], 64)
						if err != nil {
							return nil, err
						}
						filters[param.Column] = intParam
					case configModule.Boolean:
						paramString := params[paramName(i)]
						filters[param.Column] = (paramString == "true" || paramString == "1" || paramString == "TRUE")
					default:
						return nil, errors.New("Unknown param type '" + string(param.Type) + "'")
				}
			}

			output.index.GetRecords(output.config.Input, filters, 1)

			// TODO split this handler function properly (like the http service?)

			return nil, nil
		},
	}

	service.AddRoute(route)

	return output, nil
}

func (output *JsonObject) Close() error {
	output.service.DeleteRoute(output.route)
	return nil
}
