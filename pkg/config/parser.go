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
		config.parser = &IntegerParser{}
		return unmarshal(config.parser)
	case "float":
		config.parser = &FloatParser{}
		return unmarshal(config.parser)
	case "boolean":
		config.parser = &BooleanParser{}
		return unmarshal(config.parser)
	case "string":
		config.parser = &StringParser{}
		return unmarshal(config.parser)
	case "json":
		config.parser = &JsonParser{}
		return unmarshal(config.parser)
	case "split":
		config.parser = &SplitParser{}
		return unmarshal(config.parser)
	default:
		return fmt.Errorf("Error in parser config: Unknown type '%v'", objectType)
	}
}

type Parser interface {
	validate(parsers map[string]Parser, log *logrus.Entry) error
	GetName() string
	Primitive() bool
}
