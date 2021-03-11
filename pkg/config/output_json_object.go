package config

import (
	"errors"
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
	Match         []*JsonObjectOutputRelationshipMatch     `yaml:"match"`
	Relationships map[string]*JsonObjectOutputRelationship `yaml:"relationships"`
}

type JsonObjectOutputRelationshipMatch struct {
	ParentColumn string `yaml:"parentColumn"`
	ChildColumn  string `yaml:"childColumn"`
}

func (config *JsonObjectOutput) validate(log *logrus.Logger) error {
	// The service will be validated at runtime
	// The index will be validated at runtime
	// The default index value "" matches the dumb index
	// The input will be validated at runtime

	if len(config.Parameters) == 0 {
		return errors.New("jsonObject.parameters is empty. As least one is required.")
	}

	if len(config.Services) == 0 {
		return errors.New("jsonObject.services is empty. As least one is required.")
	}

	for _, parameter := range config.Parameters {
		err := parameter.validate(log)
		if err != nil {
			return err
		}
	}

	for _, relationship := range config.Relationships {
		err := relationship.validate(log)
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

func (config *JsonObjectOutputParameter) validate(log *logrus.Logger) error {
	// The existence of the column value will be validated at runtime
	// The parser will be validated at runtime

	if config.Column == "" {
		return errors.New("jsonObject.parameters[].column is empty")
	}

	if config.Parser == "" {
		log.Debug("jsonObjet.parameters[].parser not defined. Assuming 'string'")
		config.Parser = "string"
	}

	return nil
}

func (config *JsonObjectOutputRelationship) validate(log *logrus.Logger) error {
	// The index will be validated at runtime
	// The input will be validated at runtime

	for _, match := range config.Match {
		err := match.validate(log)
		if err != nil {
			return err
		}
	}

	for _, relationship := range config.Relationships {
		err := relationship.validate(log)
		if err != nil {
			return err
		}
	}

	return nil
}

func (config *JsonObjectOutputRelationshipMatch) validate(log *logrus.Logger) error {
	// The parentColumn will be validated at runtime
	// The childColumn will be validated at runtime
	return nil
}
