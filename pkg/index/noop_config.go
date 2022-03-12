package index

import (
	"errors"
	"github.com/rodb-io/rodb/pkg/input"
	"github.com/sirupsen/logrus"
)

type NoopConfig struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Logger *logrus.Entry
}

func (config *NoopConfig) Validate(inputs map[string]input.Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("noop.name is required")
	}

	return nil
}

func (config *NoopConfig) GetName() string {
	return config.Name
}

func (config *NoopConfig) DoesHandleProperty(property string) bool {
	return true
}

func (config *NoopConfig) DoesHandleInput(input input.Config) bool {
	return true
}
