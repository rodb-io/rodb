package output

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"rods/pkg/config"
	"rods/pkg/index"
	"rods/pkg/input"
	"rods/pkg/parser"
)

type Output interface {
	Name() string
	Endpoint() *regexp.Regexp
	ExpectedPayloadType() *string
	ResponseType() string
	Handle(
		params map[string]string,
		payload []byte,
		sendError func(err error) error,
		sendSucces func() io.Writer,
	) error
	Close() error
}

type List = map[string]Output

func NewFromConfig(
	config config.Output,
	inputs input.List,
	indexes index.List,
	parsers parser.List,
) (Output, error) {
	defaultIndex, defaultIndexExists := indexes["default"]
	if !defaultIndexExists {
		return nil, fmt.Errorf("Index 'default' not found in indexes list.")
	}

	if config.JsonObject != nil {
		return NewJsonObject(config.JsonObject, inputs, defaultIndex, indexes, parsers)
	}
	if config.JsonArray != nil {
		return NewJsonArray(config.JsonArray, inputs, defaultIndex, indexes, parsers)
	}

	return nil, errors.New("Failed to initialize output")
}

func NewFromConfigs(
	configs map[string]config.Output,
	inputs input.List,
	indexes index.List,
	parsers parser.List,
) (List, error) {
	outputs := make(List)
	for outputName, outputConfig := range configs {
		output, err := NewFromConfig(outputConfig, inputs, indexes, parsers)
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
