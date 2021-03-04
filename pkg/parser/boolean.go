package parser

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/util"
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
	return "(true|false|1|0|TRUE|FALSE)"
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
