package parser

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"regexp"
	"rods/pkg/config"
	"rods/pkg/util"
	"strings"
)

type Boolean struct {
	config *config.BooleanParser
	logger *logrus.Logger
}

func NewBoolean(
	config *config.BooleanParser,
	log *logrus.Logger,
) *Boolean {
	return &Boolean{
		config: config,
		logger: log,
	}
}

func (boolean *Boolean) GetRegexpPattern() string {
	values := make([]string, len(boolean.config.TrueValues)+len(boolean.config.FalseValues))
	for _, value := range boolean.config.TrueValues {
		values = append(values, regexp.QuoteMeta(value))
	}
	for _, value := range boolean.config.FalseValues {
		values = append(values, regexp.QuoteMeta(value))
	}
	return "(" + strings.Join(values, "|") + ")"
}

func (boolean *Boolean) Parse(value string) (interface{}, error) {
	if util.IsInArray(value, boolean.config.TrueValues) {
		return true, nil
	}
	if util.IsInArray(value, boolean.config.FalseValues) {
		return false, nil
	}

	return nil, fmt.Errorf("The value '%v' was found but is neither declared in trueValues or falseValues.", value)
}
