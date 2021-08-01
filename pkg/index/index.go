package index

import (
	"fmt"
	"github.com/sirupsen/logrus"
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

type Config interface {
	Validate(inputs map[string]input.Config, log *logrus.Entry) error
	GetName() string
	DoesHandleProperty(property string) bool
	DoesHandleInput(input input.Config) bool
}

type List = map[string]Index

func NewFromConfig(
	config Config,
	inputs input.List,
) (Index, error) {
	switch config.(type) {
	case *MapConfig:
		return NewMap(config.(*MapConfig), inputs)
	case *WildcardConfig:
		return NewWildcard(config.(*WildcardConfig), inputs)
	case *SqliteConfig:
		return NewSqlite(config.(*SqliteConfig), inputs)
	case *Fts5Config:
		return NewFts5(config.(*Fts5Config), inputs)
	case *NoopConfig:
		return NewNoop(config.(*NoopConfig), inputs), nil
	default:
		return nil, fmt.Errorf("Unknown index config type: %#v", config)
	}
}

func NewFromConfigs(
	configs map[string]Config,
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
