package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

type JsonArrayOutput struct {
	Name          string                   `yaml:"name"`
	Type          string                   `yaml:"type"`
	Input         string                   `yaml:"input"`
	Limit         JsonArrayOutputLimit     `yaml:"limit"`
	Offset        JsonArrayOutputOffset    `yaml:"offset"`
	Parameters    map[string]Parameter     `yaml:"parameters"`
	Relationships map[string]*Relationship `yaml:"relationships"`
	Logger        *logrus.Entry
}

type JsonArrayOutputLimit struct {
	Default   uint   `yaml:"default"`
	Max       uint   `yaml:"max"`
	Parameter string `yaml:"parameter"`
}

type JsonArrayOutputOffset struct {
	Parameter string `yaml:"parameter"`
}

func (config *JsonArrayOutput) GetName() string {
	return config.Name
}

func (config *JsonArrayOutput) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("jsonArray.name is required")
	}

	if config.Input == "" {
		return errors.New("jsonArray.input is empty. This field is required.")
	}
	input, inputExists := rootConfig.Inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("jsonObject.input: Input '%v' not found in inputs list.", config.Input)
	}

	err := config.Limit.validate(rootConfig, log)
	if err != nil {
		return err
	}

	err = config.Offset.validate(rootConfig, log)
	if err != nil {
		return err
	}

	for configParamName, configParam := range config.Parameters {
		logPrefix := fmt.Sprintf("jsonArray.parameters.%v.", configParamName)
		err := configParam.validate(rootConfig, log, logPrefix, input)
		if err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}

		if configParamName == config.Limit.Parameter {
			return fmt.Errorf("jsonArray.parameters.%v: Parameter '%v' is already used for the limit", configParamName, configParamName)
		}
		if configParamName == config.Offset.Parameter {
			return fmt.Errorf("jsonArray.parameters.%v: Parameter '%v' is already used for the offset", configParamName, configParamName)
		}
	}

	for relationshipIndex, relationship := range config.Relationships {
		logPrefix := fmt.Sprintf("jsonArray.relationships.%v.", relationshipIndex)
		err := relationship.validate(rootConfig, log, logPrefix)
		if err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}
	}

	return nil
}

func (config *JsonArrayOutputLimit) validate(rootConfig *Config, log *logrus.Entry) error {
	if config.Default == 0 {
		log.Debug("jsonArray.limit.default not set. Assuming '100'")
		config.Default = 100
	}

	if config.Max == 0 {
		log.Debug("jsonArray.limit.max not set. Assuming '1000'")
		config.Max = 1000
	}

	if config.Parameter == "" {
		log.Debug("jsonArray.limit.parameter not set. Assuming 'limit'")
		config.Parameter = "limit"
	}

	return nil
}

func (config *JsonArrayOutputOffset) validate(rootConfig *Config, log *logrus.Entry) error {
	if config.Parameter == "" {
		log.Debug("jsonArray.offset.parameter not set. Assuming 'offset'")
		config.Parameter = "offset"
	}

	return nil
}
