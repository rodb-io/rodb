package config

import (
	"github.com/sirupsen/logrus"
)

type HttpService struct {
	Port uint16 `yaml:"port"`
}

func (config *HttpService) validate(log *logrus.Logger) error {
	// Not setting the port or setting it to 0
	// means that is will be automatically assigned
	return nil
}
