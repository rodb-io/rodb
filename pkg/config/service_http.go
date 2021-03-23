package config

import (
	"errors"
	"github.com/sirupsen/logrus"
	"rods/pkg/util"
)

type HttpService struct {
	Name       string `yaml:"name"`
	Listen     string `yaml:"listen"`
	ErrorsType string `yaml:"errorsType"`
	Logger     *logrus.Entry
}

func (config *HttpService) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("http.name is required")
	}

	if config.Listen == "" {
		config.Listen = "127.0.0.1:0"
	}

	if config.ErrorsType == "" {
		log.Debugf("http.errorsType is not set. Defaulting to application/json")
		config.ErrorsType = "application/json"
	}

	if !util.IsInArray(config.ErrorsType, []string{
		"application/json",
	}) {
		return errors.New("http.errorsType: type '" + config.ErrorsType + "' is not supported.")
	}

	return nil
}
