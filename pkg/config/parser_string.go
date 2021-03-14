package config

import (
	"github.com/sirupsen/logrus"
)

type StringParser struct {
	ConvertFromCharset string `yaml:"convertFromCharset"`
}

func (config *StringParser) validate(rootConfig *Config, log *logrus.Entry) error {
	// The ConvertFromCharset will be validated at runtime.
	// The default value is empty string (= don't convert)
	return nil
}
