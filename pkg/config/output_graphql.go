package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type GraphQLOutputConfig struct{
	Service string
	Endpoint string
}

func (config *GraphQLOutputConfig) validate(log *logrus.Logger) error {
	// The service will be validated at runtime

	if config.Endpoint == "" {
		return errors.New("graphql.endpoint is not defined. This setting is required.")
	}

	return nil
}
