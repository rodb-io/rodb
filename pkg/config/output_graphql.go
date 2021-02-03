package config

import (
	"errors"
)

type GraphQLOutputConfig struct{
	Service string
	Endpoint string
}

func (config *GraphQLOutputConfig) validate() error {
	// The service will be validated at runtime

	if config.Endpoint == "" {
		return errors.New("You must specify a non-empty GraphQL endpoint")
	}

	return nil
}
