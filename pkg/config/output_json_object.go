package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

type JsonObjectOutput struct {
	Name          string                       `yaml:"name"`
	Services      []string                     `yaml:"services"`
	Input         string                       `yaml:"input"`
	Endpoint      string                       `yaml:"endpoint"`
	Parameters    []*JsonObjectOutputParameter `yaml:"parameters"`
	Relationships map[string]*Relationship     `yaml:"relationships"`
	Logger        *logrus.Entry
}

type JsonObjectOutputParameter struct {
	Column string `yaml:"column"`
	Parser string `yaml:"parser"`
	Index  string `yaml:"index"`
}

func (config *JsonObjectOutput) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("jsonObject.name is required")
	}

	alreadyExistingServices := make(map[string]bool)
	for _, serviceName := range config.Services {
		_, serviceExists := rootConfig.Services[serviceName]
		if !serviceExists {
			return fmt.Errorf("jsonObject.services: Service '%v' not found in services list.", serviceName)
		}

		if _, alreadyExists := alreadyExistingServices[serviceName]; alreadyExists {
			return fmt.Errorf("jsonObject.services: Duplicate service '%v' in array.", serviceName)
		}
		alreadyExistingServices[serviceName] = true
	}

	if len(config.Parameters) == 0 {
		return errors.New("jsonObject.parameters is empty. As least one is required.")
	}

	if len(config.Services) == 0 {
		return errors.New("jsonObject.services is empty. As least one is required.")
	}

	if config.Input == "" {
		return errors.New("jsonObject.input is empty. This field is required.")
	}
	input, inputExists := rootConfig.Inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("memoryMap.input: Input '%v' not found in inputs list.", config.Input)
	}

	for parameterIndex, parameter := range config.Parameters {
		err := parameter.validate(rootConfig, log, input)
		if err != nil {
			return fmt.Errorf("jsonObject.parameters.%v.%w", parameterIndex, err)
		}
	}

	for relationshipIndex, relationship := range config.Relationships {
		logPrefix := fmt.Sprintf("jsonObject.relationships.%v.", relationshipIndex)
		err := relationship.validate(rootConfig, log, logPrefix)
		if err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}
	}

	if config.Endpoint == "" {
		return errors.New("jsonObject.endpoint is not defined. This setting is required")
	}

	if !strings.Contains(config.Endpoint, "?") {
		return errors.New("jsonObject.endpoint must specify the identifier's location with '?'. For example \"/product/?\".")
	}

	if strings.Count(config.Endpoint, "?") != len(config.Parameters) {
		return errors.New("jsonObject.parameters: The same number of parameters than occurences of '?' in the endpoint is required")
	}

	return nil
}

func (config *JsonObjectOutputParameter) validate(
	rootConfig *Config,
	log *logrus.Entry,
	input Input,
) error {
	if config.Column == "" {
		return errors.New("column is empty")
	}

	if config.Parser == "" {
		log.Debug("jsonObject.parameters[].parser not defined. Assuming 'string'")
		config.Parser = "string"
	}
	_, parserExists := rootConfig.Parsers[config.Parser]
	if !parserExists {
		return fmt.Errorf("parser: Parser '%v' not found in parsers list.", config.Parser)
	}

	if config.Index == "" {
		log.Debugf("index is empty. Assuming 'default'.\n")
		config.Index = "default"
	}

	index, indexExists := rootConfig.Indexes[config.Index]
	if !indexExists {
		return fmt.Errorf("index: Index '%v' not found in indexes list.", config.Index)
	}
	if !index.DoesHandleInput(input) {
		return fmt.Errorf("index: Index '%v' does not handle input '%v'.", config.Index, input.Name())
	}
	if !index.DoesHandleColumn(config.Column) {
		return fmt.Errorf("column: Index '%v' does not handle column '%v'.", config.Index, config.Column)
	}

	return nil
}
