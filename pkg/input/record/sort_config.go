package record

import (
	"github.com/sirupsen/logrus"
)

type SortConfig struct {
	Logger    *logrus.Entry
	Property  string `yaml:"property"`
	Ascending *bool  `yaml:"ascending"`
}

func (config *SortConfig) IsAscending() bool {
	return config.Ascending == nil || *config.Ascending
}

func (config *SortConfig) Validate(
	log *logrus.Entry,
	logPrefix string,
) error {
	config.Logger = log

	// Property will be validated at runtime, because some fields
	// cannot be checked before runtime (json parsing for example)

	if config.Ascending == nil {
		log.Debugf(logPrefix + "ascending is not set. Assuming 'true'.\n")
		defaultAscending := true
		config.Ascending = &defaultAscending
	}

	return nil
}
