package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type Service struct {
	Http *HttpService `yaml:"http"`
}

func (config *Service) validate(rootConfig *Config, log *logrus.Entry) error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("All services must have a configuration")
	}

	if len(fields) > 1 {
		return errors.New("A services can only have one configuration")
	}

	return fields[0].validate(rootConfig, log)
}

func (config *Service) getName() string {
	fields := getAllNonNilFields(config)
	if len(fields) > 0 {
		return fields[0].getName()
	}

	return ""
}
