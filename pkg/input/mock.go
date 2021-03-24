package input

import (
	"errors"
	"rods/pkg/record"
)

type Mock struct {
	data []IterateAllResult
}

func NewMock(data []IterateAllResult) *Mock {
	return &Mock{
		data: data,
	}
}

func (mock *Mock) Name() string {
	return "mock"
}

func (mock *Mock) HasColumn(columnName string) bool {
	if len(mock.data) == 0 {
		return false
	}

	_, err := mock.data[0].Record.Get(columnName)

	return err == nil
}

func (mock *Mock) Get(position record.Position) (record.Record, error) {
	index := int(position)
	if index >= len(mock.data) {
		return nil, errors.New("There is no mocked record matching the given position")
	}

	result := mock.data[index]
	return result.Record, result.Error
}

func (mock *Mock) Size() (int64, error) {
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
