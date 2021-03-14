package source

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"rods/pkg/config"
)

type Filesystem struct {
	config *config.FilesystemSource
	opened map[io.ReadSeeker]*os.File
}

func NewFilesystem(
	config *config.FilesystemSource,
) (*Filesystem, error) {
	pathStat, err := os.Stat(config.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("The base path '%v' of the filesystem object does not exist.", config.Path)
		} else {
			return nil, err
		}
	}
	if !pathStat.IsDir() {
		return nil, fmt.Errorf("The base path '%v' of the filesystem object is not a directory.", config.Path)
	}

	return &Filesystem{
		config: config,
		opened: make(map[io.ReadSeeker]*os.File),
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
