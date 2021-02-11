package config

import (
	"github.com/sirupsen/logrus"
)

type HttpServiceConfig struct {
	Port uint16 `yaml:"port"`
}

func (config *HttpServiceConfig) validate(log *logrus.Logger) error {
	return nil
}
