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

func (config *Output) validate(log *logrus.Logger) error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("One of your outputs does not have a definition.")
	}

	if len(fields) > 1 {
		return errors.New("One of your outputs has two different definitions.")
	}

	return fields[0].validate(log)
}
