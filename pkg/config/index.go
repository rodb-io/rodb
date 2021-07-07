package config

import (
	"fmt"
	"rodb.io/pkg/index"
	"rodb.io/pkg/util"
)

type indexParser struct {
	index index.Config
}

func (config *indexParser) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in index config: %w", err)
	}

	switch objectType {
	case "map":
		config.index = &index.MapConfig{}
		return unmarshal(config.index)
	case "wildcard":
		config.index = &index.WildcardConfig{}
		return unmarshal(config.index)
	case "noop":
		config.index = &index.NoopConfig{}
		return unmarshal(config.index)
	default:
		return fmt.Errorf("Error in index config: Unknown type '%v'", objectType)
	}
}
