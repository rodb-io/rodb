package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

type MemoryMapIndex struct {
	Name             string   `yaml:"name"`
	DieOnInputChange *bool    `yaml:"dieOnInputChange"`
	Input            string   `yaml:"input"`
	Columns          []string `yaml:"columns"`
	Logger           *logrus.Entry
}

func (config *MemoryMapIndex) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("memoryMap.name is required")
	}

	_, inputExists := rootConfig.Inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("memoryMap.input: Input '%v' not found in inputs list.", config.Input)
	}

	if config.DieOnInputChange == nil {
		defaultValue := true
		log.Debugf("memoryMap.dieOnInputChange is not set. Assuming 'true'.\n")
		config.DieOnInputChange = &defaultValue
	}

	alreadyExistingColumns := make(map[string]bool)
	for _, columnName := range config.Columns {
		if _, alreadyExists := alreadyExistingColumns[columnName]; alreadyExists {
			return fmt.Errorf("memoryMap.columns: Duplicate column '%v' in array.", columnName)
		}
		alreadyExistingColumns[columnName] = true
	}

	// The columns validity will be validated at runtime

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
