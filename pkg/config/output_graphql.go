package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type GraphQLOutput struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Logger *logrus.Entry
}

func (config *GraphQLOutput) GetName() string {
	return config.Name
}

func (config *GraphQLOutput) Validate(
	inputs map[string]Input,
	indexes map[string]Index,
	parsers map[string]Parser,
	log *logrus.Entry,
) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("graphql.name is required")
	}

	return nil
}
