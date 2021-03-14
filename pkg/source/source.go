package source

import (
	"errors"
	"io"
	"rods/pkg/config"
)

type Source interface {
	Open(filePath string) (io.ReadSeeker, error)
	Close() error
	CloseReader(reader io.ReadSeeker) error
}

type List = map[string]Source

func NewFromConfig(
	config config.Source,
) (Source, error) {
	if config.Filesystem != nil {
		return NewFilesystem(config.Filesystem)
	}

	return nil, errors.New("Failed to initialize source")
}

func NewFromConfigs(
	configs map[string]config.Source,
) (List, error) {
	sources := make(List)
	for sourceName, sourceConfig := range configs {
		source, err := NewFromConfig(sourceConfig)
		if err != nil {
			return nil, err
		}
		sources[sourceName] = source
	}

	return sources, nil
}

func Close(sources List) error {
	for _, source := range sources {
		err := source.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
