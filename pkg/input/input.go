package input

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
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
	log *logrus.Logger,
) (Input, error) {
	if config.Csv != nil {
		if source, sourceExists := sources[config.Csv.Source]; !sourceExists {
			return nil, fmt.Errorf("Source '%v' not found in sources list.", config.Csv.Source)
		} else {
			return NewCsv(config.Csv, source, log)
		}
	}

	return nil, errors.New("Failed to initialize input")
}

func NewFromConfigs(
	configs map[string]config.InputConfig,
	sources source.SourceList,
	log *logrus.Logger,
) (InputList, error) {
	inputs := make(InputList)
	for inputName, InputConfig := range configs {
		input, err := NewFromConfig(InputConfig, sources, log)
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
