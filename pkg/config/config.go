package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

type Config struct {
	Parsers  map[string]Parser
	Sources  map[string]Source
	Inputs   map[string]Input
	Indexes  map[string]Index
	Services map[string]Service
	Outputs  map[string]Output
}

func NewConfigFromYaml(yamlConfig []byte, log *logrus.Logger) (*Config, error) {
	yamlConfigWithEnv := []byte(os.ExpandEnv(string(yamlConfig)))

	config := &Config{}
	err := yaml.UnmarshalStrict(yamlConfigWithEnv, config)
	if err != nil {
		return nil, err
	}

	err = config.validate(config, log)
	if err != nil {
		return nil, err
	}

	return config, err
}

func NewConfigFromYamlFile(configPath string, log *logrus.Logger) (*Config, error) {
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Cannot read config file %v: %w", configPath, err)
	}

	config, err := NewConfigFromYaml(configData, log)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse config file %v: %w", configPath, err)
	}

	return config, nil
}

func (config *Config) validate(rootConfig *Config, log *logrus.Logger) error {
	for subConfigName, subConfig := range config.Parsers {
		if err := subConfig.validate(rootConfig, log); err != nil {
			return fmt.Errorf("parsers.%v: %w", strings.ToLower(subConfigName), err)
		}
	}

	for subConfigName, subConfig := range config.Sources {
		if err := subConfig.validate(rootConfig, log); err != nil {
			return fmt.Errorf("sources.%v: %w", strings.ToLower(subConfigName), err)
		}
	}

	for subConfigName, subConfig := range config.Inputs {
		if err := subConfig.validate(rootConfig, log); err != nil {
			return fmt.Errorf("inputs.%v: %w", strings.ToLower(subConfigName), err)
		}
	}

	for subConfigName, subConfig := range config.Indexes {
		if err := subConfig.validate(rootConfig, log); err != nil {
			return fmt.Errorf("indexes.%v: %w", strings.ToLower(subConfigName), err)
		}
	}

	for subConfigName, subConfig := range config.Services {
		if err := subConfig.validate(rootConfig, log); err != nil {
			return fmt.Errorf("services.%v: %w", strings.ToLower(subConfigName), err)
		}
	}

	for subConfigName, subConfig := range config.Outputs {
		if err := subConfig.validate(rootConfig, log); err != nil {
			return fmt.Errorf("outputs.%v: %w", strings.ToLower(subConfigName), err)
		}
	}

	err := checkDuplicateEndpointsPerService(config.Outputs)
	if err != nil {
		return err
	}

	return nil
}
