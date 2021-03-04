package parser

import (
	"rods/pkg/config"
	"github.com/sirupsen/logrus"
)

type String struct{
	config *config.StringParser
	logger *logrus.Logger
}

func NewString(
	config *config.StringParser,
	log *logrus.Logger,
) *String {
	return &String{
		config: config,
		logger: log,
	}
}

func (string *String) GetRegexpPattern() string {
	return ".*"
}

func (string *String) Parse(value string) (interface{}, error) {
	return value, nil
}
