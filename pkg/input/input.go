package input

import (
	"errors"
	"rods/pkg/config"
	"rods/pkg/source"
)

type Input interface {
	Close() error
}

type InputList = map[string]Input

func NewFromConfig(
	config config.InputConfig,
	sources source.SourceList,
) (Input, error) {
	if config.Csv != nil {
		return NewCsv(config.Csv, sources)
	}

	return nil, errors.New("Failed to initialize input")
}

func NewFromConfigs(
	configs map[string]config.InputConfig,
	sources source.SourceList,
) (InputList, error) {
	inputs := make(InputList)
	for inputName, InputConfig := range configs {
		input, err := NewFromConfig(InputConfig, sources)
		if err != nil {
			return nil, err
		}
		inputs[inputName] = input
	}

	return inputs, nil
}

func Close(inputs InputList) error {
	for _, input := range inputs {
		err := input.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
