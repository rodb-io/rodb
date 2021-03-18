package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

type JsonArrayOutput struct {
	Services []string                         `yaml:"services"`
	Endpoint string                           `yaml:"endpoint"`
	Limit    JsonArrayOutputLimit             `yaml:"limit"`
	Offset   JsonArrayOutputOffset            `yaml:"offset"`
	Search   map[string]JsonArrayOutputSearch `yaml:"search"`
	Logger   *logrus.Entry
}

type JsonArrayOutputLimit struct {
	Default uint
	Max     uint
	Param   string
}

type JsonArrayOutputOffset struct {
	Param string
}

type JsonArrayOutputSearch struct {
	Index string
}

func (config *JsonArrayOutput) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Endpoint == "" {
		return errors.New("jsonArray.endpoint is not defined. This setting is required")
	}

	if len(config.Services) == 0 {
		return errors.New("jsonArray.services is empty. As least one is required.")
	}
	alreadyExistingServices := make(map[string]bool)
	for _, serviceName := range config.Services {
		_, serviceExists := rootConfig.Services[serviceName]
		if !serviceExists {
			return fmt.Errorf("jsonArray.services: Service '%v' not found in services list.", serviceName)
		}

		if _, alreadyExists := alreadyExistingServices[serviceName]; alreadyExists {
			return fmt.Errorf("jsonArray.services: Duplicate service '%v' in array.", serviceName)
		}
		alreadyExistingServices[serviceName] = true
	}

	err := config.Limit.validate(rootConfig, log)
	if err != nil {
		return err
	}

	err = config.Offset.validate(rootConfig, log)
	if err != nil {
		return err
	}

	for configSearchParamName, configSearchParam := range config.Search {
		err := configSearchParam.validate(rootConfig, log)
		if err != nil {
			return fmt.Errorf("jsonArray.search.%v.%w", configSearchParamName, err)
		}
	}

	return nil
}

func (config *JsonArrayOutputLimit) validate(rootConfig *Config, log *logrus.Entry) error {
	if config.Default == 0 {
		log.Debug("jsonArray.limit.default not set. Assuming '100'")
		config.Default = 100
	}

	if config.Max == 0 {
		log.Debug("jsonArray.limit.max not set. Assuming '1000'")
		config.Max = 1000
	}

	if config.Param == "" {
		log.Debug("jsonArray.limit.param not set. Assuming 'limit'")
		config.Param = "limit"
	}

	return nil
}

func (config *JsonArrayOutputOffset) validate(rootConfig *Config, log *logrus.Entry) error {
	if config.Param == "" {
		log.Debug("jsonArray.offset.param not set. Assuming 'offset'")
		config.Param = "offset"
	}

	return nil
}

func (config *JsonArrayOutputSearch) validate(rootConfig *Config, log *logrus.Entry) error {
	if config.Index == "" {
		log.Debugf("jsonArray.search[].index is empty. Assuming 'default'.\n")
		config.Index = "default"
	}
	_, indexExists := rootConfig.Indexes[config.Index]
	if !indexExists {
		return fmt.Errorf("index: Index '%v' not found in indexes list.", config.Index)
	}

	return nil
}
