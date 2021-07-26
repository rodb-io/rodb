package index

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/input"
	"rodb.io/pkg/util"
	"strings"
)

var SqlitePropertyCollateAllowedValues = []string{"BINARY", "NOCASE", "RTIRM"}

type SqliteConfig struct {
	Name       string                 `yaml:"name"`
	Type       string                 `yaml:"type"`
	Dsn        string                 `yaml:"dsn"`
	Input      string                 `yaml:"input"`
	Properties []SqlitePropertyConfig `yaml:"properties"`
	Logger     *logrus.Entry
}

type SqlitePropertyConfig struct {
	Name    string `yaml:"name"`
	Collate string `yaml:"collate"`
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
	for propertyIndex, property := range config.Properties {
		logPrefix := fmt.Sprintf("sqlite.properties[%v].", propertyIndex)
		if err := property.Validate(log, logPrefix); err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}

		if _, alreadyExists := alreadyExistingProperties[property.Name]; alreadyExists {
			return fmt.Errorf("sqlite.properties: Duplicate property '%v' in array.", property.Name)
		}
		alreadyExistingProperties[property.Name] = true
	}

	// The properties will be validated at runtime

	return nil
}

func (config *SqliteConfig) GetName() string {
	return config.Name
}

func (config *SqliteConfig) DoesHandleProperty(propertyName string) bool {
	isHandled := false
	for _, handledProperty := range config.Properties {
		if propertyName == handledProperty.Name {
			isHandled = true
			break
		}
	}

	return isHandled
}

func (config *SqliteConfig) DoesHandleInput(input input.Config) bool {
	return input.GetName() == config.Input
}

func (config *SqlitePropertyConfig) Validate(log *logrus.Entry, logPrefix string) error {
	if config.Name == "" {
		return errors.New("name is required")
	}

	if config.Collate == "" {
		log.Debug(logPrefix + "collate not defined. Assuming 'binary'")
		config.Collate = SqlitePropertyCollateAllowedValues[0]
	} else {
		config.Collate = strings.ToUpper(config.Collate)
	}

	if !util.IsInArray(config.Collate, SqlitePropertyCollateAllowedValues) {
		allowedValues := `"` + strings.Join(SqlitePropertyCollateAllowedValues, `", "`) + `"`
		return fmt.Errorf("collate value of '%v' is invalid. Allowed values: %v", config.Collate, allowedValues)
	}

	return nil
}
