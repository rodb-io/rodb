package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type Input struct {
	Csv *CsvInput `yaml:"csv"`
	Xml *XmlInput `yaml:"xml"`
}

func (config *Input) validate(rootConfig *Config, log *logrus.Entry) error {
	definedFields := 0
	if config.Csv != nil {
		definedFields++
		err := config.Csv.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}
	if config.Xml != nil {
		definedFields++
		err := config.Xml.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}

	if definedFields == 0 {
		return errors.New("All inputs must have a configuration")
	}
	if definedFields > 1 {
		return errors.New("An input can only have one configuration")
	}

	return nil
}

func (config *Input) Name() string {
	if config.Csv != nil {
		return config.Csv.Name
	}
	if config.Xml != nil {
		return config.Xml.Name
	}

	return ""
}

func (config *Input) PropertyParser(propertyName string) *string {
	if config.Csv != nil {
		return config.Csv.PropertyParser(propertyName)
	}
	if config.Xml != nil {
		return config.Xml.PropertyParser(propertyName)
	}

	return nil
}
