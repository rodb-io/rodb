package config

import (
	"github.com/sirupsen/logrus"
)

type StringParser struct {
	ConvertFromCharset string `yaml:"convertFromCharset"`
	Logger             *logrus.Entry
}

func (config *StringParser) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	// The ConvertFromCharset will be validated at runtime.
	// The default value is empty string (= don't convert)
	return nil
}
