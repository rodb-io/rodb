package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

type CsvInput struct {
	Source            string            `yaml:"source"`
	Path              string            `yaml:"path"`
	IgnoreFirstRow    bool              `yaml:"ignoreFirstRow"`
	Delimiter         string            `yaml:"delimiter"`
	Columns           []*CsvInputColumn `yaml:"columns"`
	ColumnIndexByName map[string]int
	Logger            *logrus.Entry
}

type CsvInputColumn struct {
	Name   string `yaml:"name"`
	Parser string `yaml:"parser"`
}

func (config *CsvInput) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	_, sourceExists := rootConfig.Sources[config.Source]
	if !sourceExists {
		return fmt.Errorf("csv.source: Source '%v' not found in sources list.", config.Source)
	}

	// The path will be validated at runtime

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
		err := column.validate(rootConfig, log)
		if err != nil {
			return fmt.Errorf("csv.columns[%v]: %w", columnIndex, err)
		}

		if _, exists := config.ColumnIndexByName[column.Name]; exists {
			return fmt.Errorf("Column names must be unique. Found '%v' twice.", column.Name)
		}
		config.ColumnIndexByName[column.Name] = columnIndex
	}

	return nil
}

func (config *CsvInputColumn) validate(rootConfig *Config, log *logrus.Entry) error {
	_, parserExists := rootConfig.Parsers[config.Parser]
	if !parserExists {
		return fmt.Errorf("Parser '%v' not found in parsers list.", config.Parser)
	}

	if config.Name == "" {
		return errors.New("name is required")
	}

	if config.Parser == "" {
		log.Debug("csv.columns[].parser not defined. Assuming 'string'")
		config.Parser = "string"
	}

	return nil
}
