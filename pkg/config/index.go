package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type IndexConfig struct{
	MemoryMapUnique *MemoryMapUniqueIndexConfig `yaml:"memoryMapUnique"`
	MemoryMapMultiple *MemoryMapMultipleIndexConfig `yaml:"memoryMapMultiple"`
}

func (config *IndexConfig) validate(log *logrus.Logger) error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("All indexes must have a configuration")
	}

	if len(fields) > 0 {
		return errors.New("An index can only have one configuration")
	}

	return fields[0].validate(log)
}
