package input

import (
	"errors"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/record"
)

type Mock struct {
	data   []IterateAllResult
	parser parser.Parser
}

func NewMock(parser parser.Parser, data []IterateAllResult) *Mock {
	return &Mock{
		data:   data,
		parser: parser,
	}
}

func (mock *Mock) Name() string {
	return "mock"
}

func (mock *Mock) HasProperty(propertyName string) bool {
	if len(mock.data) == 0 {
		return false
	}

	_, err := mock.data[0].Record.Get(propertyName)

	return err == nil
}

func (mock *Mock) GetPropertyParser(propertyName string) parser.Parser {
	return mock.parser
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
