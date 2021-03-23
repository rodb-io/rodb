package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type Source struct {
	Filesystem *FilesystemSource `yaml:"filesystem"`
}

func (config *Source) validate(rootConfig *Config, log *logrus.Entry) error {
	definedFields := 0
	if config.Filesystem != nil {
		definedFields++
		err := config.Filesystem.validate(rootConfig, log)
		if err != nil {
			return err
		}
	}

	if definedFields == 0 {
		return errors.New("All sources must have a configuration")
	}
	if definedFields > 1 {
		return errors.New("A source can only have one configuration")
	}

	return nil
}

func (config *Source) Name() string {
	if config.Filesystem != nil {
		return config.Filesystem.Name
	}

	return ""
}
