package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type FloatParser struct {
	DecimalSeparator string `yaml:"decimalSeparator"`
	IgnoreCharacters string `yaml:"ignoreCharacters"`
}

func (config *FloatParser) validate(log *logrus.Logger) error {
	if len(config.DecimalSeparator) == 0 {
		return errors.New("float.decimalSeparator is required")
	}

	return nil
}
