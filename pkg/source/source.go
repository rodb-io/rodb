package source

import (
	"errors"
	"rods/pkg/config"
	"io"
	"github.com/sirupsen/logrus"
)

type Source interface {
	Open(filePath string) (io.ReadSeeker, error)
	Close() error
	CloseReader(reader io.ReadSeeker) error
}

type SourceList = map[string]Source

func NewFromConfig(
	config config.SourceConfig,
	log *logrus.Logger,
) (Source, error) {
	if config.Filesystem != nil {
		return NewFilesystem(config.Filesystem, log)
	}

	return nil, errors.New("Failed to initialize source")
}

func NewFromConfigs(
	configs map[string]config.SourceConfig,
	log *logrus.Logger,
) (SourceList, error) {
	sources := make(SourceList)
	for sourceName, sourceConfig := range configs {
		source, err := NewFromConfig(sourceConfig, log)
		if err != nil {
			return nil, err
		}
		sources[sourceName] = source
	}

	return sources, nil
}

func Close(sources SourceList) error {
	for _, source := range sources {
		err := source.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
