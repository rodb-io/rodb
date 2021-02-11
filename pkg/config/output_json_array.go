package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type JsonArrayOutputConfig struct {
	Service  string                                 `yaml:"service"`
	Endpoint string                                 `yaml:"endpoint"`
	Limit    JsonArrayOutputLimitConfig             `yaml:"limit"`
	Offset   JsonArrayOutputOffsetConfig            `yaml:"offset"`
	Search   map[string]JsonArrayOutputSearchConfig `yaml:"search"`
}

type JsonArrayOutputLimitConfig struct {
	Default uint
	Max     uint
	Param   string
}

type JsonArrayOutputOffsetConfig struct {
	Param string
}

type JsonArrayOutputSearchConfig struct {
	Index string
}

func (config *JsonArrayOutputConfig) validate(log *logrus.Logger) error {
	// The service will be validated at runtime

	if config.Endpoint == "" {
		return errors.New("jsonArray.endpoint is not defined. This setting is required")
	}

	err := config.Limit.validate(log)
	if err != nil {
		return err
	}

	err = config.Offset.validate(log)
	if err != nil {
		return err
	}

	for _, configSearchParam := range config.Search {
		err := configSearchParam.validate(log)
		if err != nil {
			return err
		}
	}

	return nil
}

func (config *JsonArrayOutputLimitConfig) validate(log *logrus.Logger) error {
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

func (config *JsonArrayOutputOffsetConfig) validate(log *logrus.Logger) error {
	if config.Param == "" {
		log.Debug("jsonArray.offset.param not set. Assuming 'offset'")
		config.Param = "offset"
	}

	return nil
}

func (config *JsonArrayOutputSearchConfig) validate(log *logrus.Logger) error {
	// The index will be validated at runtime
	return nil
}
