package config

import (
	"errors"
)

type IndexConfig struct{
	MemoryMapUnique *MemoryMapUniqueIndexConfig
	MemoryMapMultiple *MemoryMapMultipleIndexConfig
}

func (config *IndexConfig) validate() error {
	fields := getAllNonNilFields(config)

	if len(fields) == 0 {
		return errors.New("All indexes must have a configuration")
	}

	if len(fields) > 0 {
		return errors.New("An index can only have one configuration")
	}

	return fields[0].validate()
}
