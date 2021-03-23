package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type Service struct {
	Http *HttpService `yaml:"http"`
}

func (config *Service) validate(rootConfig *Config, log *logrus.Entry) error {
	definedFields := 0
	if config.Http != nil {
		definedFields++
		err := config.Http.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}

	if definedFields == 0 {
		return errors.New("All services must have a configuration")
	}
	if definedFields > 1 {
		return errors.New("A services can only have one configuration")
	}

	return nil
}

func (config *Service) Name() string {
	if config.Http != nil {
		return config.Http.Name
	}

	return ""
}
