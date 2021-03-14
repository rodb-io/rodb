package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type MemoryMapIndex struct {
	Input   string   `yaml:"input"`
	Columns []string `yaml:"columns"`
	Logger  *logrus.Entry
}

func (config *MemoryMapIndex) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	_, inputExists := rootConfig.Inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("memoryMap.input: Input '%v' not found in inputs list.", config.Input)
	}

	// The columns will be validated at runtime

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
