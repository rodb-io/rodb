package index

import (
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/record"
)

type MemoryMap struct {
	config *config.MemoryMapIndex
	logger *logrus.Logger
	index map[interface{}]record.Position
}

func NewMemoryMap(
	config *config.MemoryMapIndex,
	log *logrus.Logger,
) (*MemoryMap, error) {
	return &MemoryMap{
		config: config,
		logger: log,
	}, nil
}

func (mm *MemoryMap) Prepare() bool {
	// TODO in record: check type and emit error in GetXXX if wrong type
	// TODO in record: have a generic GetXXX that returns an interface{} with the right type
}

func (mm *MemoryMap) DoesIndex(inputName string, columnName string) bool {
	return inputName == mm.config.Input && columnName == mm.config.Column
}

func (mm *MemoryMap) Close() error {
	return nil
}
