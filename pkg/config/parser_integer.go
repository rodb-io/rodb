package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type IntegerParser struct {
	Name             string `yaml:"name"`
	Type             string `yaml:"type"`
	IgnoreCharacters string `yaml:"ignoreCharacters"`
	Logger           *logrus.Entry
}

func (config *IntegerParser) Validate(parsers map[string]Parser, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("integer.name is required")
	}

	return nil
}

func (config *IntegerParser) GetName() string {
	return config.Name
}

func (config *IntegerParser) Primitive() bool {
	return true
}
