package index

import (
	"fmt"
	configModule "rodb.io/pkg/config"
	"rodb.io/pkg/input"
	"rodb.io/pkg/record"
)

type Index interface {
	Name() string

	// Gets the indexed records matching all the given filters
	// The returned list is expected to be ordered from the
	// smallest position to the biggest
	GetRecordPositions(input input.Input, filters map[string]interface{}) (record.PositionIterator, error)

	Close() error
}

type List = map[string]Index

func NewFromConfig(
	config configModule.Index,
	inputs input.List,
) (Index, error) {
	switch config.(type) {
	case *configModule.MapIndex:
		return NewMap(config.(*configModule.MapIndex), inputs)
	case *configModule.WildcardIndex:
		return NewWildcard(config.(*configModule.WildcardIndex), inputs)
	case *configModule.NoopIndex:
		return NewNoop(config.(*configModule.NoopIndex), inputs), nil
	default:
		return nil, fmt.Errorf("Unknown index config type: %#v", config)
	}
}

func NewFromConfigs(
	configs map[string]configModule.Index,
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

func Close(indexes List) error {
	for _, index := range indexes {
		if err := index.Close(); err != nil {
			return err
		}
	}

	return nil
}
