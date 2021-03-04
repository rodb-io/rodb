package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rods/pkg/types"
)

type CsvInput struct {
	Source            string           `yaml:"source"`
	Path              string           `yaml:"path"`
	IgnoreFirstRow    bool             `yaml:"ignoreFirstRow"`
	Delimiter         string           `yaml:"delimiter"`
	Columns           []CsvInputColumn `yaml:"columns"`
	ColumnIndexByName map[string]int
}

type CsvInputColumn struct {
	Name             string   `yaml:"name"`
	Type             string `yaml:"type"`
	typeDefinition   types.Type
	IgnoreCharacters string   `yaml:"ignoreCharacters"`
	DecimalSeparator string   `yaml:"decimalSeparator"`
	TrueValues       []string `yaml:"trueValues"`
	FalseValues      []string `yaml:"falseValues"`
}

func (config *CsvInput) validate(log *logrus.Logger) error {
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

	config.ColumnIndexByName = make(map[string]int)
	for columnIndex, column := range config.Columns {
		err := column.validate(log)
		if err != nil {
			return err
		}

		if _, exists := config.ColumnIndexByName[column.Name]; exists {
			return fmt.Errorf("Column names must be unique. Found '%v' twice.", column.Name)
		}
		config.ColumnIndexByName[column.Name] = columnIndex
	}

	return nil
}

func (config *CsvInputColumn) validate(log *logrus.Logger) error {
	if config.Name == "" {
		return errors.New("csv.columns[].name is required")
	}

	if config.Type == "" {
		log.Warn("csv.columns[].type not defined. Assuming 'string'")
		config.Type = "string"
	}

	typeDefinition, err := types.FromString(config.Type)
	if err != nil {
		return fmt.Errorf("csv.columns[].type: '%w'", err)
	}
	config.typeDefinition = typeDefinition

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

func (config *CsvInputColumn) TypeDefinition() types.Type {
	return config.TypeDefinition()
}
