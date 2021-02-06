package config

import (
	"github.com/sirupsen/logrus"
)

type FilesystemSourceConfig struct{
	Path string
}

func (config *FilesystemSourceConfig) validate(log *logrus.Logger) error {
	// The path will be validated at runtime
	return nil
}