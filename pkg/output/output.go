package output

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/index"
	"rods/pkg/parser"
	"rods/pkg/service"
)

type Output interface {
	Close() error
}

type List = map[string]Output

func NewFromConfig(
	config config.Output,
	indexes index.List,
	services service.List,
	parsers parser.List,
	log *logrus.Logger,
) (Output, error) {
	if config.JsonObject != nil {
		service, serviceExists := services[config.JsonObject.Service]
		if !serviceExists {
			return nil, fmt.Errorf("Service '%v' not found in services list.", config.JsonObject.Service)
		}

		index, indexExists := indexes[config.JsonObject.Index]
		if !indexExists {
			return nil, fmt.Errorf("Index '%v' not found in indexes list.", config.JsonObject.Index)
		}

		return NewJsonObject(config.JsonObject, index, service, parsers, log)
	}

	return nil, errors.New("Failed to initialize output")
}

func NewFromConfigs(
	configs map[string]config.Output,
	indexes index.List,
	services service.List,
	parsers parser.List,
	log *logrus.Logger,
) (List, error) {
	outputs := make(List)
	for outputName, outputConfig := range configs {
		output, err := NewFromConfig(outputConfig, indexes, services, parsers, log)
		if err != nil {
			return nil, err
		}
		outputs[outputName] = output
	}

	return outputs, nil
}

func Close(outputs List) error {
	for _, output := range outputs {
		err := output.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
