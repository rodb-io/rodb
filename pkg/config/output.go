package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type OutputConfig struct{
	GraphQL *GraphQLOutputConfig
	JsonArray *JsonArrayOutputConfig
	JsonObject *JsonObjectOutputConfig
}

func (config *OutputConfig) validate(log *logrus.Logger) error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("One of your outputs does not have a definition.")
	}

	if len(fields) > 0 {
		return errors.New("One of your outputs has two different definitions.")
	}

	return fields[0].validate(log)
}
