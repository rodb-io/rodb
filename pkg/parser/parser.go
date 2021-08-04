package parser

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type Parser interface {
	Name() string
	GetRegexpPattern() string
	Primitive() bool
	Parse(value string) (interface{}, error)
}

type Config interface {
	Validate(parsers map[string]Config, log *logrus.Entry) error
	GetName() string
	Primitive() bool
}

type List = map[string]Parser

func NewFromConfig(
	config Config,
	parsers List,
) (Parser, error) {
	switch config.(type) {
	case *StringConfig:
		return NewString(config.(*StringConfig))
	case *IntegerConfig:
		return NewInteger(config.(*IntegerConfig)), nil
	case *FloatConfig:
		return NewFloat(config.(*FloatConfig)), nil
	case *BooleanConfig:
		return NewBoolean(config.(*BooleanConfig)), nil
	case *JsonConfig:
		return NewJson(config.(*JsonConfig)), nil
	case *SplitConfig:
		return NewSplit(config.(*SplitConfig), parsers), nil
	default:
		return nil, fmt.Errorf("Unknown parser config type: %#v", config)
	}
}

func NewFromConfigs(
	configs map[string]Config,
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
	case int64:
		aInt := a.(int64)
		bInt, bIsInt := b.(int64)
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
