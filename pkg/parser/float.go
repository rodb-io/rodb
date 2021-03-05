package parser

import (
	"github.com/sirupsen/logrus"
	"regexp"
	"rods/pkg/config"
	"rods/pkg/util"
	"strconv"
	"strings"
)

type Float struct {
	config *config.FloatParser
	logger *logrus.Logger
}

func NewFloat(
	config *config.FloatParser,
	log *logrus.Logger,
) *Float {
	return &Float{
		config: config,
		logger: log,
	}
}

func (float *Float) GetRegexpPattern() string {
	separator := regexp.QuoteMeta(float.config.DecimalSeparator)
	ignore := regexp.QuoteMeta(float.config.IgnoreCharacters)
	ignoreBegin := ""
	if ignore != "" {
		ignoreBegin = "[" + ignore + "]*"
	}
	return ignoreBegin + "[-]?[0-9" + ignore + "]+([" + separator + "][0-9" + ignore + "]+)?"
}

func (float *Float) Parse(value string) (interface{}, error) {
	cleanedValue := util.RemoveCharacters(value, float.config.IgnoreCharacters)
	if float.config.DecimalSeparator != "." {
		cleanedValue = strings.ReplaceAll(cleanedValue, float.config.DecimalSeparator, ".")
	}

	return strconv.ParseFloat(cleanedValue, 64)
}
