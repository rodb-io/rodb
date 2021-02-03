package config

import (
	"errors"
)

type OutputConfig struct{
	GraphQL *GraphQLOutputConfig
	JsonArray *JsonArrayOutputConfig
	JsonObject *JsonObjectOutputConfig
}

func (config *OutputConfig) validate() error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("All outputs must have a configuration")
	}

	if len(fields) > 0 {
		return errors.New("An output can only have one configuration")
	}

	return fields[0].validate()
}
