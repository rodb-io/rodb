package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

type MapIndex struct {
	Name       string   `yaml:"name"`
	Type       string   `yaml:"type"`
	Input      string   `yaml:"input"`
	Properties []string `yaml:"properties"`
	Logger     *logrus.Entry
}

func (config *MapIndex) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("map.name is required")
	}

	_, inputExists := rootConfig.Inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("map.input: Input '%v' not found in inputs list.", config.Input)
	}

	alreadyExistingProperties := make(map[string]bool)
	for _, propertyName := range config.Properties {
		if _, alreadyExists := alreadyExistingProperties[propertyName]; alreadyExists {
			return fmt.Errorf("map.properties: Duplicate property '%v' in array.", propertyName)
		}
		alreadyExistingProperties[propertyName] = true
	}

	// The properties validity will be validated at runtime

	return nil
}

func (config *MapIndex) GetName() string {
	return config.Name
}

func (config *MapIndex) DoesHandleProperty(property string) bool {
	isHandled := false
	for _, handledProperty := range config.Properties {
		if property == handledProperty {
			isHandled = true
			break
		}
	}

	return isHandled
}

func (config *MapIndex) DoesHandleInput(input Input) bool {
	return input.GetName() == config.Input
}
