package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type InputConfig struct{
	Csv *CsvInputConfig
}

func (config *InputConfig) validate(log *logrus.Logger) error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("All inputs must have a configuration")
	}

	if len(fields) > 0 {
		return errors.New("An input can only have one configuration")
	}

	return fields[0].validate(log)
}
