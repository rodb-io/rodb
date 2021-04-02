package source

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
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
	watcher            *fsnotify.Watcher
}

func NewFilesystem(
	config *config.FilesystemSource,
) (*Filesystem, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fs := &Filesystem{
		config:             config,
		opened:             make(map[io.ReadSeeker]*os.File),
		openedWatchCounter: map[string]int{},
		openedLock:         &sync.Mutex{},
		watcher:            watcher,
	}

	fs.startWatchProcess()

	return fs, nil
}

func (fs *Filesystem) Name() string {
	return fs.config.Name
}

func (fs *Filesystem) startWatchProcess() {
	go func() {
		for {
			select {
			case event, ok := <-fs.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					message := fmt.Sprintf("The file '%v' has been modified by another process", event.Name)
					if *fs.config.DieOnInputChange {
						fs.config.Logger.Fatalln(message + ". Quitting because it may have corrupted data and 'dieOnInputChange' is 'true'.")
					} else {
						fs.config.Logger.Warnln(message + ", but 'dieOnInputChange' is 'false'. This could have unpredictable consequences.")
					}
				}
			case err, ok := <-fs.watcher.Errors:
				if !ok {
					return
				}
				fs.config.Logger.Errorf("Error while watching file: %v", err)
			}
		}
	}()
}

func (fs *Filesystem) Open(filePath string) (io.ReadSeeker, error) {
	fs.openedLock.Lock()
	defer fs.openedLock.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	err = fs.watcher.Add(filePath)
	if err != nil {
		return nil, err
	}

	if openedWatchCounter, counterExists := fs.openedWatchCounter[filePath]; counterExists {
		fs.openedWatchCounter[filePath] = openedWatchCounter + 1
	} else {
		fs.openedWatchCounter[filePath] = 1
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

	err := fs.watcher.Close()
	if err != nil {
		return err
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

	path := file.Name()
	if openedWatchCounter, counterExists := fs.openedWatchCounter[path]; counterExists {
		if openedWatchCounter <= 1 {
			delete(fs.openedWatchCounter, path)
			err := fs.watcher.Remove(file.Name())
			if err != nil {
				return err
			}
		} else {
			fs.openedWatchCounter[path] = openedWatchCounter - 1
		}
	} else {
		return errors.New("Trying to remove a non-added filesystem watcher.")
	}

	delete(fs.opened, reader)

	return file.Close()
}
