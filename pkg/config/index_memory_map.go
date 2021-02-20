package config

import (
	"github.com/sirupsen/logrus"
)

type MemoryMapIndex struct {
	Input   string   `yaml:"input"`
	Columns []string `yaml:"columns"`
}

func (config *MemoryMapIndex) validate(log *logrus.Logger) error {
	// The input and columns will be validated at runtime
	return nil
}
