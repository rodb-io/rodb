package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/util"
)

type parserParser struct {
	parser Parser
}

func (config *parserParser) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in parser config: %w", err)
	}

	switch objectType {
	case "integer":
		config.parser = &IntegerConfig{}
		return unmarshal(config.parser)
	case "float":
		config.parser = &FloatConfig{}
		return unmarshal(config.parser)
	case "boolean":
		config.parser = &BooleanConfig{}
		return unmarshal(config.parser)
	case "string":
		config.parser = &StringConfig{}
		return unmarshal(config.parser)
	case "json":
		config.parser = &JsonConfig{}
		return unmarshal(config.parser)
	case "split":
		config.parser = &SplitConfig{}
		return unmarshal(config.parser)
	default:
		return fmt.Errorf("Error in parser config: Unknown type '%v'", objectType)
	}
}

type Parser interface {
	Validate(parsers map[string]Parser, log *logrus.Entry) error
	GetName() string
	Primitive() bool
}
