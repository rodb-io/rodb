package input

import (
	"errors"
	"rodb.io/pkg/config"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/record"
)

type Input interface {
	Name() string
	Get(position record.Position) (record.Record, error)
	Size() (int64, error)

	// Iterates all the records in the input, ordered
	// from the smallest to the biggest position
	// The second returned parameter is a callback that
	// must be used to close the relevant resources
	IterateAll() (record.Iterator, func() error, error)

	Close() error
}

type List = map[string]Input

func NewFromConfig(
	config config.Input,
	parsers parser.List,
) (Input, error) {
	if config.Csv != nil {
		return NewCsv(config.Csv, parsers)
	}
	if config.Xml != nil {
		return NewXml(config.Xml, parsers)
	}
	if config.Json != nil {
		return NewJson(config.Json, parsers)
	}

	return nil, errors.New("Failed to initialize input")
}

func NewFromConfigs(
	configs map[string]config.Input,
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
		err := input.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
