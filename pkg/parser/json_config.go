package parser

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type JsonConfig struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Logger *logrus.Entry
}

func (config *JsonConfig) Validate(parsers map[string]Parser, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("json.name is required")
	}

	return nil
}

func (config *JsonConfig) GetName() string {
	return config.Name
}

func (config *JsonConfig) Primitive() bool {
	return false
}
