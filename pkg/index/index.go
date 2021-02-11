package index

import (
	"errors"
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/input"
)

type Index interface {
	DoesIndex(inputName string, columnName string) bool
	Close() error
}

type IndexList = map[string]Index

func NewFromConfig(
	config config.Index,
	log *logrus.Logger,
) (Index, error) {
	if config.MemoryMap != nil {
		return NewMemoryMap(config.MemoryMap, log)
	}

	return nil, errors.New("Failed to initialize index")
}

func NewFromConfigs(
	configs map[string]config.Index,
	inputs input.InputList,
	log *logrus.Logger,
) (IndexList, error) {
	sources := make(IndexList)
	for sourceName, sourceConfig := range configs {
		index, err := NewFromConfig(sourceConfig, log)
		if err != nil {
			return nil, err
		}
		sources[sourceName] = index
	}

	dumbIndex, err := NewDumb(log)
	if err != nil {
		return nil, err
	}
	sources[""] = dumbIndex

	return sources, nil
}

func Close(sources IndexList) error {
	for _, index := range sources {
		err := index.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
