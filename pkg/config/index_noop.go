package config

import (
	"github.com/sirupsen/logrus"
)

type NoopIndex struct {
}

func (config *NoopIndex) validate(rootConfig *Config, log *logrus.Logger) error {
	return nil
}

func (config *NoopIndex) DoesHandleColumn(column string) bool {
	return true
}

func (config *NoopIndex) DoesHandleInput(input string) bool {
	return true
}
