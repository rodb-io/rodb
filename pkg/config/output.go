package config

import (
	"fmt"
	"github.com/rodb-io/rodb/pkg/output"
	"github.com/rodb-io/rodb/pkg/util"
)

type outputParser struct {
	output output.Config
}

func (config *outputParser) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in output config: %w", err)
	}

	switch objectType {
	case "graphql":
		config.output = &output.GraphQLConfig{}
		return unmarshal(config.output)
	case "jsonArray":
		config.output = &output.JsonArrayConfig{}
		return unmarshal(config.output)
	case "jsonObject":
		config.output = &output.JsonObjectConfig{}
		return unmarshal(config.output)
	default:
		return fmt.Errorf("Error in output config: Unknown type '%v'", objectType)
	}
}
