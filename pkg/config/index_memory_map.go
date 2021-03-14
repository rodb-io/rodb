package config

import (
	"github.com/sirupsen/logrus"
)

type MemoryMapIndex struct {
	Input   string   `yaml:"input"`
	Columns []string `yaml:"columns"`
	Logger  *logrus.Entry
}

func (config *MemoryMapIndex) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	// The input and columns will be validated at runtime
	return nil
}

func (config *MemoryMapIndex) DoesHandleColumn(column string) bool {
	isHandled := false
	for _, handledColumn := range config.Columns {
		if column == handledColumn {
			isHandled = true
			break
		}
	}

	return isHandled
}

func (config *MemoryMapIndex) DoesHandleInput(input string) bool {
	return input == config.Input
}
