package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

type SplitParser struct {
	Name      string  `yaml:"name"`
	Delimiter *string `yaml:"delimiter"`
	Parser    string  `yaml:"parser"`
	Logger    *logrus.Entry
}

func (config *SplitParser) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("split.name is required")
	}

	if config.Delimiter == nil {
		return errors.New("split.delimiter is required")
	}

	_, parserExists := rootConfig.Parsers[config.Parser]
	if !parserExists {
		return fmt.Errorf("Parser '%v' not found in parsers list.", config.Parser)
	}

	return nil
}

func (config *SplitParser) Primitive() bool {
	return false
}
