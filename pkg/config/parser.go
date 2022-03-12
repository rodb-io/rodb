package config

import (
	"fmt"
	"github.com/rodb-io/rodb/pkg/parser"
	"github.com/rodb-io/rodb/pkg/util"
)

type parserParser struct {
	parser parser.Config
}

func (config *parserParser) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in parser config: %w", err)
	}

	switch objectType {
	case "integer":
		config.parser = &parser.IntegerConfig{}
		return unmarshal(config.parser)
	case "float":
		config.parser = &parser.FloatConfig{}
		return unmarshal(config.parser)
	case "boolean":
		config.parser = &parser.BooleanConfig{}
		return unmarshal(config.parser)
	case "string":
		config.parser = &parser.StringConfig{}
		return unmarshal(config.parser)
	case "json":
		config.parser = &parser.JsonConfig{}
		return unmarshal(config.parser)
	case "split":
		config.parser = &parser.SplitConfig{}
		return unmarshal(config.parser)
	default:
		return fmt.Errorf("Error in parser config: Unknown type '%v'", objectType)
	}
}
