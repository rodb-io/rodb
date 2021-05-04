package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

type JsonObjectOutput struct {
	Name          string                   `yaml:"name"`
	Type          string                   `yaml:"type"`
	Input         string                   `yaml:"input"`
	Parameters    map[string]*Parameter    `yaml:"parameters"`
	Relationships map[string]*Relationship `yaml:"relationships"`
	Logger        *logrus.Entry
}

func (config *JsonObjectOutput) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("jsonObject.name is required")
	}

	if len(config.Parameters) == 0 {
		return errors.New("jsonObject.parameters is empty. As least one is required.")
	}

	if config.Input == "" {
		return errors.New("jsonObject.input is empty. This field is required.")
	}
	input, inputExists := rootConfig.Inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("jsonObject.input: Input '%v' not found in inputs list.", config.Input)
	}

	for parameterName, parameter := range config.Parameters {
		logPrefix := fmt.Sprintf("jsonObject.parameters.%v.", parameterName)
		err := parameter.validate(rootConfig, log, logPrefix, input)
		if err != nil {
			return fmt.Errorf("jsonObject.parameters.%v.%w", parameterName, err)
		}
	}

	for relationshipIndex, relationship := range config.Relationships {
		logPrefix := fmt.Sprintf("jsonObject.relationships.%v.", relationshipIndex)
		err := relationship.validate(rootConfig, log, logPrefix)
		if err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}
	}

	return nil
}
