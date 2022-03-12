package output

import (
	"errors"
	"github.com/sirupsen/logrus"
	indexPackage "github.com/rodb-io/rodb/pkg/index"
	inputPackage "github.com/rodb-io/rodb/pkg/input"
	parserPackage "github.com/rodb-io/rodb/pkg/parser"
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
