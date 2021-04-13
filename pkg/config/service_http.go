package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"rodb.io/pkg/util"
)

type HttpService struct {
	Name       string             `yaml:"name"`
	Listen     string             `yaml:"listen"`
	ErrorsType string             `yaml:"errorsType"`
	Routes     []HttpServiceRoute `yaml:"routes"`
	Logger     *logrus.Entry
}

type HttpServiceRoute struct {
	Output string `yaml:"output"`
	Path   string `yaml:"path"`
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

	alreadyExistingPaths := make(map[string]bool)
	for i, routeConfig := range config.Routes {
		err := routeConfig.validate(rootConfig, log)
		if err != nil {
			return fmt.Errorf("http.route[%v].%w", i, err)
		}

		if _, alreadyExists := alreadyExistingPaths[routeConfig.Path]; alreadyExists {
			return fmt.Errorf("http.output[%v]: Duplicate path '%v' in array.", i, routeConfig.Path)
		}
		alreadyExistingPaths[routeConfig.Path] = true
	}

	if !util.IsInArray(config.ErrorsType, []string{
		"application/json",
	}) {
		return errors.New("http.errorsType: type '" + config.ErrorsType + "' is not supported.")
	}

	return nil
}

func (config *HttpServiceRoute) validate(rootConfig *Config, log *logrus.Entry) error {
	if config.Output == "" {
		return fmt.Errorf("output is empty. This field is required")
	}

	if config.Path == "" {
		return fmt.Errorf("path is empty. This field is required")
	}

	_, outputExists := rootConfig.Outputs[config.Output]
	if !outputExists {
		return fmt.Errorf("output '%v' not found in outputs list.", config.Output)
	}

	return nil
}
