package config

import (
	"github.com/sirupsen/logrus"
)

type MemoryMapMultipleIndexConfig struct{
	Input string `yaml:"input"`
	Column string `yaml:"column"`
}

func (config *MemoryMapMultipleIndexConfig) validate(log *logrus.Logger) error {
	// The input and column will be validated at runtime
	return nil
}
