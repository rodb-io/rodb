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

func (config *Parser) validate(rootConfig *Config, log *logrus.Entry) error {
	definedFields := 0
	if config.Integer != nil {
		definedFields++
		err := config.Integer.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}
	if config.Float != nil {
		definedFields++
		err := config.Float.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}
	if config.Boolean != nil {
		definedFields++
		err := config.Boolean.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}
	if config.String != nil {
		definedFields++
		err := config.String.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}

	if definedFields == 0 {
		return errors.New("One of your parsers does not have a definition.")
	}
	if definedFields > 1 {
		return errors.New("One of your parsers has two different definitions.")
	}

	return nil
}

func (config *Parser) Name() string {
	if config.Integer != nil {
		return config.Integer.Name
	}
	if config.Float != nil {
		return config.Float.Name
	}
	if config.Boolean != nil {
		return config.Boolean.Name
	}
	if config.String != nil {
		return config.String.Name
	}

	return ""
}
