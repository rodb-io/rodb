package config

import (
	"github.com/sirupsen/logrus"
)

type MemoryMapIndex struct {
	Input  string `yaml:"input"`
	Column string `yaml:"column"`
}

func (config *MemoryMapIndex) validate(log *logrus.Logger) error {
	// The input and column will be validated at runtime
	return nil
}
