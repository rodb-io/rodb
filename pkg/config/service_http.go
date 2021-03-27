package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rods/pkg/util"
)

type HttpService struct {
	Name       string   `yaml:"name"`
	Listen     string   `yaml:"listen"`
	ErrorsType string   `yaml:"errorsType"`
	Outputs    []string `yaml:"outputs"`
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

	for i, output := range config.Outputs {
		if output == "" {
			return fmt.Errorf("http.output[%v]: is empty. This field is required", i)
		}
		_, outputExists := rootConfig.Outputs[output]
		if !outputExists {
			return fmt.Errorf("http.output[%v]: Output '%v' not found in outputs list.", i, output)
		}

	}

	if !util.IsInArray(config.ErrorsType, []string{
		"application/json",
	}) {
		return errors.New("http.errorsType: type '" + config.ErrorsType + "' is not supported.")
	}

	return nil
}
