package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/util"
)

type Service struct {
	Http *HttpService
}

func (config *Service) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in service config: %w", err)
	}

	switch objectType {
	case "http":
		config.Http = &HttpService{}
		return unmarshal(config.Http)
	default:
		return fmt.Errorf("Error in service config: Unknown type '%v'", objectType)
	}
}

func (config *Service) validate(rootConfig *Config, log *logrus.Entry) error {
	definedFields := 0
	if config.Http != nil {
		definedFields++
		err := config.Http.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}

	if definedFields == 0 {
		return errors.New("All services must have a configuration")
	}
	if definedFields > 1 {
		return errors.New("A services can only have one configuration")
	}

	return nil
}

func (config *Service) Name() string {
	if config.Http != nil {
		return config.Http.Name
	}

	return ""
}
