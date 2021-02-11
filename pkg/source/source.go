package source

import (
	"errors"
	"github.com/sirupsen/logrus"
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
	log *logrus.Logger,
) (Source, error) {
	if config.Filesystem != nil {
		return NewFilesystem(config.Filesystem, log)
	}

	return nil, errors.New("Failed to initialize source")
}

func NewFromConfigs(
	configs map[string]config.Source,
	log *logrus.Logger,
) (List, error) {
	sources := make(List)
	for sourceName, sourceConfig := range configs {
		source, err := NewFromConfig(sourceConfig, log)
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
