package output

import (
	"github.com/sirupsen/logrus"
	"regexp"
	"rods/pkg/config"
	indexModule "rods/pkg/index"
	serviceModule "rods/pkg/service"
	"strings"
)

type JsonObject struct {
	config  *config.JsonObjectOutput
	index   indexModule.Index
	service serviceModule.Service
	logger  *logrus.Logger
	route   *serviceModule.Route
}

func NewJsonObject(
	config *config.JsonObjectOutput,
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

	endpoint := strings.Replace(config.Endpoint, "?", "(?P<id>.*)", 1)
	route := &serviceModule.Route{
		Endpoint:            regexp.MustCompile("^" + endpoint + "$"),
		ExpectedPayloadType: nil,
		ResponseType:        "application/json",
		Handler: func(params map[string]string, payload []byte) ([]byte, error) {
			output.index.GetRecords(
				output.config.Input,
				map[string]interface{} {
					"TODO": params["id"],
				},
				1
			)

			// TODO split this handler function properly (like the http service?)
			// TODO possibility of multiple params in the url: expose the regexp rather than a simple param?
			// TODO find the right param id for searching in the record
			// TODO what if types does not match (have int in record, but got string in the url)

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
