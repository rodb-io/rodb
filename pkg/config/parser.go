package config

import (
	"fmt"
	"rodb.io/pkg/util"
	"rodb.io/pkg/parser"
)

type parserParser struct {
	parser parser.Config
}

func (config *parserParser) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in parser config: %w", err)
	}

	config.parser, err = parser.NewConfigFromType(objectType)
	if err != nil {
		return fmt.Errorf("Error in parser config: %w", err)
	}

	return unmarshal(config.parser)
}
