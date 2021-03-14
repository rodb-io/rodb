package config

import (
	"github.com/sirupsen/logrus"
)

type FilesystemSource struct {
	Path   string `yaml:"path"`
	Logger *logrus.Entry
}

func (config *FilesystemSource) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	// The path will be validated at runtime
	return nil
}
