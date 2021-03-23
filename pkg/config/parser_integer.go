package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type IntegerParser struct {
	Name             string `yaml:"name"`
	IgnoreCharacters string `yaml:"ignoreCharacters"`
	Logger           *logrus.Entry
}

func (config *IntegerParser) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("integer.name is required")
	}

	return nil
}

func (config *IntegerParser) getName() string {
	return config.Name
}
