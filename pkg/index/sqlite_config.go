package index

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/input"
)

type SqliteConfig struct {
	Name       string   `yaml:"name"`
	Type       string   `yaml:"type"`
	Dsn        string   `yaml:"dsn"`
	Input      string   `yaml:"input"`
	Properties []string `yaml:"properties"`
	Logger     *logrus.Entry
}

func (config *SqliteConfig) Validate(inputs map[string]input.Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("sqlite.name is required")
	}

	if config.Dsn == "" {
		return errors.New("sqlite.dsn is required")
	}
	// The DSN will be validated at runtime

	_, inputExists := inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("sqlite.input: Input '%v' not found in inputs list.", config.Input)
	}

	alreadyExistingProperties := make(map[string]bool)
	for _, propertyName := range config.Properties {
		if _, alreadyExists := alreadyExistingProperties[propertyName]; alreadyExists {
			return fmt.Errorf("sqlite.properties: Duplicate property '%v' in array.", propertyName)
		}
		alreadyExistingProperties[propertyName] = true
	}

	// The properties will be validated at runtime

	return nil
}

func (config *SqliteConfig) GetName() string {
	return config.Name
}

func (config *SqliteConfig) DoesHandleProperty(property string) bool {
	isHandled := false
	for _, handledProperty := range config.Properties {
		if property == handledProperty {
			isHandled = true
			break
		}
	}

	return isHandled
}

func (config *SqliteConfig) DoesHandleInput(input input.Config) bool {
	return input.GetName() == config.Input
}
