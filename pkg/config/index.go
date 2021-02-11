package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type Index struct {
	MemoryMap *MemoryMapIndex `yaml:"memoryMap"`
}

func (config *Index) validate(log *logrus.Logger) error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("All indexes must have a configuration")
	}

	if len(fields) > 1 {
		return errors.New("An index can only have one configuration")
	}

	return fields[0].validate(log)
}
