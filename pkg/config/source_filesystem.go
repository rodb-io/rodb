package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type FilesystemSource struct {
	Name   string `yaml:"name"`
	Path   string `yaml:"path"`
	Logger *logrus.Entry
}

func (config *FilesystemSource) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("filesystem.name is required")
	}

	// The path will be validated at runtime
	return nil
}

func (config *FilesystemSource) getName() string {
	return config.Name
}
