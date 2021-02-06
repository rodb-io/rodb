package config

import (
	"fmt"
	"reflect"
	"os"
	"strings"
	yaml "gopkg.in/yaml.v2"
	"github.com/sirupsen/logrus"
)

type Config struct{
	Sources map[string]SourceConfig
	Inputs map[string]InputConfig
	Indexes map[string]IndexConfig
	Services map[string]ServiceConfig
	Outputs map[string]OutputConfig
}

func NewConfigFromYaml(yamlConfig []byte, log *logrus.Logger) (*Config, error) {
	yamlConfigWithEnv := []byte(os.ExpandEnv(string(yamlConfig)))

	config := &Config{}
	err := yaml.UnmarshalStrict(yamlConfigWithEnv, config)
	if err != nil {
		return nil, err
	}
	config.validate(log)

	return config, err
}

func (config *Config) validate(log *logrus.Logger) error {
	reflectConfig := reflect.ValueOf(config)
	for fieldIndex := 0; fieldIndex < reflectConfig.NumField(); fieldIndex++ {
		field := reflectConfig.Field(fieldIndex).Interface().(map[string]validable)
		fieldName := reflectConfig.Type().Field(fieldIndex).Name
		for categoryKey, categoryConfig := range field {
			err := categoryConfig.validate(log)
			if err != nil {
				return fmt.Errorf("%v.%v: %v", strings.ToLower(fieldName), categoryKey, err)
			}
		}
	}

	err := checkDuplicateEndpointsPerService(config.Outputs)
	if err != nil {
		return err
	}

	return nil
}
