package config

import (
	"github.com/sirupsen/logrus"
)

type IntegerParser struct {
	IgnoreCharacters string `yaml:"ignoreCharacters"`
}

func (config *IntegerParser) validate(log *logrus.Logger) error {
	return nil
}
