package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type FilesystemSource struct {
	Name             string `yaml:"name"`
	DieOnInputChange *bool  `yaml:"dieOnInputChange"`
	Logger           *logrus.Entry
}

func (config *FilesystemSource) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("filesystem.name is required")
	}

	if config.DieOnInputChange == nil {
		defaultValue := true
		log.Debugf("filesystem.dieOnInputChange is not set. Assuming 'true'.\n")
		config.DieOnInputChange = &defaultValue
	}

	return nil
}
