package parser

import (
	"errors"
	"rods/pkg/config"
	"github.com/sirupsen/logrus"
)

type Parser interface{
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
		return NewInteger(config.Integer, log)
	}
	if config.Float != nil {
		return NewFloat(config.Float, log)
	}
	if config.Boolean != nil {
		return NewBoolean(config.Boolean, log)
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

	// Handling the default parsers
	for parserName, parserConfig := range map[string]config.Parser {
		"string": {
			String: &config.StringParser{},
		},
		"integer": {
			Integer: &config.IntegerParser{
				IgnoreCharacters: "",
			},
		},
		"float": {
			Float: &config.FloatParser{
				DecimalSeparator: ".",
				IgnoreCharacters: "",
			},
		},
		"boolean": {
			Boolean: &config.BooleanParser{
				TrueValues: []string{"true", "1", "TRUE"},
				FalseValues: []string{"false", "0", "FALSE"},
			},
		},
	} {
		if _, exists := parsers[parserName]; exists {
			log.Warnf("You have declared a parser named '%v', which is a reserved keyword. Your parser will not work.\n", parserName)
		}
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
