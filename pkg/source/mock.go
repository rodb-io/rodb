package source

import (
	"io"
	"strings"
)

type Mock struct {
	data     string
	watchers []*Watcher
}

func NewMock(data string) *Mock {
	return &Mock{
		data:     data,
		watchers: make([]*Watcher, 0),
	}
}

func (mock *Mock) Open(filePath string) (io.ReadSeeker, error) {
	return strings.NewReader(mock.data), nil
}

func (mock *Mock) Watch(filePath string, watcher *Watcher) error {
	mock.watchers = append(mock.watchers, watcher)
	return nil
}

func (mock *Mock) Close() error {
	return nil
}

func (mock *Mock) CloseWatcher(filePath string, watcher *Watcher) error {
	newWatchers := make([]*Watcher, 0)
	for _, currentWatcher := range mock.watchers {
		if currentWatcher != watcher {
			newWatchers = append(newWatchers, currentWatcher)
		}
	}
	mock.watchers = newWatchers

	return nil
}

func (mock *Mock) CloseReader(reader io.ReadSeeker) error {
	return nil
}
