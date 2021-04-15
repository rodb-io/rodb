package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type Sort struct {
	Logger    *logrus.Entry
	Column    string `yaml:"column"`
	Ascending *bool  `yaml:"ascending"`
}

func (config *Sort) validate(
	rootConfig *Config,
	input Input,
	log *logrus.Entry,
	logPrefix string,
) error {
	config.Logger = log

	parserName := input.ColumnParser(config.Column)
	if parserName == nil {
		return fmt.Errorf("column: Could not find the associated parser.")
	}
	parser, parserExists := rootConfig.Parsers[*parserName]
	if !parserExists {
		return fmt.Errorf("column: The associated parser '%v' does not exist.", *parserName)
	}
	if !parser.Primitive() {
		return fmt.Errorf("column: Cannot be used to sort because it does not have a primitive type.")
	}

	if config.Ascending == nil {
		log.Debugf(logPrefix + "ascending is not set. Assuming 'true'.\n")
		defaultAscending := true
		config.Ascending = &defaultAscending
	}

	return nil
}
