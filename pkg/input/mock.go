package input

import (
	"errors"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/record"
)

type Mock struct {
	data   []record.Record
	parser parser.Parser
}

func NewMock(parser parser.Parser, data []record.Record) *Mock {
	return &Mock{
		data:   data,
		parser: parser,
	}
}

func (mock *Mock) Name() string {
	return "mock"
}

func (mock *Mock) Get(position record.Position) (record.Record, error) {
	index := int(position)
	if index >= len(mock.data) {
		return nil, errors.New("There is no mocked record matching the given position")
	}

	result := mock.data[index]
	return result, nil
}

func (mock *Mock) Size() (int64, error) {
	return int64(len(mock.data)), nil
}

func (mock *Mock) IterateAll() (record.Iterator, func() error, error) {
	i := 0
	iterator := func() (record.Record, error) {
		for i < len(mock.data) {
			record := mock.data[i]
			i++
			return record, nil
		}

		return nil, nil
	}

	end := func() error {
		return nil
	}

	return iterator, end, nil
}

func (mock *Mock) Close() error {
	return nil
}
