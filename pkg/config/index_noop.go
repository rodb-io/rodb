package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type NoopIndex struct {
	Name   string `yaml:"name"`
	Logger *logrus.Entry
}

func (config *NoopIndex) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("noop.name is required")
	}

	return nil
}

func (config *NoopIndex) DoesHandleProperty(property string) bool {
	return true
}

func (config *NoopIndex) DoesHandleInput(input Input) bool {
	return true
}
