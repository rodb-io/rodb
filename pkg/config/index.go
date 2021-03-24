package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type Index struct {
	MemoryMap *MemoryMapIndex `yaml:"memoryMap"`
	Noop      *NoopIndex      `yaml:"noop"`
}

func (config *Index) validate(rootConfig *Config, log *logrus.Entry) error {
	definedFields := 0
	if config.MemoryMap != nil {
		definedFields++
		err := config.MemoryMap.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}
	if config.Noop != nil {
		definedFields++
		err := config.Noop.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}

	if definedFields == 0 {
		return errors.New("All indexes must have a configuration")
	}
	if definedFields > 1 {
		return errors.New("An index can only have one configuration")
	}

	return nil
}

func (config *Index) Name() string {
	if config.MemoryMap != nil {
		return config.MemoryMap.Name
	}
	if config.Noop != nil {
		return config.Noop.Name
	}

	return ""
}

func (config *Index) DoesHandleColumn(column string) bool {
	if config.MemoryMap != nil {
		return config.MemoryMap.DoesHandleColumn(column)
	}
	if config.Noop != nil {
		return config.Noop.DoesHandleColumn(column)
	}

	return false
}

func (config *Index) DoesHandleInput(input Input) bool {
	if config.MemoryMap != nil {
		return config.MemoryMap.DoesHandleInput(input)
	}
	if config.Noop != nil {
		return config.Noop.DoesHandleInput(input)
	}

	return false
}
