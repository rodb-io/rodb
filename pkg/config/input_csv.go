package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type CsvInputConfig struct{
	Source string
	Path string
	IgnoreFirstRow bool
	Delimiter rune
	Columns map[string]CsvInputColumnConfig
}

type CsvInputColumnConfig struct{
	Type string
	IgnoreCharacters string
	DecimalSeparator string
	TrueValues []string
	FalseValues []string
}

func (config *CsvInputConfig) validate(log *logrus.Logger) error {
	// The source and path will be validated at runtime
	if len(config.Columns) == 0 {
		return errors.New("A csv input must have at least one column")
	}

	if config.Delimiter == 0 {
		log.Debug("csv.delimiter not defined. Assuming ','")
		config.Delimiter = ','
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
