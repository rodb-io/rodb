package parser

import (
	"errors"
	"rods/pkg/config"
)

type Parser interface {
	GetRegexpPattern() string
	Parse(value string) (interface{}, error)
}

type List = map[string]Parser

func NewFromConfig(
	config config.Parser,
) (Parser, error) {
	if config.String != nil {
		return NewString(config.String)
	}
	if config.Integer != nil {
		return NewInteger(config.Integer), nil
	}
	if config.Float != nil {
		return NewFloat(config.Float), nil
	}
	if config.Boolean != nil {
		return NewBoolean(config.Boolean), nil
	}

	return nil, errors.New("Failed to initialize parser")
}

func NewFromConfigs(
	configs map[string]config.Parser,
) (List, error) {
	parsers := make(List)
	for parserName, parserConfig := range configs {
		parser, err := NewFromConfig(parserConfig)
		if err != nil {
			return nil, err
		}
		parsers[parserName] = parser
	}

	return parsers, nil
}

func Close(parsers List) error {
	return nil
}
