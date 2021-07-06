package index

import (
	"fmt"
	configPackage "rodb.io/pkg/config"
	"rodb.io/pkg/input"
	"rodb.io/pkg/input/record"
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
	config configPackage.Index,
	inputs input.List,
) (Index, error) {
	switch config.(type) {
	case *configPackage.MapIndex:
		return NewMap(config.(*configPackage.MapIndex), inputs)
	case *configPackage.WildcardIndex:
		return NewWildcard(config.(*configPackage.WildcardIndex), inputs)
	case *configPackage.NoopIndex:
		return NewNoop(config.(*configPackage.NoopIndex), inputs), nil
	default:
		return nil, fmt.Errorf("Unknown index config type: %#v", config)
	}
}

func NewFromConfigs(
	configs map[string]configPackage.Index,
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
