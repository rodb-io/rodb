package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type ParsedConfig struct {
	Parsers  []Parser
	Sources  []Source
	Inputs   []Input
	Indexes  []Index
	Services []Service
	Outputs  []Output
}

type Config struct {
	Parsers  map[string]Parser
	Sources  map[string]Source
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

	parsedConfig := &ParsedConfig{}
	err = yaml.UnmarshalStrict(yamlConfigWithEnv, parsedConfig)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse config file %v: %w", configPath, err)
	}

	config, err := NewConfigFromParsedConfig(parsedConfig)
	if err != nil {
		return nil, err
	}

	config.addDefaultConfigs(log)

	err = config.validate(config, log)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func NewConfigFromParsedConfig(parsedConfig *ParsedConfig) (*Config, error) {
	config := &Config{
		Parsers:  map[string]Parser{},
		Sources:  map[string]Source{},
		Inputs:   map[string]Input{},
		Indexes:  map[string]Index{},
		Services: map[string]Service{},
		Outputs:  map[string]Output{},
	}

	for _, parser := range parsedConfig.Parsers {
		name := parser.getName()
		if _, exists := config.Parsers[name]; exists {
			return nil, fmt.Errorf("Duplicate name '%v' for parser.", name)
		}
		config.Parsers[name] = parser
	}
	for _, source := range parsedConfig.Sources {
		name := source.getName()
		if _, exists := config.Sources[name]; exists {
			return nil, fmt.Errorf("Duplicate name '%v' for source.", name)
		}
		config.Sources[name] = source
	}
	for _, input := range parsedConfig.Inputs {
		name := input.getName()
		if _, exists := config.Inputs[name]; exists {
			return nil, fmt.Errorf("Duplicate name '%v' for input.", name)
		}
		config.Inputs[name] = input
	}
	for _, index := range parsedConfig.Indexes {
		name := index.getName()
		if _, exists := config.Indexes[name]; exists {
			return nil, fmt.Errorf("Duplicate name '%v' for index.", name)
		}
		config.Indexes[name] = index
	}
	for _, service := range parsedConfig.Services {
		name := service.getName()
		if _, exists := config.Services[name]; exists {
			return nil, fmt.Errorf("Duplicate name '%v' for service.", name)
		}
		config.Services[name] = service
	}
	for _, output := range parsedConfig.Outputs {
		name := output.getName()
		if _, exists := config.Outputs[name]; exists {
			return nil, fmt.Errorf("Duplicate name '%v' for output.", name)
		}
		config.Outputs[name] = output
	}

	return config, nil
}

func (config *Config) addDefaultConfigs(log *logrus.Logger) {
	if _, exists := config.Indexes["default"]; exists {
		log.Warnf("You have declared an index named 'default', which will replace the internally used one.\n")
	} else {
		config.Indexes["default"] = Index{
			Noop: &NoopIndex{
				Name: "default",
			},
		}
	}

	for parserName, parserConfig := range map[string]Parser{
		"string": {
			String: &StringParser{
				Name: "string",
			},
		},
		"integer": {
			Integer: &IntegerParser{
				Name:             "integer",
				IgnoreCharacters: "",
			},
		},
		"float": {
			Float: &FloatParser{
				Name:             "float",
				DecimalSeparator: ".",
				IgnoreCharacters: "",
			},
		},
		"boolean": {
			Boolean: &BooleanParser{
				Name:        "boolean",
				TrueValues:  []string{"true", "1", "TRUE"},
				FalseValues: []string{"false", "0", "FALSE"},
			},
		},
	} {
		if _, exists := config.Parsers[parserName]; exists {
			log.Warnf("You have declared a parser named '%v', which will replace the default one.\n", parserName)
		} else {
			config.Parsers[parserName] = parserConfig
		}
	}
}

func (config *Config) validate(rootConfig *Config, log *logrus.Logger) error {
	for subConfigName, subConfig := range config.Parsers {
		if err := subConfig.validate(rootConfig, log.WithField("object", "parsers."+subConfigName)); err != nil {
			return fmt.Errorf("parsers.%v: %w", subConfigName, err)
		}
	}

	for subConfigName, subConfig := range config.Sources {
		if err := subConfig.validate(rootConfig, log.WithField("object", "sources."+subConfigName)); err != nil {
			return fmt.Errorf("sources.%v: %w", subConfigName, err)
		}
	}

	for subConfigName, subConfig := range config.Inputs {
		if err := subConfig.validate(rootConfig, log.WithField("object", "inputs."+subConfigName)); err != nil {
			return fmt.Errorf("inputs.%v: %w", subConfigName, err)
		}
	}

	for subConfigName, subConfig := range config.Indexes {
		if err := subConfig.validate(rootConfig, log.WithField("object", "indexes."+subConfigName)); err != nil {
			return fmt.Errorf("indexes.%v: %w", subConfigName, err)
		}
	}

	for subConfigName, subConfig := range config.Services {
		if err := subConfig.validate(rootConfig, log.WithField("object", "services."+subConfigName)); err != nil {
			return fmt.Errorf("services.%v: %w", subConfigName, err)
		}
	}

	for subConfigName, subConfig := range config.Outputs {
		if err := subConfig.validate(rootConfig, log.WithField("object", "outputs."+subConfigName)); err != nil {
			return fmt.Errorf("outputs.%v: %w", subConfigName, err)
		}
	}

	err := checkDuplicateEndpointsPerService(config.Outputs)
	if err != nil {
		return err
	}

	return nil
}
