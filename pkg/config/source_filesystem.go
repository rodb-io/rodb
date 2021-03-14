package config

import (
	"github.com/sirupsen/logrus"
)

type FilesystemSource struct {
	Path string `yaml:"path"`
}

func (config *FilesystemSource) validate(rootConfig *Config, log *logrus.Logger) error {
	// The path will be validated at runtime
	return nil
}
