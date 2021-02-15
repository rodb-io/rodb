package index

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/input"
)

type Index interface {
	Prepare() error
	DoesIndex(inputName string, columnName string) bool
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

	return nil, errors.New("Failed to initialize index")
}

func NewFromConfigs(
	configs map[string]config.Index,
	inputs input.List,
	log *logrus.Logger,
) (List, error) {
	sources := make(List)
	for sourceName, sourceConfig := range configs {
		index, err := NewFromConfig(sourceConfig, inputs, log)
		if err != nil {
			return nil, err
		}
		sources[sourceName] = index
	}

	dumbIndex, err := NewDumb(inputs, log)
	if err != nil {
		return nil, err
	}
	sources[""] = dumbIndex

	return sources, nil
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

func Prepare(indexes List) error {
	for indexName, index := range indexes {
		err := index.Prepare()
		if err != nil {
			return fmt.Errorf("Error preparing index '%v': %v", indexName, err)
		}
	}

	return nil
}
