package config

import (
	"github.com/sirupsen/logrus"
)

type MemoryMapUniqueIndexConfig struct{
	Input string
	Column string
}

func (config *MemoryMapUniqueIndexConfig) validate(log *logrus.Logger) error {
	// The input and column will be validated at runtime
	return nil
}
