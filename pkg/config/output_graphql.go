package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

type GraphQLOutput struct {
	Services []string `yaml:"services"`
	Endpoint string   `yaml:"endpoint"`
	Logger   *logrus.Entry
}

func (config *GraphQLOutput) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Endpoint == "" {
		return errors.New("graphql.endpoint is not defined. This setting is required.")
	}

	if len(config.Services) == 0 {
		return errors.New("graphql.services is empty. As least one is required.")
	}
	for _, serviceName := range config.Services {
		_, serviceExists := rootConfig.Services[serviceName]
		if !serviceExists {
			return fmt.Errorf("graphql.services: Service '%v' not found in services list.", serviceName)
		}
	}

	return nil
}
