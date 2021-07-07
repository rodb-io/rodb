package parser

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type StringConfig struct {
	Name               string `yaml:"name"`
	Type               string `yaml:"type"`
	ConvertFromCharset string `yaml:"convertFromCharset"`
	Logger             *logrus.Entry
}

func (config *StringConfig) Validate(parsers map[string]Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("string.name is required")
	}

	// The ConvertFromCharset will be validated at runtime.
	// The default value is empty string (= don't convert)
	return nil
}

func (config *StringConfig) GetName() string {
	return config.Name
}

func (config *StringConfig) Primitive() bool {
	return true
}
