package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type JsonParser struct {
	Name   string `yaml:"name"`
	Logger *logrus.Entry
}

func (config *JsonParser) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("json.name is required")
	}

	return nil
}

func (config *JsonParser) Primitive() bool {
	return false
}
