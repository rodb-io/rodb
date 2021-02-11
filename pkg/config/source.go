package config

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type Source struct {
	Filesystem *FilesystemSource `yaml:"filesystem"`
}

func (config *Source) validate(log *logrus.Logger) error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("All sources must have a configuration")
	}

	if len(fields) > 1 {
		return errors.New("A source can only have one configuration")
	}

	return fields[0].validate(log)
}
