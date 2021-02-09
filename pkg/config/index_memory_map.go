package config

import (
	"github.com/sirupsen/logrus"
)

type MemoryMapIndexConfig struct{
	Input string `yaml:"input"`
	Column string `yaml:"column"`
}

func (config *MemoryMapIndexConfig) validate(log *logrus.Logger) error {
	// The input and column will be validated at runtime
	return nil
}
