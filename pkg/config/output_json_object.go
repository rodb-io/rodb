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

func (config *JsonObjectOutput) GetName() string {
	return config.Name
}

func (config *JsonObjectOutput) validate(
	inputs map[string]Input,
	indexes map[string]Index,
	parsers map[string]Parser,
	log *logrus.Entry,
) error {
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
	input, inputExists := inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("jsonObject.input: Input '%v' not found in inputs list.", config.Input)
	}

	for parameterName, parameter := range config.Parameters {
		logPrefix := fmt.Sprintf("jsonObject.parameters.%v.", parameterName)
		if err := parameter.validate(indexes, parsers, log, logPrefix, input); err != nil {
			return fmt.Errorf("jsonObject.parameters.%v.%w", parameterName, err)
		}
	}

	for relationshipIndex, relationship := range config.Relationships {
		logPrefix := fmt.Sprintf("jsonObject.relationships.%v.", relationshipIndex)
		if err := relationship.validate(indexes, inputs, log, logPrefix); err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}
	}

	return nil
}
