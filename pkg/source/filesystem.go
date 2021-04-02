package source

import (
	"errors"
	"io"
	"os"
	"rods/pkg/config"
	"sync"
)

type Filesystem struct {
	config             *config.FilesystemSource
	opened             map[io.ReadSeeker]*os.File
	openedWatchCounter map[string]int
	openedLock         *sync.Mutex
}

func NewFilesystem(
	config *config.FilesystemSource,
) (*Filesystem, error) {
	fs := &Filesystem{
		config:             config,
		opened:             make(map[io.ReadSeeker]*os.File),
		openedWatchCounter: map[string]int{},
		openedLock:         &sync.Mutex{},
	}

	return fs, nil
}

func (fs *Filesystem) Name() string {
	return fs.config.Name
}

func (fs *Filesystem) Open(filePath string) (io.ReadSeeker, error) {
	fs.openedLock.Lock()
	defer fs.openedLock.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	reader := io.ReadSeeker(file)
	fs.opened[reader] = file

	return reader, nil
}

func (fs *Filesystem) Size(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
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
	fs.openedLock.Lock()
	defer fs.openedLock.Unlock()

	file, exists := fs.opened[reader]
	if !exists {
		return errors.New("Trying to close a non-opened filesystem source.")
	}

	delete(fs.opened, reader)

	return file.Close()
}
