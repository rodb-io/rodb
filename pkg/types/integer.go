package types

import (
	"strconv"
)

type Integer struct{
}

func NewInteger() *Integer {
	return &Integer{
	}
}

func (integer *Integer) GetRegexpPattern() string {
	return "[-]?[0-9]+"
}

func (integer *Integer) Parse(value string) (interface{}, error) {
	return strconv.Atoi(value)
}
