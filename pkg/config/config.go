package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type configParser struct {
	Parsers  []parserParser
	Inputs   []inputParser
	Indexes  []indexParser
	Services []serviceParser
	Outputs  []outputParser
}

type Config struct {
	Parsers  map[string]Parser
	Inputs   map[string]Input
	Indexes  map[string]Index
	Services map[string]Service
	Outputs  map[string]Output
}

func NewConfigFromYamlFile(configPath string, log *logrus.Logger) (*Config, error) {
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Cannot read config file %v: %w", configPath, err)
	}

	yamlConfigWithEnv := []byte(os.ExpandEnv(string(configData)))

	parsedConfig := &configParser{}
	if err := yaml.UnmarshalStrict(yamlConfigWithEnv, parsedConfig); err != nil {
		return nil, fmt.Errorf("Cannot parse config file %v: %w", configPath, err)
	}

	config, err := NewConfigFromParsedConfig(parsedConfig)
	if err != nil {
		return nil, err
	}

	config.addDefaultConfigs(log)

	if err := config.Validate(log); err != nil {
		return nil, err
	}

	return config, nil
}

func NewConfigFromParsedConfig(parsedConfig *configParser) (*Config, error) {
	config := &Config{
		Parsers:  map[string]Parser{},
		Inputs:   map[string]Input{},
		Indexes:  map[string]Index{},
		Services: map[string]Service{},
		Outputs:  map[string]Output{},
	}

	for _, parser := range parsedConfig.Parsers {
		name := parser.parser.GetName()
		if _, exists := config.Parsers[name]; exists {
			return nil, fmt.Errorf("Duplicate name '%v' for parser.", name)
		}
		config.Parsers[name] = parser.parser
	}
	for _, input := range parsedConfig.Inputs {
		name := input.input.GetName()
		if _, exists := config.Inputs[name]; exists {
			return nil, fmt.Errorf("Duplicate name '%v' for input.", name)
		}
		config.Inputs[name] = input.input
	}
	for _, index := range parsedConfig.Indexes {
		name := index.index.GetName()
		if _, exists := config.Indexes[name]; exists {
			return nil, fmt.Errorf("Duplicate name '%v' for index.", name)
		}
		config.Indexes[name] = index.index
	}
	for _, service := range parsedConfig.Services {
		name := service.service.GetName()
		if _, exists := config.Services[name]; exists {
			return nil, fmt.Errorf("Duplicate name '%v' for service.", name)
		}
		config.Services[name] = service.service
	}
	for _, output := range parsedConfig.Outputs {
		name := output.output.GetName()
		if _, exists := config.Outputs[name]; exists {
			return nil, fmt.Errorf("Duplicate name '%v' for output.", name)
		}
		config.Outputs[name] = output.output
	}

	return config, nil
}

func (config *Config) addDefaultConfigs(log *logrus.Logger) {
	defaultIndex := &NoopIndex{
		Name: "default",
	}
	if _, exists := config.Indexes[defaultIndex.GetName()]; exists {
		log.Warnf("You have declared an index named 'default', which will replace the internally used one.\n")
	} else {
		config.Indexes[defaultIndex.GetName()] = defaultIndex
	}

	for _, parserConfig := range []Parser{
		&StringParser{
			Name: "string",
		},
		&IntegerParser{
			Name:             "integer",
			IgnoreCharacters: "",
		},
		&FloatParser{
			Name:             "float",
			DecimalSeparator: ".",
			IgnoreCharacters: "",
		},
		&BooleanParser{
			Name:        "boolean",
			TrueValues:  []string{"true", "1", "TRUE"},
			FalseValues: []string{"false", "0", "FALSE"},
		},
		&JsonParser{
			Name: "json",
		},
	} {
		if _, exists := config.Parsers[parserConfig.GetName()]; exists {
			log.Warnf("You have declared a parser named '%v', which will replace the default one.\n", parserConfig.GetName())
		} else {
			config.Parsers[parserConfig.GetName()] = parserConfig
		}
	}
}

func (config *Config) Validate(log *logrus.Logger) error {
	for subConfigName, subConfig := range config.Parsers {
		if err := subConfig.Validate(config.Parsers, log.WithField("object", "parsers."+subConfigName)); err != nil {
			return fmt.Errorf("parsers.%v: %w", subConfigName, err)
		}
	}

	for subConfigName, subConfig := range config.Inputs {
		if err := subConfig.Validate(config.Parsers, log.WithField("object", "inputs."+subConfigName)); err != nil {
			return fmt.Errorf("inputs.%v: %w", subConfigName, err)
		}
	}

	for subConfigName, subConfig := range config.Indexes {
		if err := subConfig.Validate(config.Inputs, log.WithField("object", "indexes."+subConfigName)); err != nil {
			return fmt.Errorf("indexes.%v: %w", subConfigName, err)
		}
	}

	for subConfigName, subConfig := range config.Services {
		if err := subConfig.Validate(config.Outputs, log.WithField("object", "services."+subConfigName)); err != nil {
			return fmt.Errorf("services.%v: %w", subConfigName, err)
		}
	}

	for subConfigName, subConfig := range config.Outputs {
		if err := subConfig.Validate(config.Inputs, config.Indexes, config.Parsers, log.WithField("object", "outputs."+subConfigName)); err != nil {
			return fmt.Errorf("outputs.%v: %w", subConfigName, err)
		}
	}

	return nil
}
