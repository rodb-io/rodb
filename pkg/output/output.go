package output

import (
	"fmt"
	"io"
	configPackage "rodb.io/pkg/config"
	"rodb.io/pkg/index"
	"rodb.io/pkg/input"
	"rodb.io/pkg/parser"
)

type Output interface {
	Name() string
	ExpectedPayloadType() *string
	ResponseType() string
	Handle(
		params map[string]string,
		payload []byte,
		sendError func(err error) error,
		sendSucces func() io.Writer,
	) error
	HasParameter(paramName string) bool
	GetParameterParser(paramName string) (parser.Parser, error)
	Close() error
}

type List = map[string]Output

func NewFromConfig(
	config configPackage.Output,
	inputs input.List,
	indexes index.List,
	parsers parser.List,
) (Output, error) {
	defaultIndex, defaultIndexExists := indexes["default"]
	if !defaultIndexExists {
		return nil, fmt.Errorf("Index 'default' not found in indexes list.")
	}

	switch config.(type) {
	case *configPackage.JsonObjectOutput:
		return NewJsonObject(config.(*configPackage.JsonObjectOutput), inputs, defaultIndex, indexes, parsers)
	case *configPackage.JsonArrayOutput:
		return NewJsonArray(config.(*configPackage.JsonArrayOutput), inputs, defaultIndex, indexes, parsers)
	default:
		return nil, fmt.Errorf("Unknown output config type: %#v", config)
	}
}

func NewFromConfigs(
	configs map[string]configPackage.Output,
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
		if err := output.Close(); err != nil {
			return err
		}
	}

	return nil
}
