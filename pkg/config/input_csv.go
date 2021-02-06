package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type CsvInputConfig struct{
	Source string `yaml:"source"`
	Path string `yaml:"path"`
	IgnoreFirstRow bool `yaml:"ignoreFirstRow"`
	Delimiter string `yaml:"delimiter"`
	Columns map[string]CsvInputColumnConfig `yaml:"columns"`
}

type CsvInputColumnConfig struct{
	Type string `yaml:"type"`
	IgnoreCharacters string `yaml:"ignoreCharacters"`
	DecimalSeparator string `yaml:"decimalSeparator"`
	TrueValues []string `yaml:"trueValues"`
	FalseValues []string `yaml:"falseValues"`
}

func (config *CsvInputConfig) validate(log *logrus.Logger) error {
	// The source and path will be validated at runtime
	if len(config.Columns) == 0 {
		return errors.New("A csv input must have at least one column")
	}

	if config.Delimiter == "" {
		log.Debug("csv.delimiter not defined. Assuming ','")
		config.Delimiter = ","
	}

	if len(config.Delimiter) > 1 {
		return errors.New("csv.delimiter must be a single character")
	}

	for _, column := range config.Columns {
		err := column.validate(log)
		if err != nil {
			return err
		}
	}

	return nil
}

func (config *CsvInputColumnConfig) validate(log *logrus.Logger) error {
	if config.Type == "" {
		log.Warn("csv.columns[].type not defined. Assuming 'string'")
		config.Type = "string"
	}

	if !isCsvInputColumnTypeValid(config.Type) {
		return errors.New("csv.columns[].type = '" + config.Type + "' is invalid")
	}

	if config.Type == "float" && len(config.DecimalSeparator) == 0 {
		return errors.New("csv.columns[].decimalSeparator is required when type = 'float'")
	}

	if config.Type == "boolean" {
		if len(config.TrueValues) == 0 {
			return errors.New("csv.columns[].trueValues is required when type = 'boolean'")
		}
		if len(config.TrueValues) == 0 {
			return errors.New("csv.columns[].falseValues is required when type = 'boolean'")
		}
	}

	return nil
}
