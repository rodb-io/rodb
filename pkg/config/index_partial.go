package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

type PartialIndex struct {
	Name       string   `yaml:"name"`
	Type       string   `yaml:"type"`
	Path       string   `yaml:"path"`
	Input      string   `yaml:"input"`
	Properties []string `yaml:"properties"`
	IgnoreCase *bool    `yaml:"ignoreCase"`
	Logger     *logrus.Entry
}

func (config *PartialIndex) ShouldIgnoreCase() bool {
	return config.IgnoreCase != nil && *config.IgnoreCase
}

func (config *PartialIndex) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("partial.name is required")
	}

	if config.Path == "" {
		return errors.New("partial.path is required")
	}
	fileInfo, err := os.Stat(config.Path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("partial.path: Error checking the path: %w", err)
	}
	if fileInfo.IsDir() {
		return errors.New("partial.path: This path already exists and is a directory")
	}

	_, inputExists := rootConfig.Inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("partial.input: Input '%v' not found in inputs list.", config.Input)
	}

	alreadyExistingProperties := make(map[string]bool)
	for _, propertyName := range config.Properties {
		if _, alreadyExists := alreadyExistingProperties[propertyName]; alreadyExists {
			return fmt.Errorf("partial.properties: Duplicate property '%v' in array.", propertyName)
		}
		alreadyExistingProperties[propertyName] = true
	}

	if config.IgnoreCase == nil {
		log.Debug("partial.ignoreCase not set. Assuming false")
		falseValue := false
		config.IgnoreCase = &falseValue
	}

	// The properties validity will be validated at runtime

	return nil
}

func (config *PartialIndex) GetName() string {
	return config.Name
}

func (config *PartialIndex) DoesHandleProperty(property string) bool {
	isHandled := false
	for _, handledProperty := range config.Properties {
		if property == handledProperty {
			isHandled = true
			break
		}
	}

	return isHandled
}

func (config *PartialIndex) DoesHandleInput(input Input) bool {
	return input.GetName() == config.Input
}
