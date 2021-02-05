package config

import (
	"errors"
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

func (config *CsvInputConfig) validate() error {
	// The source and path will be validated at runtime
	if len(config.Columns) == 0 {
		return errors.New("A csv input must have at least one column")
	}

	if config.Delimiter == 0 {
		config.Delimiter = ','
	}

	for _, column := range config.Columns {
		err := column.validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (config *CsvInputColumnConfig) validate() error {
	if config.Type == "" {
		config.Type = "string"
	}

	if !isCsvInputColumnTypeValid(config.Type) {
		return errors.New("The csv column type '" + config.Type + "' is invalid")
	}

	if config.Type == "float" && len(config.DecimalSeparator) == 0 {
		return errors.New("You must define the decimalSeparator when using a float column")
	}

	if config.Type == "boolean" {
		if len(config.TrueValues) == 0 {
			return errors.New("You must define the trueValues when using a boolean column")
		}
		if len(config.TrueValues) == 0 {
			return errors.New("You must define the falseValues when using a boolean column")
		}
	}

	return nil
}
