package input

import (
	"fmt"
	configPackage "rodb.io/pkg/config"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/record"
	"time"
)

type Input interface {
	Name() string
	Get(position record.Position) (record.Record, error)
	Size() (int64, error)
	ModTime() (time.Time, error)

	// Iterates all the records in the input, ordered
	// from the smallest to the biggest position
	// The second returned parameter is a callback that
	// must be used to close the relevant resources
	IterateAll() (record.Iterator, func() error, error)

	Close() error
}

type List = map[string]Input

func NewFromConfig(
	config configPackage.Input,
	parsers parser.List,
) (Input, error) {
	switch config.(type) {
	case *configPackage.CsvInput:
		return NewCsv(config.(*configPackage.CsvInput), parsers)
	case *configPackage.XmlInput:
		return NewXml(config.(*configPackage.XmlInput), parsers)
	case *configPackage.JsonInput:
		return NewJson(config.(*configPackage.JsonInput))
	default:
		return nil, fmt.Errorf("Unknown input config type: %#v", config)
	}
}

func NewFromConfigs(
	configs map[string]configPackage.Input,
	parsers parser.List,
) (List, error) {
	inputs := make(List)
	for inputName, inputConfig := range configs {
		input, err := NewFromConfig(inputConfig, parsers)
		if err != nil {
			return nil, err
		}
		inputs[inputName] = input
	}

	return inputs, nil
}

func Close(inputs List) error {
	for _, input := range inputs {
		if err := input.Close(); err != nil {
			return err
		}
	}

	return nil
}
