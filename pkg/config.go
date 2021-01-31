package pkg

import (
	yaml "gopkg.in/yaml.v2"
)

type Config struct{
	Test string
}

func NewConfigFromYaml(yamlConfig []byte) (Config, error) {
	config := Config{}
	err := yaml.UnmarshalStrict(yamlConfig, &config)
	return config, err
}
