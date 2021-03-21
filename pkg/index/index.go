package index

import (
	"errors"
	"rods/pkg/config"
	"rods/pkg/input"
	"rods/pkg/record"
)

type Index interface {
	// Gets the indexed records matching all the given filters
	// The returned list is expected to be ordered from the
	// smallest position to the biggest
	GetRecordPositions(inputName string, filters map[string]interface{}, limit uint) (record.PositionList, error)

	Close() error
}

type List = map[string]Index

func NewFromConfig(
	config config.Index,
	inputs input.List,
) (Index, error) {
	if config.MemoryMap != nil {
		return NewMemoryMap(config.MemoryMap, inputs)
	}
	if config.Noop != nil {
		return NewNoop(inputs), nil
	}

	return nil, errors.New("Failed to initialize index")
}

func NewFromConfigs(
	configs map[string]config.Index,
	inputs input.List,
) (List, error) {
	indexes := make(List)
	for indexName, indexConfig := range configs {
		index, err := NewFromConfig(indexConfig, inputs)
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
