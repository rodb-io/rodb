package config

import (
	"errors"
	"github.com/sirupsen/logrus"
	"os"
)

type JsonInput struct {
	Name             string `yaml:"name"`
	Path             string `yaml:"path"`
	DieOnInputChange *bool  `yaml:"dieOnInputChange"`
	Logger           *logrus.Entry
}

func (config *JsonInput) ShouldDieOnInputChange() bool {
	return config.DieOnInputChange == nil || *config.DieOnInputChange
}

func (config *JsonInput) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("json.name is required")
	}

	if config.DieOnInputChange == nil {
		defaultValue := true
		log.Debugf("json.dieOnInputChange is not set. Assuming 'true'.\n")
		config.DieOnInputChange = &defaultValue
	}

	fileInfo, err := os.Stat(config.Path)
	if os.IsNotExist(err) {
		return errors.New("The json file '" + config.Path + "' does not exist")
	}
	if fileInfo.IsDir() {
		return errors.New("The path '" + config.Path + "' is not a file")
	}

	return nil
}
