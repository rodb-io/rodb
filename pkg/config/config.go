package config

import (
	"fmt"
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
	for subConfigName, subConfig := range config.Sources {
		if err := subConfig.validate(log); err != nil {
			return fmt.Errorf("sources.%v: %v", strings.ToLower(subConfigName), err)
		}
	}

	for subConfigName, subConfig := range config.Inputs {
		if err := subConfig.validate(log); err != nil {
			return fmt.Errorf("inputs.%v: %v", strings.ToLower(subConfigName), err)
		}
	}

	for subConfigName, subConfig := range config.Indexes {
		if err := subConfig.validate(log); err != nil {
			return fmt.Errorf("indexes.%v: %v", strings.ToLower(subConfigName), err)
		}
	}

	for subConfigName, subConfig := range config.Services {
		if err := subConfig.validate(log); err != nil {
			return fmt.Errorf("services.%v: %v", strings.ToLower(subConfigName), err)
		}
	}

	for subConfigName, subConfig := range config.Outputs {
		if err := subConfig.validate(log); err != nil {
			return fmt.Errorf("outputs.%v: %v", strings.ToLower(subConfigName), err)
		}
	}

	err := checkDuplicateEndpointsPerService(config.Outputs)
	if err != nil {
		return err
	}

	return nil
}
