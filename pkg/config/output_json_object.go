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

func (config *JsonObjectOutput) validate(rootConfig *Config, log *logrus.Logger) error {
	// The service will be validated at runtime
	// The default index value "" matches the dumb index

	if len(config.Parameters) == 0 {
		return errors.New("jsonObject.parameters is empty. As least one is required.")
	}

	if len(config.Services) == 0 {
		return errors.New("jsonObject.services is empty. As least one is required.")
	}

	index, indexExists := rootConfig.Indexes[config.Index]
	if !indexExists {
		return fmt.Errorf("Index '%v' not found in indexes list.", config.Index)
	}
	if !index.DoesHandleInput(config.Input) {
		return fmt.Errorf("Index '%v' does not handle input '%v'.", config.Index, config.Input)
	}

	for _, parameter := range config.Parameters {
		err := parameter.validate(rootConfig, log, config.Index, index)
		if err != nil {
			return err
		}
	}

	for _, relationship := range config.Relationships {
		err := relationship.validate(rootConfig, log, config.Index, index)
		if err != nil {
			return err
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
	log *logrus.Logger,
	indexName string,
	index Index,
) error {
	// The parser will be validated at runtime

	if config.Column == "" {
		return errors.New("jsonObject.parameters[].column is empty")
	}

	if config.Parser == "" {
		log.Debug("jsonObjet.parameters[].parser not defined. Assuming 'string'")
		config.Parser = "string"
	}

	if !index.DoesHandleColumn(config.Column) {
		return fmt.Errorf("Index '%v' does not handle column '%v'.", indexName, config.Column)
	}

	return nil
}

func (config *JsonObjectOutputRelationship) validate(
	rootConfig *Config,
	log *logrus.Logger,
	parentIndexName string,
	parentIndex Index,
) error {
	if config.Limit == 0 && config.IsArray {
		log.Debug("jsonObjet.relationships[].limit is not set. All relationships will be returned.")
	}

	childIndex, childIndexExists := rootConfig.Indexes[config.Index]
	if !childIndexExists {
		return fmt.Errorf("Index '%v' not found in indexes list.", config.Index)
	}
	if !childIndex.DoesHandleInput(config.Input) {
		return fmt.Errorf("Index '%v' does not handle input '%v'.", config.Index, config.Input)
	}

	for _, match := range config.Match {
		err := match.validate(
			rootConfig,
			log,
			parentIndexName,
			parentIndex,
			config.Index,
			childIndex,
		)
		if err != nil {
			return err
		}
	}

	for _, relationship := range config.Relationships {
		err := relationship.validate(rootConfig, log, config.Index, childIndex)
		if err != nil {
			return err
		}
	}

	return nil
}

func (config *JsonObjectOutputRelationshipMatch) validate(
	rootConfig *Config,
	log *logrus.Logger,
	parentIndexName string,
	parentIndex Index,
	childIndexName string,
	childIndex Index,
) error {
	if !parentIndex.DoesHandleColumn(config.ParentColumn) {
		return fmt.Errorf("Index '%v' does not handle column '%v'.", parentIndexName, config.ParentColumn)
	}
	if !childIndex.DoesHandleColumn(config.ChildColumn) {
		return fmt.Errorf("Index '%v' does not handle column '%v'.", childIndexName, config.ChildColumn)
	}

	return nil
}
