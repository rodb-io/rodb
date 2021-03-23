package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type Output struct {
	GraphQL    *GraphQLOutput    `yaml:"graphql"`
	JsonArray  *JsonArrayOutput  `yaml:"jsonArray"`
	JsonObject *JsonObjectOutput `yaml:"jsonObject"`
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

func (config *Output) Services() []string {
	if config.GraphQL != nil {
		return config.GraphQL.Services
	}
	if config.JsonArray != nil {
		return config.JsonArray.Services
	}
	if config.JsonObject != nil {
		return config.JsonObject.Services
	}

	return []string{}
}

func (config *Output) Endpoint() string {
	if config.GraphQL != nil {
		return config.GraphQL.Endpoint
	}
	if config.JsonArray != nil {
		return config.JsonArray.Endpoint
	}
	if config.JsonObject != nil {
		return config.JsonObject.Endpoint
	}

	return ""
}
