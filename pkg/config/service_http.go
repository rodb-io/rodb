package config

import (
	"errors"
	"github.com/sirupsen/logrus"
	"rods/pkg/util"
)

type HttpService struct {
	Listen     string `yaml:"listen"`
	ErrorsType string `yaml:"errorsType"`
}

func (config *HttpService) validate(rootConfig *Config, log *logrus.Logger) error {
	if config.Listen == "" {
		config.Listen = "127.0.0.1:0"
	}

	if config.ErrorsType == "" {
		log.Debugf("services.http.errorsType is not set. Defaulting to application/json")
		config.ErrorsType = "application/json"
	}

	if !util.IsInArray(config.ErrorsType, []string{
		"application/json",
	}) {
		return errors.New("services.http.errorsType: type '" + config.ErrorsType + "' is not supported.")
	}

	return nil
}
