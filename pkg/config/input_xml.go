package config

import (
	"errors"
	"fmt"
	"github.com/antchfx/xpath"
	"github.com/sirupsen/logrus"
	"os"
)

type XmlInput struct {
	Name                string              `yaml:"name"`
	Path                string              `yaml:"path"`
	DieOnInputChange    *bool               `yaml:"dieOnInputChange"`
	Properties          []*XmlInputProperty `yaml:"properties"`
	RecordXPath         string              `yaml:"recordXpath"`
	PropertyIndexByName map[string]int
	Logger              *logrus.Entry
}

type XmlInputProperty struct {
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

	if len(config.Properties) == 0 {
		return errors.New("An xml input must have at least one property")
	}

	fileInfo, err := os.Stat(config.Path)
	if os.IsNotExist(err) {
		return errors.New("The xml file '" + config.Path + "' does not exist")
	}
	if fileInfo.IsDir() {
		return errors.New("The path '" + config.Path + "' is not a file")
	}

	config.PropertyIndexByName = make(map[string]int)
	for propertyIndex, property := range config.Properties {
		err := property.validate(rootConfig, log)
		if err != nil {
			return fmt.Errorf("xml.properties[%v]: %w", propertyIndex, err)
		}

		if _, exists := config.PropertyIndexByName[property.Name]; exists {
			return fmt.Errorf("Property names must be unique. Found '%v' twice.", property.Name)
		}
		config.PropertyIndexByName[property.Name] = propertyIndex
	}

	return nil
}

func (config *XmlInput) PropertyParser(propertyName string) *string {
	for _, property := range config.Properties {
		if property.Name == propertyName {
			return &property.Parser
		}
	}

	return nil
}

func (config *XmlInputProperty) validate(rootConfig *Config, log *logrus.Entry) error {
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
		log.Debug("xml.properties[].parser not defined. Assuming 'string'")
		config.Parser = "string"
	}

	return nil
}
