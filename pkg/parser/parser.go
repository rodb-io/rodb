package parser

import (
	"errors"
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
)

type Parser interface {
	GetRegexpPattern() string
	Parse(value string) (interface{}, error)
}

type List = map[string]Parser

func NewFromConfig(
	config config.Parser,
	log *logrus.Logger,
) (Parser, error) {
	if config.String != nil {
		return NewString(config.String, log)
	}
	if config.Integer != nil {
		return NewInteger(config.Integer, log), nil
	}
	if config.Float != nil {
		return NewFloat(config.Float, log), nil
	}
	if config.Boolean != nil {
		return NewBoolean(config.Boolean, log), nil
	}

	return nil, errors.New("Failed to initialize parser")
}

func NewFromConfigs(
	configs map[string]config.Parser,
	log *logrus.Logger,
) (List, error) {
	parsers := make(List)
	for parserName, parserConfig := range configs {
		parser, err := NewFromConfig(parserConfig, log)
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
