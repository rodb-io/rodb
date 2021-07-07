package config

import (
	"fmt"
	"rodb.io/pkg/util"
	"rodb.io/pkg/input"
)

type inputParser struct {
	input input.Config
}

func (config *inputParser) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in input config: %w", err)
	}

	switch objectType {
	case "csv":
		config.input = &input.CsvConfig{}
		return unmarshal(config.input)
	case "xml":
		config.input = &input.XmlConfig{}
		return unmarshal(config.input)
	case "json":
		config.input = &input.JsonConfig{}
		return unmarshal(config.input)
	default:
		return fmt.Errorf("Error in input config: Unknown type '%v'", objectType)
	}
}
