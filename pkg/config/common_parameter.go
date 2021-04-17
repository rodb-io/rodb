package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

type Parameter struct {
	Property string `yaml:"property"`
	Index    string `yaml:"index"`
	Parser   string `yaml:"parser"`
}

func (config *Parameter) validate(
	rootConfig *Config,
	log *logrus.Entry,
	logPrefix string,
	input Input,
) error {
	if config.Property == "" {
		return errors.New("property is empty")
	}

	if config.Index == "" {
		log.Debugf(logPrefix + "index is empty. Assuming 'default'.\n")
		config.Index = "default"
	}
	index, indexExists := rootConfig.Indexes[config.Index]
	if !indexExists {
		return fmt.Errorf("index: Index '%v' not found in indexes list.", config.Index)
	}
	if !index.DoesHandleInput(input) {
		return fmt.Errorf("index: Index '%v' does not handle input '%v'.", config.Index, input.Name())
	}
	if !index.DoesHandleProperty(config.Property) {
		return fmt.Errorf("property: Index '%v' does not handle property '%v'.", config.Index, config.Property)
	}

	if config.Parser == "" {
		log.Debug(logPrefix + "parser not defined. Assuming 'string'")
		config.Parser = "string"
	}
	parser, parserExists := rootConfig.Parsers[config.Parser]
	if !parserExists {
		return fmt.Errorf("parser: Parser '%v' not found in parsers list.", config.Parser)
	}

	if !parser.Primitive() {
		return fmt.Errorf("parser: The parser '%v' is not a primitive type and cannot be used as a parameter.", config.Parser)
	}

	return nil
}
