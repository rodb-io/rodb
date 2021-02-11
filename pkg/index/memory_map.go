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

}

func (mm *MemoryMap) DoesIndex(inputName string, columnName string) bool {
	return inputName == mm.config.Input && columnName == mm.config.Column
}

func (mm *MemoryMap) Close() error {
	return nil
}
