package config

import (
	"errors"
	"github.com/sirupsen/logrus"
	"strings"
)

type JsonObjectOutput struct {
	Service  string `yaml:"service"`
	Endpoint string `yaml:"endpoint"`
	Index    string `yaml:"index"`
	Input    string `yaml:"input"`
}

func (config *JsonObjectOutput) validate(log *logrus.Logger) error {
	// The service will be validated at runtime
	// The index will be validated at runtime
	// The default index value "" matches the dumb index
	// The input will be validated at runtime

	if config.Endpoint == "" {
		return errors.New("You must specify a non-empty jsonObject endpoint")
	}

	if !strings.Contains(config.Endpoint, "?") {
		return errors.New("A jsonObject endpoint must specify the identifier's location with '?'. For example \"/product/?\".")
	}

	return nil
}
