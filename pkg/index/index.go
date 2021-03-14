package index

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/input"
	"rods/pkg/record"
)

type Index interface {
	GetRecords(inputName string, filters map[string]interface{}, limit uint) ([]record.Record, error)
	Close() error
}

type List = map[string]Index

func NewFromConfig(
	config config.Index,
	inputs input.List,
	log *logrus.Logger,
) (Index, error) {
	if config.MemoryMap != nil {
		if input, inputExists := inputs[config.MemoryMap.Input]; !inputExists {
			return nil, fmt.Errorf("Input '%v' not found in inputs list.", config.MemoryMap.Input)
		} else {
			return NewMemoryMap(config.MemoryMap, input, log)
		}
	}
	if config.Noop != nil {
		return NewNoop(inputs, log), nil
	}

	return nil, errors.New("Failed to initialize index")
}

func NewFromConfigs(
	configs map[string]config.Index,
	inputs input.List,
	log *logrus.Logger,
) (List, error) {
	indexes := make(List)
	for indexName, indexConfig := range configs {
		index, err := NewFromConfig(indexConfig, inputs, log)
		if err != nil {
			return nil, err
		}
		indexes[indexName] = index
	}

	return indexes, nil
}

func Close(sources List) error {
	for _, index := range sources {
		err := index.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
