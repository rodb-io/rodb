package types

import (
	"strconv"
)

type Float struct{
}

func NewFloat() *Float {
	return &Float{
	}
}

func (float *Float) GetRegexpPattern() string {
	return "[-]?[0-9]+([.][0-9]+)?"
}

func (float *Float) Parse(value string) (interface{}, error) {
	return strconv.ParseFloat(value, 64)
}
