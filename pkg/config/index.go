package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/util"
)

type indexParser struct {
	index Index
}

func (config *indexParser) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in index config: %w", err)
	}

	switch objectType {
	case "map":
		config.index = &MapIndex{}
		return unmarshal(config.index)
	case "wildcard":
		config.index = &WildcardIndex{}
		return unmarshal(config.index)
	case "noop":
		config.index = &NoopIndex{}
		return unmarshal(config.index)
	default:
		return fmt.Errorf("Error in index config: Unknown type '%v'", objectType)
	}
}

type Index interface {
	validate(rootConfig *Config, log *logrus.Entry) error
	GetName() string
	DoesHandleProperty(property string) bool
	DoesHandleInput(input Input) bool
}
