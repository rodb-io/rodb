package config

import (
	"github.com/sirupsen/logrus"
)

type HttpService struct {
	Listen string `yaml:"listen"`
}

func (config *HttpService) validate(rootConfig *Config, log *logrus.Logger) error {
	if config.Listen == "" {
		config.Listen = "127.0.0.1:0"
	}

	return nil
}
