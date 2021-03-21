package output

import (
	"errors"
	"rods/pkg/config"
	"rods/pkg/index"
	"rods/pkg/input"
	"rods/pkg/parser"
	"rods/pkg/service"
)

type Output interface {
	Close() error
}

type List = map[string]Output

func NewFromConfig(
	config config.Output,
	inputs input.List,
	indexes index.List,
	services service.List,
	parsers parser.List,
) (Output, error) {
	if config.JsonObject != nil {
		return NewJsonObject(config.JsonObject, inputs, indexes, services, parsers)
	}
	if config.JsonArray != nil {
		return NewJsonArray(config.JsonArray, inputs, indexes, services, parsers)
	}

	return nil, errors.New("Failed to initialize output")
}

func NewFromConfigs(
	configs map[string]config.Output,
	inputs input.List,
	indexes index.List,
	services service.List,
	parsers parser.List,
) (List, error) {
	outputs := make(List)
	for outputName, outputConfig := range configs {
		output, err := NewFromConfig(outputConfig, inputs, indexes, services, parsers)
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
