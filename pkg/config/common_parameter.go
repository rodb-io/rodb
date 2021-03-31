package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

type Parameter struct {
	Column string `yaml:"column"`
	Index  string `yaml:"index"`
	Parser string `yaml:"parser"`
}

func (config *Parameter) validate(
	rootConfig *Config,
	log *logrus.Entry,
	logPrefix string,
	input Input,
) error {
	if config.Column == "" {
		return errors.New("column is empty")
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
	if !index.DoesHandleColumn(config.Column) {
		return fmt.Errorf("column: Index '%v' does not handle column '%v'.", config.Index, config.Column)
	}

	if config.Parser == "" {
		log.Debug(logPrefix + "parser not defined. Assuming 'string'")
		config.Parser = "string"
	}
	_, parserExists := rootConfig.Parsers[config.Parser]
	if !parserExists {
		return fmt.Errorf("parser: Parser '%v' not found in parsers list.", config.Parser)
	}

	return nil
}
