package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type Index struct {
	MemoryMap *MemoryMapIndex `yaml:"memoryMap"`
	Noop      *NoopIndex      `yaml:"noop"`
}

func (config *Index) validate(rootConfig *Config, log *logrus.Logger) error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("All indexes must have a configuration")
	}

	if len(fields) > 1 {
		return errors.New("An index can only have one configuration")
	}

	return fields[0].validate(rootConfig, log)
}

func (config *Index) DoesHandleColumn(column string) bool {
	fields := getAllNonNilFields(config)
	if len(fields) == 0 {
		return false
	}

	switch fields[0].(type) {
	case *MemoryMapIndex:
		return fields[0].(*MemoryMapIndex).DoesHandleColumn(column)
	case *NoopIndex:
		return fields[0].(*NoopIndex).DoesHandleColumn(column)
	}

	return false
}

func (config *Index) DoesHandleInput(input string) bool {
	fields := getAllNonNilFields(config)
	if len(fields) == 0 {
		return false
	}

	switch fields[0].(type) {
	case *MemoryMapIndex:
		return fields[0].(*MemoryMapIndex).DoesHandleInput(input)
	case *NoopIndex:
		return fields[0].(*NoopIndex).DoesHandleInput(input)
	}

	return false
}
