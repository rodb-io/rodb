package source

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"rods/pkg/config"
)

type Filesystem struct{
	config *config.FilesystemSourceConfig
	opened map[io.ReadSeeker]*os.File
}

func NewFilesystem(config *config.FilesystemSourceConfig) (*Filesystem, error) {
	return &Filesystem{
		config: config,
	}, nil
}

func (fs *Filesystem) Open(filePath string) (io.ReadSeeker, error) {
	path := filepath.Join(fs.config.Path, filePath)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader := io.ReadSeeker(file)
	fs.opened[reader] = file

	return reader, nil
}

func (fs *Filesystem) Close() error {
	for _, file := range fs.opened {
		err := file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *Filesystem) CloseReader(reader io.ReadSeeker) error {
	file, exists := fs.opened[reader]
	if !exists {
		return errors.New("Trying to close a non-opened filesystem source.")
	}

	delete(fs.opened, reader)

	return file.Close()
}
