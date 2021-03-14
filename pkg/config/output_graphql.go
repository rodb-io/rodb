package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type GraphQLOutput struct {
	Services []string `yaml:"services"`
	Endpoint string   `yaml:"endpoint"`
}

func (config *GraphQLOutput) validate(rootConfig *Config, log *logrus.Entry) error {
	// The service will be validated at runtime

	if config.Endpoint == "" {
		return errors.New("graphql.endpoint is not defined. This setting is required.")
	}

	if len(config.Services) == 0 {
		return errors.New("graphql.services is empty. As least one is required.")
	}

	return nil
}
