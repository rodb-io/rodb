package parser

import (
	"regexp"
	"rodb.io/pkg/util"
	"strconv"
)

type Integer struct {
	config *IntegerConfig
}

func NewInteger(
	config *IntegerConfig,
) *Integer {
	return &Integer{
		config: config,
	}
}

func (integer *Integer) Name() string {
	return integer.config.Name
}

func (integer *Integer) Primitive() bool {
	return integer.config.Primitive()
}

func (integer *Integer) GetRegexpPattern() string {
	ignore := regexp.QuoteMeta(integer.config.IgnoreCharacters)
	ignoreBegin := ""
	if ignore != "" {
		ignoreBegin = "[" + ignore + "]*"
	}
	return ignoreBegin + "[-]?[0-9" + ignore + "]+"
}

func (integer *Integer) Parse(value string) (interface{}, error) {
	cleanedValue := util.RemoveCharacters(value, integer.config.IgnoreCharacters)
	integerValue, err := strconv.Atoi(cleanedValue)
	return int64(integerValue), err
}
