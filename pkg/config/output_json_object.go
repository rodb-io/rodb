package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

type JsonObjectOutput struct {
	Services      []string                                 `yaml:"services"`
	Input         string                                   `yaml:"input"`
	Endpoint      string                                   `yaml:"endpoint"`
	Index         string                                   `yaml:"index"`
	Parameters    []*JsonObjectOutputParameter             `yaml:"parameters"`
	Relationships map[string]*JsonObjectOutputRelationship `yaml:"relationships"`
	Logger        *logrus.Entry
}

type JsonObjectOutputParameter struct {
	Column string `yaml:"column"`
	Parser string `yaml:"parser"`
}

type JsonObjectOutputRelationship struct {
	Input         string                                   `yaml:"input"`
	Index         string                                   `yaml:"index"`
	IsArray       bool                                     `yaml:"isArray"`
	Limit         uint                                     `yaml:"limit"`
	Match         []*JsonObjectOutputRelationshipMatch     `yaml:"match"`
	Relationships map[string]*JsonObjectOutputRelationship `yaml:"relationships"`
}

type JsonObjectOutputRelationshipMatch struct {
	ParentColumn string `yaml:"parentColumn"`
	ChildColumn  string `yaml:"childColumn"`
}

func (config *JsonObjectOutput) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

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

	if config.Index == "" {
		log.Debugf("jsonObject.index is empty. Assuming 'default'.\n")
		config.Index = "default"
	}

	index, indexExists := rootConfig.Indexes[config.Index]
	if !indexExists {
		return fmt.Errorf("jsonObject.index: Index '%v' not found in indexes list.", config.Index)
	}
	if !index.DoesHandleInput(config.Input) {
		return fmt.Errorf("jsonObject.index: Index '%v' does not handle input '%v'.", config.Index, config.Input)
	}

	for parameterIndex, parameter := range config.Parameters {
		err := parameter.validate(rootConfig, log, config.Index, index)
		if err != nil {
			return fmt.Errorf("jsonObject.parameters.%v.%w", parameterIndex, err)
		}
	}

	for relationshipIndex, relationship := range config.Relationships {
		err := relationship.validate(rootConfig, log, config.Index, index)
		if err != nil {
			return fmt.Errorf("jsonObject.relationships.%v.%w", relationshipIndex, err)
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
	indexName string,
	index Index,
) error {
	if config.Column == "" {
		return errors.New("column is empty")
	}

	if config.Parser == "" {
		log.Debug("jsonObjet.parameters[].parser not defined. Assuming 'string'")
		config.Parser = "string"
	}
	_, parserExists := rootConfig.Parsers[config.Parser]
	if !parserExists {
		return fmt.Errorf("parser: Parser '%v' not found in parsers list.", config.Parser)
	}

	if !index.DoesHandleColumn(config.Column) {
		return fmt.Errorf("column: Index '%v' does not handle column '%v'.", indexName, config.Column)
	}

	return nil
}

func (config *JsonObjectOutputRelationship) validate(
	rootConfig *Config,
	log *logrus.Entry,
	parentIndexName string,
	parentIndex Index,
) error {
	if config.Limit == 0 && config.IsArray {
		log.Debug("jsonObject.relationships[].limit is not set. All relationships will be returned.")
	}

	if config.Index == "" {
		log.Debugf("jsonObject.relationships[].index is empty. Assuming 'default'.\n")
		config.Index = "default"
	}

	childIndex, childIndexExists := rootConfig.Indexes[config.Index]
	if !childIndexExists {
		return fmt.Errorf("index: Index '%v' not found in indexes list.", config.Index)
	}
	if !childIndex.DoesHandleInput(config.Input) {
		return fmt.Errorf("input: Index '%v' does not handle input '%v'.", config.Index, config.Input)
	}

	alreadyExistingChildColumn := make(map[string]bool)
	for matchIndex, match := range config.Match {
		err := match.validate(
			rootConfig,
			log,
			config.Index,
			childIndex,
		)
		if err != nil {
			return fmt.Errorf("match.%v.%w", matchIndex, err)
		}

		if _, alreadyExists := alreadyExistingChildColumn[match.ChildColumn]; alreadyExists {
			return fmt.Errorf("match.%v.childColumn: Duplicate filter on childColumn %v", matchIndex, match.ChildColumn)
		}
		alreadyExistingChildColumn[match.ChildColumn] = true
	}

	for relationshipName, relationship := range config.Relationships {
		err := relationship.validate(rootConfig, log, config.Index, childIndex)
		if err != nil {
			return fmt.Errorf("relationships.%v.%w", relationshipName, err)
		}
	}

	return nil
}

func (config *JsonObjectOutputRelationshipMatch) validate(
	rootConfig *Config,
	log *logrus.Entry,
	childIndexName string,
	childIndex Index,
) error {
	// The parentColumn and childColumn will be validated at runtime (must be validated against the input and not the index)

	if !childIndex.DoesHandleColumn(config.ChildColumn) {
		return fmt.Errorf("childColumn: Index '%v' does not handle column '%v'.", childIndexName, config.ChildColumn)
	}

	return nil
}
