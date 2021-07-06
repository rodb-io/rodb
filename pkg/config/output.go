package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/util"
)

type outputParser struct {
	output Output
}

func (config *outputParser) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in output config: %w", err)
	}

	switch objectType {
	case "graphql":
		config.output = &GraphQLOutput{}
		return unmarshal(config.output)
	case "jsonArray":
		config.output = &JsonArrayOutput{}
		return unmarshal(config.output)
	case "jsonObject":
		config.output = &JsonObjectOutput{}
		return unmarshal(config.output)
	default:
		return fmt.Errorf("Error in output config: Unknown type '%v'", objectType)
	}
}

type Output interface {
	Validate(
		inputs map[string]Input,
		indexes map[string]Index,
		parsers map[string]Parser,
		log *logrus.Entry,
	) error
	GetName() string
}
