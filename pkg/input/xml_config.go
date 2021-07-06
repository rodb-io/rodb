package input

import (
	"errors"
	"fmt"
	"github.com/antchfx/xpath"
	"github.com/sirupsen/logrus"
	"os"
)

type XmlInputPropertyType string

const (
	XmlInputPropertyTypePrimitive = XmlInputPropertyType("primitive")
	XmlInputPropertyTypeArray     = XmlInputPropertyType("array")
	XmlInputPropertyTypeObject    = XmlInputPropertyType("object")
)

type XmlConfig struct {
	Name             string              `yaml:"name"`
	Type             string              `yaml:"type"`
	Path             string              `yaml:"path"`
	DieOnInputChange *bool               `yaml:"dieOnInputChange"`
	Properties       []*XmlPropertyConfig `yaml:"properties"`
	RecordXPath      string              `yaml:"recordXpath"`
	Logger           *logrus.Entry
}

type XmlPropertyConfig struct {
	Name          string               `yaml:"name"`
	Type          XmlInputPropertyType `yaml:"type"`
	Parser        string               `yaml:"parser"`
	XPath         string               `yaml:"xpath"`
	Items         *XmlPropertyConfig    `yaml:"items"`
	Properties    []*XmlPropertyConfig  `yaml:"properties"`
	CompiledXPath *xpath.Expr
}

func (config *XmlConfig) GetName() string {
	return config.Name
}

func (config *XmlConfig) ShouldDieOnInputChange() bool {
	return config.DieOnInputChange == nil || *config.DieOnInputChange
}

func (config *XmlConfig) Validate(parsers map[string]Parser, log *logrus.Entry) error {
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

	alreadyExistingNames := make(map[string]bool)
	for propertyIndex, property := range config.Properties {
		logPrefix := fmt.Sprintf("xml.properties[%v].", propertyIndex)
		if err := property.Validate(parsers, true, log, logPrefix); err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}

		if _, exists := alreadyExistingNames[property.Name]; exists {
			return fmt.Errorf("Property names must be unique. Found '%v' twice.", property.Name)
		}
		alreadyExistingNames[property.Name] = true
	}

	return nil
}

func (config *XmlPropertyConfig) Validate(
	parsers map[string]Parser,
	nameRequired bool,
	log *logrus.Entry,
	logPrefix string,
) error {
	if nameRequired && config.Name == "" {
		return errors.New("name is required")
	}

	var err error
	config.CompiledXPath, err = xpath.Compile(config.XPath)
	if err != nil {
		return fmt.Errorf("xpath: Invalid xpath expression: %w", err)
	}

	if config.Type == "" {
		log.Debugf(logPrefix + "type is not set. Assuming 'primitive'.\n")
		config.Type = "primitive"
	}

	switch config.Type {
	case XmlInputPropertyTypePrimitive:
		if config.Parser == "" {
			log.Debug(logPrefix + "parser not defined. Assuming 'string'")
			config.Parser = "string"
		}

		_, parserExists := parsers[config.Parser]
		if !parserExists {
			return fmt.Errorf("parser: Parser '%v' not found in parsers list.", config.Parser)
		}

		if config.Items != nil {
			return fmt.Errorf("items can only be used on array properties.")
		}
		if config.Properties != nil && len(config.Properties) > 0 {
			return fmt.Errorf("properties can only be used on object properties.")
		}
	case XmlInputPropertyTypeArray:
		if config.Parser != "" {
			return fmt.Errorf("parser '%v' specified, but the property is not a primitive.", config.Parser)
		}

		if config.Properties != nil && len(config.Properties) > 0 {
			return fmt.Errorf("properties can only be used on object properties.")
		}

		if config.Items == nil {
			return fmt.Errorf("items is required for arrays.")
		}

		itemsLogPrefix := fmt.Sprintf("%vitems.", logPrefix)
		if err := config.Items.Validate(parsers, false, log, itemsLogPrefix); err != nil {
			return fmt.Errorf("items.%w", err)
		}
	case XmlInputPropertyTypeObject:
		if config.Parser != "" {
			return fmt.Errorf("parser '%v' specified, but the property is not a primitive.", config.Parser)
		}

		if config.Items != nil {
			return fmt.Errorf("items can only be used on array properties.")
		}

		if config.Properties == nil || len(config.Properties) == 0 {
			return errors.New("properties is required for objects.")
		}

		alreadyExistingNames := make(map[string]bool)
		for propertyIndex, property := range config.Properties {
			propertyLogPrefix := fmt.Sprintf("%vproperties[%v].", logPrefix, propertyIndex)
			if err := property.Validate(parsers, true, log, propertyLogPrefix); err != nil {
				return fmt.Errorf("properties[%v].%w", propertyIndex, err)
			}

			if _, exists := alreadyExistingNames[property.Name]; exists {
				return fmt.Errorf("Property names must be unique. Found '%v' twice.", property.Name)
			}
			alreadyExistingNames[property.Name] = true
		}
	default:
		return fmt.Errorf("type '%v' is invalid.", config.Type)
	}

	return nil
}
