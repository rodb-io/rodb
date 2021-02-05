package config

import (
	"reflect"
	yaml "gopkg.in/yaml.v2"
)

type Config struct{
	Sources map[string]SourceConfig
	Inputs map[string]InputConfig
	Indexes map[string]IndexConfig
	Services map[string]ServiceConfig
	Outputs map[string]OutputConfig
}

func NewConfigFromYaml(yamlConfig []byte) (*Config, error) {
	config := &Config{}
	err := yaml.UnmarshalStrict(yamlConfig, config)
	if err != nil {
		return nil, err
	}

	config.validate()

	return config, err
}

func (config *Config) validate() error {
	reflectConfig := reflect.ValueOf(config)
	for fieldIndex := 0; fieldIndex < reflectConfig.NumField(); fieldIndex++ {
		field := reflectConfig.Field(fieldIndex).Interface().(map[string]validable)
		for _, categoryConfig := range field {
			err := categoryConfig.validate()
			if err != nil {
				return err
			}
		}
	}

	err := checkDuplicateEndpointsPerService(config.Outputs)
	if err != nil {
		return err
	}

	return nil
}

// TODO unit test for utils.go
// TODO when setting a default value, warn in the console
