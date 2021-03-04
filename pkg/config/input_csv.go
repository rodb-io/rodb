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
	Name   string   `yaml:"name"`
	Parser string `yaml:"parser"`
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
	// The parser will be validated at runtime
	if config.Name == "" {
		return errors.New("csv.columns[].name is required")
	}

	if config.Parser == "" {
		log.Debug("csv.columns[].parser not defined. Assuming 'string'")
		config.Parser = "string"
	}

	return nil
}
