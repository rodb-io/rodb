package source

import (
	"errors"
	"rods/pkg/config"
	"io"
)

type Source interface {
	Open(filePath string) (io.ReadSeeker, error)
}

type SourceList = map[string]Source

func NewFromConfig(config config.SourceConfig) (Source, error) {
	if config.Filesystem != nil {
		return NewFilesystem(config.Filesystem)
	}

	return nil, errors.New("Failed to initialize source")
}

func NewFromConfigs(configs map[string]config.SourceConfig) (SourceList, error) {
	sources := make(SourceList)
	for sourceName, sourceConfig := range configs {
		source, err := NewFromConfig(sourceConfig)
		if err != nil {
			return nil, err
		}
		sources[sourceName] = source
	}

	return sources, nil
}
