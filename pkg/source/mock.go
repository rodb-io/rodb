package source

import (
	"io"
	"strings"
)

type Mock struct {
	data string
}

func NewMock(data string) *Mock {
	return &Mock{
		data: data,
	}
}

func (mock *Mock) Open(filePath string) (io.ReadSeeker, error) {
	return strings.NewReader(mock.data), nil
}

func (mock *Mock) Close() error {
	return nil
}

func (mock *Mock) CloseReader(reader io.ReadSeeker) error {
	return nil
}
