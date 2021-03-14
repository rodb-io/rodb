package config

import (
	"github.com/sirupsen/logrus"
)

type IntegerParser struct {
	IgnoreCharacters string `yaml:"ignoreCharacters"`
}

func (config *IntegerParser) validate(rootConfig *Config, log *logrus.Entry) error {
	return nil
}
