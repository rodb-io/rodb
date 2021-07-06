package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type FloatParser struct {
	Name             string `yaml:"name"`
	Type             string `yaml:"type"`
	DecimalSeparator string `yaml:"decimalSeparator"`
	IgnoreCharacters string `yaml:"ignoreCharacters"`
	Logger           *logrus.Entry
}

func (config *FloatParser) validate(parsers map[string]Parser, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("float.name is required")
	}

	if len(config.DecimalSeparator) == 0 {
		return errors.New("float.decimalSeparator is required")
	}

	return nil
}

func (config *FloatParser) GetName() string {
	return config.Name
}

func (config *FloatParser) Primitive() bool {
	return true
}
