package output

import (
	"errors"
	"github.com/sirupsen/logrus"
	indexPackage "rodb.io/pkg/index"
	inputPackage "rodb.io/pkg/input"
	parserPackage "rodb.io/pkg/parser"
)

type GraphQLConfig struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Logger *logrus.Entry
}

func (config *GraphQLConfig) GetName() string {
	return config.Name
}

func (config *GraphQLConfig) Validate(
	inputs map[string]inputPackage.Config,
	indexes map[string]indexPackage.Config,
	parsers map[string]parserPackage.Config,
	log *logrus.Entry,
) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("graphql.name is required")
	}

	return nil
}
