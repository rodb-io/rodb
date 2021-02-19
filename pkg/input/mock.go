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

func (mock *Mock) Get(position record.Position) (record.Record, error) {
	index := int(position)
	if index >= len(mock.data) {
		return nil, errors.New("There is no mocked record matching the given position")
	}

	result := mock.data[index]
	return result.Record, result.Error
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
