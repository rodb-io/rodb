package parser

import (
	"fmt"
	configModule "rodb.io/pkg/config"
)

type Parser interface {
	Name() string
	GetRegexpPattern() string
	Primitive() bool
	Parse(value string) (interface{}, error)
}

type List = map[string]Parser

func NewFromConfig(
	config configModule.Parser,
	parsers List,
) (Parser, error) {
	switch config.(type) {
	case *configModule.StringParser:
		return NewString(config.(*configModule.StringParser))
	case *configModule.IntegerParser:
		return NewInteger(config.(*configModule.IntegerParser)), nil
	case *configModule.FloatParser:
		return NewFloat(config.(*configModule.FloatParser)), nil
	case *configModule.BooleanParser:
		return NewBoolean(config.(*configModule.BooleanParser)), nil
	case *configModule.JsonParser:
		return NewJson(config.(*configModule.JsonParser)), nil
	case *configModule.SplitParser:
		return NewSplit(config.(*configModule.SplitParser), parsers), nil
	default:
		return nil, fmt.Errorf("Unknown parser config type: %#v", config)
	}
}

func NewFromConfigs(
	configs map[string]configModule.Parser,
) (List, error) {
	parsers := make(List)
	for parserName, parserConfig := range configs {
		parser, err := NewFromConfig(parserConfig, parsers)
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

// Compares two values outputted by the parsers
// returns nil if a = b, true if a < b, false if a > b
func Compare(a interface{}, b interface{}) (*bool, error) {
	switch a.(type) {
	case string:
		aString := a.(string)
		bString, bIsString := b.(string)
		if !bIsString {
			return nil, fmt.Errorf("Cannot compare a string with '%#v'", b)
		}

		if aString == bString {
			return nil, nil
		}
		result := aString < bString
		return &result, nil
	case int:
		aInt := a.(int)
		bInt, bIsInt := b.(int)
		if !bIsInt {
			return nil, fmt.Errorf("Cannot compare an integer with '%#v'", b)
		}

		if aInt == bInt {
			return nil, nil
		}
		result := aInt < bInt
		return &result, nil
	case float64:
		aFloat := a.(float64)
		bFloat, bIsFloat := b.(float64)
		if !bIsFloat {
			return nil, fmt.Errorf("Cannot compare a float with '%#v'", b)
		}

		if aFloat == bFloat {
			return nil, nil
		}
		result := aFloat < bFloat
		return &result, nil
	case bool:
		aBool := a.(bool)
		bBool, bIsBool := b.(bool)
		if !bIsBool {
			return nil, fmt.Errorf("Cannot compare a boolean with '%#v'", b)
		}

		if aBool == bBool {
			return nil, nil
		}
		result := (aBool == false && bBool == true)
		return &result, nil
	}

	return nil, fmt.Errorf("Unhandled type for sorting object '%#v'", a)
}
