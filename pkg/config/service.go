package config

import (
	"errors"
)

type ServiceConfig struct{
	Http *HttpServiceConfig
}

func (config *ServiceConfig) validate() error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("All services must have a configuration")
	}

	if len(fields) > 0 {
		return errors.New("A services can only have one configuration")
	}

	return fields[0].validate()
}
