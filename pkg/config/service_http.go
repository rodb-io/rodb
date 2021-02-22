package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type HttpService struct {
	Port uint16 `yaml:"port"`
}

func (config *HttpService) validate(log *logrus.Logger) error {
	if config.Port == 0 {
		return errors.New("http.port is not defined. This setting is required.")
	}

	return nil
}
