package types

import (
	"errors"
)

type Type interface{
	GetRegexpPattern() string
	Parse(value string) (interface{}, error)
}

func NewFromString(typeName string) (Type, error) {
	switch typeName {
		case "string":
			return NewString(), nil
		case "integer":
			return NewInteger(), nil
		case "float":
			return NewFloat(), nil
		case "boolean":
			return NewBoolean(), nil
		default:
			return nil, errors.New("Unknown type '" + string(typeName) + "'")
	}
}
