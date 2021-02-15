package index

import (
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/input"
	"rods/pkg/record"
)

type MemoryMap struct {
	config *config.MemoryMapIndex
	input  input.Input
	logger *logrus.Logger
	index  map[interface{}][]record.Position
}

func NewMemoryMap(
	config *config.MemoryMapIndex,
	input input.Input,
	log *logrus.Logger,
) (*MemoryMap, error) {
	return &MemoryMap{
		config: config,
		logger: log,
	}, nil
}

func (mm *MemoryMap) Prepare() error {
	mm.index = make(map[interface{}][]record.Position)

	for result := range mm.input.IterateAll() {
		if result.Error != nil {
			return result.Error
		}

		value, err := result.Record.Get(mm.config.Column)
		if err != nil {
			return err
		}

		valueIndexes, valueIndexesExists := mm.index[value]
		if valueIndexesExists {
			valueIndexes = append(valueIndexes, result.Record.Position())
		} else {
			mm.index[value] = []record.Position{result.Record.Position()}
		}
	}

	return nil
}

func (mm *MemoryMap) DoesIndex(inputName string, columnName string) bool {
	return inputName == mm.config.Input && columnName == mm.config.Column
}

func (mm *MemoryMap) Close() error {
	return nil
}
