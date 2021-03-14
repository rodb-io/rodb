package config

import (
	"github.com/sirupsen/logrus"
)

type IntegerParser struct {
	IgnoreCharacters string `yaml:"ignoreCharacters"`
	Logger           *logrus.Entry
}

func (config *IntegerParser) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	return nil
}
