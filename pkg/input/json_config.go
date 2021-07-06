package input

import (
	"errors"
	"github.com/sirupsen/logrus"
	"os"
)

type JsonConfig struct {
	Name             string `yaml:"name"`
	Type             string `yaml:"type"`
	Path             string `yaml:"path"`
	DieOnInputChange *bool  `yaml:"dieOnInputChange"`
	Logger           *logrus.Entry
}

func (config *JsonConfig) GetName() string {
	return config.Name
}

func (config *JsonConfig) ShouldDieOnInputChange() bool {
	return config.DieOnInputChange == nil || *config.DieOnInputChange
}

func (config *JsonConfig) Validate(parsers map[string]Parser, log *logrus.Entry) error {
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
