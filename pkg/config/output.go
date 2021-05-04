package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/util"
)

type Output struct {
	GraphQL    *GraphQLOutput
	JsonArray  *JsonArrayOutput
	JsonObject *JsonObjectOutput
}

func (config *Output) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in output config: %w", err)
	}

	switch objectType {
	case "graphql":
		config.GraphQL = &GraphQLOutput{}
		return unmarshal(config.GraphQL)
	case "jsonArray":
		config.JsonArray = &JsonArrayOutput{}
		return unmarshal(config.JsonArray)
	case "jsonObject":
		config.JsonObject = &JsonObjectOutput{}
		return unmarshal(config.JsonObject)
	default:
		return fmt.Errorf("Error in output config: Unknown type '%v'", objectType)
	}
}

func (config *Output) validate(rootConfig *Config, log *logrus.Entry) error {
	definedFields := 0
	if config.GraphQL != nil {
		definedFields++
		err := config.GraphQL.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}
	if config.JsonArray != nil {
		definedFields++
		err := config.JsonArray.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}
	if config.JsonObject != nil {
		definedFields++
		err := config.JsonObject.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}

	if definedFields == 0 {
		return errors.New("One of your outputs does not have a definition.")
	}
	if definedFields > 1 {
		return errors.New("One of your outputs has two different definitions.")
	}

	return nil
}

func (config *Output) Name() string {
	if config.GraphQL != nil {
		return config.GraphQL.Name
	}
	if config.JsonArray != nil {
		return config.JsonArray.Name
	}
	if config.JsonObject != nil {
		return config.JsonObject.Name
	}

	return ""
}
