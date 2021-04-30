package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

type CsvInput struct {
	Name              string            `yaml:"name"`
	Path              string            `yaml:"path"`
	DieOnInputChange  *bool             `yaml:"dieOnInputChange"`
	IgnoreFirstRow    bool              `yaml:"ignoreFirstRow"`
	AutodetectColumns bool              `yaml:"autodetectColumns"`
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

	if config.Name == "" {
		return errors.New("csv.name is required")
	}

	if config.DieOnInputChange == nil {
		defaultValue := true
		log.Debugf("csv.dieOnInputChange is not set. Assuming 'true'.\n")
		config.DieOnInputChange = &defaultValue
	}

	if config.AutodetectColumns {
		if !config.IgnoreFirstRow {
			log.Debugf("csv.autodetectColumns is enabled, but 'ignoreFirstRow' is not. The header row will be included in the data.\n")
		}

		if config.Columns == nil {
			config.Columns = make([]*CsvInputColumn, 0)
		}
		if len(config.Columns) != 0 {
			return errors.New("A csv input with 'autodetectColumns' set to 'true' must not define columns.")
		}
	} else {
		if len(config.Columns) == 0 {
			return errors.New("A csv input must have at least one column unless 'autodetectColumns' is set to 'true'")
		}
	}

	fileInfo, err := os.Stat(config.Path)
	if os.IsNotExist(err) {
		return errors.New("The csv file '" + config.Path + "' does not exist")
	}
	if fileInfo.IsDir() {
		return errors.New("The path '" + config.Path + "' is not a file")
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
		logPrefix := fmt.Sprintf("csv.columns[%v].", columnIndex)
		err := column.validate(rootConfig, log, logPrefix)
		if err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}

		if _, exists := config.ColumnIndexByName[column.Name]; exists {
			return fmt.Errorf("Column names must be unique. Found '%v' twice.", column.Name)
		}
		config.ColumnIndexByName[column.Name] = columnIndex
	}

	return nil
}

func (config *CsvInputColumn) validate(rootConfig *Config, log *logrus.Entry, logPrefix string) error {
	_, parserExists := rootConfig.Parsers[config.Parser]
	if !parserExists {
		return fmt.Errorf("parser: Parser '%v' not found in parsers list.", config.Parser)
	}

	if config.Name == "" {
		return errors.New("name is required")
	}

	if config.Parser == "" {
		log.Debug(logPrefix + "parser not defined. Assuming 'string'")
		config.Parser = "string"
	}

	return nil
}
