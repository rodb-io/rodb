package input

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/parser"
	"rods/pkg/record"
	"rods/pkg/source"
)

type Input interface {
	Get(position record.Position) (record.Record, error)
	IterateAll() <-chan IterateAllResult
	Close() error
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
	log *logrus.Logger,
) (Input, error) {
	if config.Csv != nil {
		if source, sourceExists := sources[config.Csv.Source]; !sourceExists {
			return nil, fmt.Errorf("Source '%v' not found in sources list.", config.Csv.Source)
		} else {
			return NewCsv(config.Csv, source, parsers, log)
		}
	}

	return nil, errors.New("Failed to initialize input")
}

func NewFromConfigs(
	configs map[string]config.Input,
	sources source.List,
	parsers parser.List,
	log *logrus.Logger,
) (List, error) {
	inputs := make(List)
	for inputName, inputConfig := range configs {
		input, err := NewFromConfig(inputConfig, sources, parsers, log)
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
