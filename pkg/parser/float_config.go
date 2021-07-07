package parser

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type FloatConfig struct {
	Name             string `yaml:"name"`
	Type             string `yaml:"type"`
	DecimalSeparator string `yaml:"decimalSeparator"`
	IgnoreCharacters string `yaml:"ignoreCharacters"`
	Logger           *logrus.Entry
}

func (config *FloatConfig) Validate(parsers map[string]Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("float.name is required")
	}

	if len(config.DecimalSeparator) == 0 {
		return errors.New("float.decimalSeparator is required")
	}

	return nil
}

func (config *FloatConfig) GetName() string {
	return config.Name
}

func (config *FloatConfig) Primitive() bool {
	return true
}
