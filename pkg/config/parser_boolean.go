package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type BooleanParser struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	TrueValues  []string `yaml:"trueValues"`
	FalseValues []string `yaml:"falseValues"`
	Logger      *logrus.Entry
}

func (config *BooleanParser) Validate(parsers map[string]Parser, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("boolean.name is required")
	}

	if len(config.TrueValues) == 0 {
		return errors.New("boolean.trueValues is required")
	}
	if len(config.TrueValues) == 0 {
		return errors.New("boolean.falseValues is required")
	}

	return nil
}

func (config *BooleanParser) GetName() string {
	return config.Name
}

func (config *BooleanParser) Primitive() bool {
	return true
}
