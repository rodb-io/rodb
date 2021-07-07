package config

import (
	"fmt"
	"rodb.io/pkg/service"
	"rodb.io/pkg/util"
)

type serviceParser struct {
	service service.Config
}

func (config *serviceParser) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in service config: %w", err)
	}

	switch objectType {
	case "http":
		config.service = &service.HttpConfig{}
		return unmarshal(config.service)
	default:
		return fmt.Errorf("Error in service config: Unknown type '%v'", objectType)
	}
}
