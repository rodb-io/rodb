package parser

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type IntegerConfig struct {
	Name             string `yaml:"name"`
	Type             string `yaml:"type"`
	IgnoreCharacters string `yaml:"ignoreCharacters"`
	Logger           *logrus.Entry
}

func (config *IntegerConfig) Validate(parsers map[string]Parser, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("integer.name is required")
	}

	return nil
}

func (config *IntegerConfig) GetName() string {
	return config.Name
}

func (config *IntegerConfig) Primitive() bool {
	return true
}
