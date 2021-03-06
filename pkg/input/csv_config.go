package input

import (
	"errors"
	"fmt"
	"github.com/rodb-io/rodb/pkg/parser"
	"github.com/sirupsen/logrus"
	"os"
)

type CsvConfig struct {
	Name              string             `yaml:"name"`
	Type              string             `yaml:"type"`
	Path              string             `yaml:"path"`
	DieOnInputChange  *bool              `yaml:"dieOnInputChange"`
	IgnoreFirstRow    bool               `yaml:"ignoreFirstRow"`
	AutodetectColumns bool               `yaml:"autodetectColumns"`
	Delimiter         string             `yaml:"delimiter"`
	Columns           []*CsvColumnConfig `yaml:"columns"`
	ColumnIndexByName map[string]int
	Logger            *logrus.Entry
}

type CsvColumnConfig struct {
	Name   string `yaml:"name"`
	Parser string `yaml:"parser"`
}

func (config *CsvConfig) GetName() string {
	return config.Name
}

func (config *CsvConfig) ShouldDieOnInputChange() bool {
	return config.DieOnInputChange == nil || *config.DieOnInputChange
}

func (config *CsvConfig) Validate(parsers map[string]parser.Config, log *logrus.Entry) error {
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
			config.Columns = make([]*CsvColumnConfig, 0)
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
		if err := column.Validate(parsers, log, logPrefix); err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}

		if _, exists := config.ColumnIndexByName[column.Name]; exists {
			return fmt.Errorf("Column names must be unique. Found '%v' twice.", column.Name)
		}
		config.ColumnIndexByName[column.Name] = columnIndex
	}

	return nil
}

func (config *CsvColumnConfig) Validate(parsers map[string]parser.Config, log *logrus.Entry, logPrefix string) error {
	if config.Name == "" {
		return errors.New("name is required")
	}

	if config.Parser == "" {
		log.Debug(logPrefix + "parser not defined. Assuming 'string'")
		config.Parser = "string"
	}

	_, parserExists := parsers[config.Parser]
	if !parserExists {
		return fmt.Errorf("parser: Parser '%v' not found in parsers list.", config.Parser)
	}

	return nil
}
