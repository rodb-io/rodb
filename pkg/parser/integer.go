package parser

import (
	"github.com/sirupsen/logrus"
	"regexp"
	"rods/pkg/config"
	"rods/pkg/util"
	"strconv"
)

type Integer struct {
	config *config.IntegerParser
	logger *logrus.Logger
}

func NewInteger(
	config *config.IntegerParser,
	log *logrus.Logger,
) *Integer {
	return &Integer{
		config: config,
		logger: log,
	}
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
