package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type GraphQLOutput struct {
	Name     string   `yaml:"name"`
	Endpoint string   `yaml:"endpoint"`
	Logger   *logrus.Entry
}

func (config *GraphQLOutput) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("graphql.name is required")
	}

	if config.Endpoint == "" {
		return errors.New("graphql.endpoint is not defined. This setting is required.")
	}

	return nil
}
