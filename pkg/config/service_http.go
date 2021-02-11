package config

import (
	"github.com/sirupsen/logrus"
)

type HttpService struct {
	Port uint16 `yaml:"port"`
}

func (config *HttpService) validate(log *logrus.Logger) error {
	return nil
}
