package parser

import (
	"regexp"
	"rodb.io/pkg/config"
	"rodb.io/pkg/util"
	"strconv"
)

type Integer struct {
	config *config.IntegerParser
}

func NewInteger(
	config *config.IntegerParser,
) *Integer {
	return &Integer{
		config: config,
	}
}

func (integer *Integer) Name() string {
	return integer.config.Name
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
	return strconv.Atoi(cleanedValue)
}
