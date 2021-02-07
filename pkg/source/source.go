package source

import (
	"rods/pkg/config"
	"io"
)

type Source interface {
	Open(filePath string) (io.ReadSeeker, error)
}

func NewFromConfig(config config.SourceConfig) Source {
	if config.Filesystem != nil {
		return NewFilesystem(config.Filesystem)
	}

	return nil
}

func NewFromConfigs(configs map[string]config.SourceConfig) map[string]Source {
	sources := make(map[string]Source)
	for sourceName, sourceConfig := range configs {
		sources[sourceName] = NewFromConfig(sourceConfig)
	}

	return sources
}
