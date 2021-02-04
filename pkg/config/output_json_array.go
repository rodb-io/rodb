package config

import (
	"errors"
)

type JsonArrayOutputConfig struct{
	Service string
	Endpoint string
	Limit JsonArrayOutputLimitConfig
	Offset JsonArrayOutputOffsetConfig
	Search map[string]JsonArrayOutputSearchConfig
}

type JsonArrayOutputLimitConfig struct{
	Default uint
	Max uint
	Param string
}

type JsonArrayOutputOffsetConfig struct{
	Param string
}

type JsonArrayOutputSearchConfig struct{
	Index string
}

func (config *JsonArrayOutputConfig) validate() error {
	// The service will be validated at runtime

	if config.Endpoint == "" {
		return errors.New("You must specify a non-empty jsonArray endpoint")
	}

	err := config.Limit.validate()
	if err != nil {
		return err
	}

	err = config.Offset.validate()
	if err != nil {
		return err
	}

	for _, configSearchParam := range config.Search {
		err := configSearchParam.validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (config *JsonArrayOutputLimitConfig) validate() error {
	if config.Default == 0 {
		config.Default = 100
	}

	if config.Max == 0 {
		config.Max = 1000
	}

	if config.Param == "" {
		config.Param = "limit"
	}

	return nil
}

func (config *JsonArrayOutputOffsetConfig) validate() error {
	if config.Param == "" {
		config.Param = "offset"
	}

	return nil
}

func (config *JsonArrayOutputSearchConfig) validate() error {
	// The index will be validated at runtime
	return nil
}
