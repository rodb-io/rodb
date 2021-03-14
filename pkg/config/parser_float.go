package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type FloatParser struct {
	DecimalSeparator string `yaml:"decimalSeparator"`
	IgnoreCharacters string `yaml:"ignoreCharacters"`
	Logger           *logrus.Entry
}

func (config *FloatParser) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if len(config.DecimalSeparator) == 0 {
		return errors.New("float.decimalSeparator is required")
	}

	return nil
}
