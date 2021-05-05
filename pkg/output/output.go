package output

import (
	"fmt"
	"io"
	configModule "rodb.io/pkg/config"
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
	config configModule.Output,
	inputs input.List,
	indexes index.List,
	parsers parser.List,
) (Output, error) {
	defaultIndex, defaultIndexExists := indexes["default"]
	if !defaultIndexExists {
		return nil, fmt.Errorf("Index 'default' not found in indexes list.")
	}

	switch config.(type) {
	case *configModule.JsonObjectOutput:
		return NewJsonObject(config.(*configModule.JsonObjectOutput), inputs, defaultIndex, indexes, parsers)
	case *configModule.JsonArrayOutput:
		return NewJsonArray(config.(*configModule.JsonArrayOutput), inputs, defaultIndex, indexes, parsers)
	default:
		return nil, fmt.Errorf("Unknown output config type: %#v", config)
	}
}

func NewFromConfigs(
	configs map[string]configModule.Output,
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
