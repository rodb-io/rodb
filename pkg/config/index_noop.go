package config

import (
	"github.com/sirupsen/logrus"
)

type NoopIndex struct {
	Logger *logrus.Entry
}

func (config *NoopIndex) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	return nil
}

func (config *NoopIndex) DoesHandleColumn(column string) bool {
	return true
}

func (config *NoopIndex) DoesHandleInput(input string) bool {
	return true
}
