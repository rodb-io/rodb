package types

import (
	"errors"
)

type Type interface{
	GetRegexpPattern() string
	Parse(value string) (interface{}, error)
}

var types = map[string]Type {
	"string": NewString(),
	"integer": NewInteger(),
	"float": NewFloat(),
	"boolean": NewBoolean(),
}

func FromString(typeName string) (Type, error) {
	if typeObject, typeExists := types[typeName]; typeExists {
		return typeObject, nil
	} else {
		return nil, errors.New("Unknown type '" + string(typeName) + "'")
	}
}
