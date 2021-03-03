package types

import (
	"errors"
)

type Boolean struct{
}

func NewBoolean() *Boolean {
	return &Boolean{
	}
}

func (boolean *Boolean) GetRegexpPattern() string {
	return "(true|false|1|0|TRUE|FALSE)"
}

func (boolean *Boolean) Parse(value string) (interface{}, error) {
	if value == "true" || value == "1" || value == "TRUE" {
		return true, nil
	}

	if value == "false" || value == "0" || value == "FALSE" {
		return false, nil
	}

	return nil, errors.New("The value '" + value + "' cannot be parsed as a boolean.")
}
