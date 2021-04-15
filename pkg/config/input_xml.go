package config

import (
	"errors"
	"fmt"
	"github.com/antchfx/xpath"
	"github.com/sirupsen/logrus"
	"os"
)

type XmlInput struct {
	Name              string            `yaml:"name"`
	Path              string            `yaml:"path"`
	DieOnInputChange  *bool             `yaml:"dieOnInputChange"`
	Columns           []*XmlInputColumn `yaml:"columns"`
	RecordXPath       string            `yaml:"recordXpath"`
	ColumnIndexByName map[string]int
	Logger            *logrus.Entry
}

type XmlInputColumn struct {
	Name          string `yaml:"name"`
	Parser        string `yaml:"parser"`
	XPath         string `yaml:"xpath"`
	CompiledXPath *xpath.Expr
}

func (config *XmlInput) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("xml.name is required")
	}

	if config.DieOnInputChange == nil {
		defaultValue := true
		log.Debugf("xml.dieOnInputChange is not set. Assuming 'true'.\n")
		config.DieOnInputChange = &defaultValue
	}

	_, err := xpath.Compile(config.RecordXPath)
	if err != nil {
		return fmt.Errorf("recordXpath: Invalid xpath expression: %w", err)
	}

	if len(config.Columns) == 0 {
		return errors.New("An xml input must have at least one column")
	}

	fileInfo, err := os.Stat(config.Path)
	if os.IsNotExist(err) {
		return errors.New("The xml file '" + config.Path + "' does not exist")
	}
	if fileInfo.IsDir() {
		return errors.New("The path '" + config.Path + "' is not a file")
	}

	config.ColumnIndexByName = make(map[string]int)
	for columnIndex, column := range config.Columns {
		err := column.validate(rootConfig, log)
		if err != nil {
			return fmt.Errorf("xml.columns[%v]: %w", columnIndex, err)
		}

		if _, exists := config.ColumnIndexByName[column.Name]; exists {
			return fmt.Errorf("Column names must be unique. Found '%v' twice.", column.Name)
		}
		config.ColumnIndexByName[column.Name] = columnIndex
	}

	return nil
}

func (config *XmlInput) ColumnParser(columnName string) *string {
	for _, column := range config.Columns {
		if column.Name == columnName {
			return &column.Parser
		}
	}

	return nil
}

func (config *XmlInputColumn) validate(rootConfig *Config, log *logrus.Entry) error {
	_, parserExists := rootConfig.Parsers[config.Parser]
	if !parserExists {
		return fmt.Errorf("Parser '%v' not found in parsers list.", config.Parser)
	}

	if config.Name == "" {
		return errors.New("name is required")
	}

	var err error
	config.CompiledXPath, err = xpath.Compile(config.XPath)
	if err != nil {
		return fmt.Errorf("xpath: Invalid xpath expression: %w", err)
	}

	if config.Parser == "" {
		log.Debug("xml.columns[].parser not defined. Assuming 'string'")
		config.Parser = "string"
	}

	return nil
}
