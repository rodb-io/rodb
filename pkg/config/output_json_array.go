package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

type JsonArrayOutput struct {
	Name          string                              `yaml:"name"`
	Input         string                              `yaml:"input"`
	Endpoint      string                              `yaml:"endpoint"`
	Limit         JsonArrayOutputLimit                `yaml:"limit"`
	Offset        JsonArrayOutputOffset               `yaml:"offset"`
	Parameters    map[string]JsonArrayOutputParameter `yaml:"parameters"`
	Relationships map[string]*Relationship            `yaml:"relationships"`
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

type JsonArrayOutputParameter struct {
	Column string `yaml:"column"`
	Index  string `yaml:"index"`
	Parser string `yaml:"parser"`
}

func (config *JsonArrayOutput) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("jsonArray.name is required")
	}

	if config.Input == "" {
		return errors.New("jsonArray.input is empty. This field is required.")
	}

	if config.Endpoint == "" {
		return errors.New("jsonArray.endpoint is not defined. This setting is required")
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
		err := configParam.validate(rootConfig, log, logPrefix)
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

func (config *JsonArrayOutputParameter) validate(rootConfig *Config, log *logrus.Entry, logPrefix string) error {
	if config.Column == "" {
		return errors.New("column is empty")
	}

	if config.Index == "" {
		log.Debugf(logPrefix + "index is empty. Assuming 'default'.\n")
		config.Index = "default"
	}
	index, indexExists := rootConfig.Indexes[config.Index]
	if !indexExists {
		return fmt.Errorf("index: Index '%v' not found in indexes list.", config.Index)
	}
	if !index.DoesHandleColumn(config.Column) {
		return fmt.Errorf("column: Index '%v' does not handle column '%v'.", config.Index, config.Column)
	}

	if config.Parser == "" {
		log.Debug(logPrefix + "parser not defined. Assuming 'string'")
		config.Parser = "string"
	}
	_, parserExists := rootConfig.Parsers[config.Parser]
	if !parserExists {
		return fmt.Errorf("parser: Parser '%v' not found in parsers list.", config.Parser)
	}

	return nil
}
