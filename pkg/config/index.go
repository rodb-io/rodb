package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/util"
)

type Index struct {
	MemoryMap     *MemoryMapIndex
	MemoryPartial *MemoryPartialIndex
	Noop          *NoopIndex
}

func (config *Index) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in index config: %w", err)
	}

	switch objectType {
	case "memoryMap":
		config.MemoryMap = &MemoryMapIndex{}
		return unmarshal(config.MemoryMap)
	case "memoryPartial":
		config.MemoryPartial = &MemoryPartialIndex{}
		return unmarshal(config.MemoryPartial)
	case "noop":
		config.Noop = &NoopIndex{}
		return unmarshal(config.Noop)
	default:
		return fmt.Errorf("Error in index config: Unknown type '%v'", objectType)
	}
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
	if config.MemoryPartial != nil {
		definedFields++
		err := config.MemoryPartial.validate(rootConfig, log)
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
	if config.MemoryPartial != nil {
		return config.MemoryPartial.Name
	}
	if config.Noop != nil {
		return config.Noop.Name
	}

	return ""
}

func (config *Index) DoesHandleProperty(property string) bool {
	if config.MemoryMap != nil {
		return config.MemoryMap.DoesHandleProperty(property)
	}
	if config.MemoryPartial != nil {
		return config.MemoryPartial.DoesHandleProperty(property)
	}
	if config.Noop != nil {
		return config.Noop.DoesHandleProperty(property)
	}

	return false
}

func (config *Index) DoesHandleInput(input Input) bool {
	if config.MemoryMap != nil {
		return config.MemoryMap.DoesHandleInput(input)
	}
	if config.MemoryPartial != nil {
		return config.MemoryPartial.DoesHandleInput(input)
	}
	if config.Noop != nil {
		return config.Noop.DoesHandleInput(input)
	}

	return false
}
