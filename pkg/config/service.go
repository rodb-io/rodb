package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type ServiceConfig struct{
	Http *HttpServiceConfig `yaml:"http"`
}

func (config *ServiceConfig) validate(log *logrus.Logger) error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("All services must have a configuration")
	}

	if len(fields) > 0 {
		return errors.New("A services can only have one configuration")
	}

	return fields[0].validate(log)
}
