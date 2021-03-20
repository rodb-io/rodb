package config

import (
	"github.com/sirupsen/logrus"
)

type Sort struct {
	Logger    *logrus.Entry
	Column    string `yaml:"column"`
	Ascending *bool  `yaml:"ascending"`
}

func (config *Sort) validate(
	rootConfig *Config,
	log *logrus.Entry,
	logPrefix string,
) error {
	config.Logger = log

	if config.Ascending == nil {
		log.Debugf(logPrefix + "ascending is not set. Assuming 'true'.\n")
		defaultAscending := true
		config.Ascending = &defaultAscending
	}

	return nil
}
