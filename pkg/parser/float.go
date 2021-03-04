package parser

import (
	"strconv"
	"strings"
	"rods/pkg/config"
	"rods/pkg/util"
	"github.com/sirupsen/logrus"
)

type Float struct{
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
	return "[-]?[0-9]+([.][0-9]+)?"
}

func (float *Float) Parse(value string) (interface{}, error) {
	cleanedValue := util.RemoveCharacters(value, float.config.IgnoreCharacters)
	if float.config.DecimalSeparator != "." {
		cleanedValue = strings.ReplaceAll(cleanedValue, float.config.DecimalSeparator, ".")
	}

	return strconv.ParseFloat(cleanedValue, 64)
}
