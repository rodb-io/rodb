package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

type WildcardIndex struct {
	Name       string   `yaml:"name"`
	Type       string   `yaml:"type"`
	Path       string   `yaml:"path"`
	Input      string   `yaml:"input"`
	Properties []string `yaml:"properties"`
	IgnoreCase *bool    `yaml:"ignoreCase"`
	Logger     *logrus.Entry
}

func (config *WildcardIndex) ShouldIgnoreCase() bool {
	return config.IgnoreCase != nil && *config.IgnoreCase
}

func (config *WildcardIndex) Validate(inputs map[string]Input, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("wildcard.name is required")
	}

	if config.Path == "" {
		return errors.New("wildcard.path is required")
	}
	fileInfo, err := os.Stat(config.Path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("wildcard.path: Error checking the path: %w", err)
	}
	if err == nil && fileInfo.IsDir() {
		return errors.New("wildcard.path: This path already exists and is a directory")
	}

	_, inputExists := inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("wildcard.input: Input '%v' not found in inputs list.", config.Input)
	}

	alreadyExistingProperties := make(map[string]bool)
	for _, propertyName := range config.Properties {
		if _, alreadyExists := alreadyExistingProperties[propertyName]; alreadyExists {
			return fmt.Errorf("wildcard.properties: Duplicate property '%v' in array.", propertyName)
		}
		alreadyExistingProperties[propertyName] = true
	}

	if config.IgnoreCase == nil {
		log.Debug("wildcard.ignoreCase not set. Assuming false")
		falseValue := false
		config.IgnoreCase = &falseValue
	}

	// The properties validity will be validated at runtime

	return nil
}

func (config *WildcardIndex) GetName() string {
	return config.Name
}

func (config *WildcardIndex) DoesHandleProperty(property string) bool {
	isHandled := false
	for _, handledProperty := range config.Properties {
		if property == handledProperty {
			isHandled = true
			break
		}
	}

	return isHandled
}

func (config *WildcardIndex) DoesHandleInput(input Input) bool {
	return input.GetName() == config.Input
}
