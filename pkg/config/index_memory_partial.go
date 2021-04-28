package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

type MemoryPartialIndex struct {
	Name       string   `yaml:"name"`
	Input      string   `yaml:"input"`
	Properties []string `yaml:"properties"`
	Logger     *logrus.Entry
}

func (config *MemoryPartialIndex) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("memoryPartial.name is required")
	}

	_, inputExists := rootConfig.Inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("memoryPartial.input: Input '%v' not found in inputs list.", config.Input)
	}

	alreadyExistingProperties := make(map[string]bool)
	for _, propertyName := range config.Properties {
		if _, alreadyExists := alreadyExistingProperties[propertyName]; alreadyExists {
			return fmt.Errorf("memoryPartial.properties: Duplicate property '%v' in array.", propertyName)
		}
		alreadyExistingProperties[propertyName] = true
	}

	// The properties validity will be validated at runtime

	return nil
}

func (config *MemoryPartialIndex) DoesHandleProperty(property string) bool {
	isHandled := false
	for _, handledProperty := range config.Properties {
		if property == handledProperty {
			isHandled = true
			break
		}
	}

	return isHandled
}

func (config *MemoryPartialIndex) DoesHandleInput(input Input) bool {
	return input.Name() == config.Input
}
