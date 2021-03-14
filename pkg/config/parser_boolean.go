package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type BooleanParser struct {
	TrueValues  []string `yaml:"trueValues"`
	FalseValues []string `yaml:"falseValues"`
	Logger      *logrus.Entry
}

func (config *BooleanParser) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if len(config.TrueValues) == 0 {
		return errors.New("boolean.trueValues is required")
	}
	if len(config.TrueValues) == 0 {
		return errors.New("boolean.falseValues is required")
	}

	return nil
}
