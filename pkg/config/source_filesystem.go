package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type FilesystemSource struct {
	Name             string `yaml:"name"`
	Logger           *logrus.Entry
}

func (config *FilesystemSource) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("filesystem.name is required")
	}

	return nil
}
