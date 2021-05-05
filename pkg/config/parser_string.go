package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type StringParser struct {
	Name               string `yaml:"name"`
	Type               string `yaml:"type"`
	ConvertFromCharset string `yaml:"convertFromCharset"`
	Logger             *logrus.Entry
}

func (config *StringParser) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("string.name is required")
	}

	// The ConvertFromCharset will be validated at runtime.
	// The default value is empty string (= don't convert)
	return nil
}

func (config *StringParser) GetName() string {
	return config.Name
}

func (config *StringParser) Primitive() bool {
	return true
}
