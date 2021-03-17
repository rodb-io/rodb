package input

import (
	"errors"
	"rods/pkg/config"
	"rods/pkg/parser"
	"rods/pkg/record"
	"rods/pkg/source"
)

type Input interface {
	Get(position record.Position) (record.Record, error)
	IterateAll() <-chan IterateAllResult
	Close() error
	Watch(watcher *source.Watcher) error
	CloseWatcher(watcher *source.Watcher) error
}

type List = map[string]Input

type IterateAllResult struct {
	Record record.Record
	Error  error
}

func NewFromConfig(
	config config.Input,
	sources source.List,
	parsers parser.List,
) (Input, error) {
	if config.Csv != nil {
		return NewCsv(config.Csv, sources, parsers)
	}

	return nil, errors.New("Failed to initialize input")
}

func NewFromConfigs(
	configs map[string]config.Input,
	sources source.List,
	parsers parser.List,
) (List, error) {
	inputs := make(List)
	for inputName, inputConfig := range configs {
		input, err := NewFromConfig(inputConfig, sources, parsers)
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
