package input

import (
	"errors"
	"rods/pkg/record"
	"rods/pkg/source"
)

type Mock struct {
	data    []IterateAllResult
	watcher *source.Watcher
}

func NewMock(data []IterateAllResult) *Mock {
	return &Mock{
		data:    data,
		watcher: nil,
	}
}

func (mock *Mock) Get(position record.Position) (record.Record, error) {
	index := int(position)
	if index >= len(mock.data) {
		return nil, errors.New("There is no mocked record matching the given position")
	}

	result := mock.data[index]
	return result.Record, result.Error
}

func (mock *Mock) Size(filePath string) (int64, error) {
	return int64(len(mock.data)), nil
}

func (mock *Mock) IterateAll() <-chan IterateAllResult {
	channel := make(chan IterateAllResult)

	go func() {
		defer close(channel)

		for _, row := range mock.data {
			channel <- row
		}
	}()

	return channel
}

func (mock *Mock) Close() error {
	return nil
}

func (mock *Mock) Watch(watcher *source.Watcher) error {
	mock.watcher = watcher
	return nil
}

func (mock *Mock) TriggerWatcher() {
	if mock.watcher != nil {
		mock.watcher.OnChange()
	}
}

func (mock *Mock) CloseWatcher(watcher *source.Watcher) error {
	mock.watcher = nil
	return nil
}
