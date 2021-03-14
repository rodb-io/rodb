package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type Input struct {
	Csv *CsvInput `yaml:"csv"`
}

func (config *Input) validate(rootConfig *Config, log *logrus.Entry) error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("All inputs must have a configuration")
	}

	if len(fields) > 1 {
		return errors.New("An input can only have one configuration")
	}

	return fields[0].validate(rootConfig, log)
}
