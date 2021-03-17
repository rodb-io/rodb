package source

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"os"
	"path/filepath"
	"rods/pkg/config"
	"sync"
)

type Filesystem struct {
	config       *config.FilesystemSource
	opened       map[io.ReadSeeker]*os.File
	watcher      *fsnotify.Watcher
	watchers     map[string][]*Watcher
	watchersLock *sync.Mutex
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

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fs := &Filesystem{
		config:       config,
		opened:       make(map[io.ReadSeeker]*os.File),
		watcher:      watcher,
		watchers:     make(map[string][]*Watcher),
		watchersLock: &sync.Mutex{},
	}

	fs.startWatchProcess()

	return fs, nil
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
					fs.watchersLock.Lock()

					watchers, watchersArrayExists := fs.watchers[event.Name]
					if watchersArrayExists {
						for _, watcher := range watchers {
							watcher.OnChange()
						}
					} else {
						fs.config.Logger.Warnf("Received watch event '%v', but no watcher found.", event.String())
					}

					fs.watchersLock.Unlock()
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

func (fs *Filesystem) getFilePath(filePath string) string {
	return filepath.Join(fs.config.Path, filePath)
}

func (fs *Filesystem) Open(filePath string) (io.ReadSeeker, error) {
	path := fs.getFilePath(filePath)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader := io.ReadSeeker(file)
	fs.opened[reader] = file

	return reader, nil
}

func (fs *Filesystem) Size(filePath string) (int64, error) {
	path := fs.getFilePath(filePath)
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}

func (fs *Filesystem) Watch(filePath string, watcher *Watcher) error {
	fs.watchersLock.Lock()
	defer fs.watchersLock.Unlock()

	path := fs.getFilePath(filePath)
	if watchers, watchersArrayExists := fs.watchers[path]; watchersArrayExists {
		watchers = append(watchers, watcher)
		fs.watchers[path] = watchers
	} else {
		watchers = []*Watcher{watcher}
		fs.watchers[path] = watchers

		err := fs.watcher.Add(path)
		if err != nil {
			return err
		}
	}

	return nil
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

func (fs *Filesystem) CloseWatcher(filePath string, watcher *Watcher) error {
	fs.watchersLock.Lock()
	defer fs.watchersLock.Unlock()

	path := fs.getFilePath(filePath)
	watchers, watchersArrayExists := fs.watchers[path]
	if !watchersArrayExists {
		return errors.New("Trying to close a non-opened filesystem watcher.")
	}

	newWatchers := make([]*Watcher, 0)
	for _, currentWatcher := range watchers {
		if currentWatcher != watcher {
			newWatchers = append(newWatchers, currentWatcher)
		}
	}

	if len(newWatchers) == 0 {
		delete(fs.watchers, path)
		err := fs.watcher.Remove(path)
		if err != nil {
			return err
		}
	} else {
		fs.watchers[path] = newWatchers
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
