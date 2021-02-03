package config

import (
	"errors"
)

type SourceConfig struct{
	Filesystem *FilesystemSourceConfig
}

func (config *SourceConfig) validate() error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("All sources must have a configuration")
	}

	if len(fields) > 0 {
		return errors.New("A source can only have one configuration")
	}

	return fields[0].validate()
}
