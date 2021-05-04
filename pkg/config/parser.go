package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/util"
)

type Parser struct {
	Integer *IntegerParser
	Float   *FloatParser
	Boolean *BooleanParser
	String  *StringParser
	Json    *JsonParser
	Split   *SplitParser
}

func (config *Parser) UnmarshalYAML(unmarshal func(interface{}) error) error {
	objectType, err := util.GetTypeFromConfigUnmarshaler(unmarshal)
	if err != nil {
		return fmt.Errorf("Error in parser config: %w", err)
	}

	switch objectType {
	case "integer":
		config.Integer = &IntegerParser{}
		return unmarshal(config.Integer)
	case "float":
		config.Float = &FloatParser{}
		return unmarshal(config.Float)
	case "boolean":
		config.Boolean = &BooleanParser{}
		return unmarshal(config.Boolean)
	case "string":
		config.String = &StringParser{}
		return unmarshal(config.String)
	case "json":
		config.Json = &JsonParser{}
		return unmarshal(config.Json)
	case "split":
		config.Split = &SplitParser{}
		return unmarshal(config.Split)
	default:
		return fmt.Errorf("Error in parser config: Unknown type '%v'", objectType)
	}
}

func (config *Parser) validate(rootConfig *Config, log *logrus.Entry) error {
	definedFields := 0
	if config.Integer != nil {
		definedFields++
		err := config.Integer.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}
	if config.Float != nil {
		definedFields++
		err := config.Float.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}
	if config.Boolean != nil {
		definedFields++
		err := config.Boolean.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}
	if config.String != nil {
		definedFields++
		err := config.String.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}
	if config.Json != nil {
		definedFields++
		err := config.Json.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}
	if config.Split != nil {
		definedFields++
		err := config.Split.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}

	if definedFields == 0 {
		return errors.New("One of your parsers does not have a definition.")
	}
	if definedFields > 1 {
		return errors.New("One of your parsers has two different definitions.")
	}

	return nil
}

func (config *Parser) Name() string {
	if config.Integer != nil {
		return config.Integer.Name
	}
	if config.Float != nil {
		return config.Float.Name
	}
	if config.Boolean != nil {
		return config.Boolean.Name
	}
	if config.String != nil {
		return config.String.Name
	}
	if config.Json != nil {
		return config.Json.Name
	}
	if config.Split != nil {
		return config.Split.Name
	}

	return ""
}

func (config *Parser) Primitive() bool {
	if config.Integer != nil {
		return config.Integer.Primitive()
	}
	if config.Float != nil {
		return config.Float.Primitive()
	}
	if config.Boolean != nil {
		return config.Boolean.Primitive()
	}
	if config.String != nil {
		return config.String.Primitive()
	}
	if config.Json != nil {
		return config.Json.Primitive()
	}
	if config.Split != nil {
		return config.Split.Primitive()
	}

	return false
}
