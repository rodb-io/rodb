package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type Parser struct {
	Integer *IntegerParser `yaml:"integer"`
	Float   *FloatParser   `yaml:"float"`
	Boolean *BooleanParser `yaml:"boolean"`
	String  *StringParser  `yaml:"string"`
}

func (config *Parser) validate(rootConfig *Config, log *logrus.Logger) error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("One of your parsers does not have a definition.")
	}

	if len(fields) > 1 {
		return errors.New("One of your parsers has two different definitions.")
	}

	return fields[0].validate(rootConfig, log)
}
