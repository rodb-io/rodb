package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/util"
)

type inputParser struct {
	input Input
}

func (config *inputParser) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in input config: %w", err)
	}

	switch objectType {
	case "csv":
		config.input = &CsvInput{}
		return unmarshal(config.input)
	case "xml":
		config.input = &XmlInput{}
		return unmarshal(config.input)
	case "json":
		config.input = &JsonInput{}
		return unmarshal(config.input)
	default:
		return fmt.Errorf("Error in input config: Unknown type '%v'", objectType)
	}
}

type Input interface {
	validate(rootConfig *Config, log *logrus.Entry) error
	GetName() string
	ShouldDieOnInputChange() bool
}
