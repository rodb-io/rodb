package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type Sort struct {
	Logger    *logrus.Entry
	Property  string `yaml:"property"`
	Ascending *bool  `yaml:"ascending"`
}

func (config *Sort) validate(
	rootConfig *Config,
	input Input,
	log *logrus.Entry,
	logPrefix string,
) error {
	config.Logger = log

	parserName, err := input.PropertyParser(config.Property)
	if err != nil {
		return fmt.Errorf("property: %w", err)
	}
	if parserName != nil {
		parser, parserExists := rootConfig.Parsers[*parserName]
		if !parserExists {
			return fmt.Errorf("property: The associated parser '%v' does not exist.", *parserName)
		}
		if !parser.Primitive() {
			return fmt.Errorf("property: Cannot be used to sort because it does not have a primitive type.")
		}
	}

	if config.Ascending == nil {
		log.Debugf(logPrefix + "ascending is not set. Assuming 'true'.\n")
		defaultAscending := true
		config.Ascending = &defaultAscending
	}

	return nil
}
