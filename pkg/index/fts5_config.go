package index

import (
	"errors"
	"fmt"
	"github.com/rodb-io/rodb/pkg/input"
	"github.com/sirupsen/logrus"
)

var Fts5PropertyCollateAllowedValues = []string{"BINARY", "NOCASE", "RTIRM"}

type Fts5Config struct {
	Name       string   `yaml:"name"`
	Type       string   `yaml:"type"`
	Dsn        string   `yaml:"dsn"`
	Input      string   `yaml:"input"`
	Prefix     []int    `yaml:"prefix"`
	Tokenize   string   `yaml:"tokenize"`
	Properties []string `yaml:"properties"`
	Logger     *logrus.Entry
}

func (config *Fts5Config) Validate(inputs map[string]input.Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("sqlite.name is required")
	}

	if config.Dsn == "" {
		return errors.New("sqlite.dsn is required")
	}
	// The DSN will be validated at runtime

	if config.Prefix == nil {
		config.Prefix = make([]int, 0)
	}

	if config.Tokenize == "" {
		config.Tokenize = "unicode61"
		log.Debug("sqlite.tokenize is not set. Assuming 'unicode61'")
	}

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

func (config *Fts5Config) GetName() string {
	return config.Name
}

func (config *Fts5Config) DoesHandleProperty(propertyName string) bool {
	return propertyName == "match"
}

func (config *Fts5Config) DoesHandleInput(input input.Config) bool {
	return input.GetName() == config.Input
}
