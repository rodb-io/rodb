package config

import (
	"errors"
	"strings"
	"github.com/sirupsen/logrus"
)

type JsonObjectOutputConfig struct{
	Service string
	Endpoint string
	Index string
}

func (config *JsonObjectOutputConfig) validate(log *logrus.Logger) error {
	// The service will be validated at runtime
	// The index will be validated at runtime

	if config.Endpoint == "" {
		return errors.New("You must specify a non-empty jsonObject endpoint")
	}

	if !strings.Contains(config.Endpoint, "?") {
		return errors.New("A jsonObject endpoint must specify the identifier's location with '?'. For example \"/product/?\".")
	}

	return nil
}
