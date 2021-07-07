package parser

import (
	"regexp"
	"rodb.io/pkg/util"
	"strconv"
	"strings"
)

type Float struct {
	config *FloatConfig
}

func NewFloat(
	config *FloatConfig,
) *Float {
	return &Float{
		config: config,
	}
}

func (float *Float) Name() string {
	return float.config.Name
}

func (float *Float) Primitive() bool {
	return float.config.Primitive()
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
