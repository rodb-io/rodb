package output

import (
	"regexp"
)

type Mock struct {
}

func NewMock() *Mock {
	mock := &Mock{}

	return mock
}

func (mock *Mock) Endpoint() *regexp.Regexp {
	return regexp.MustCompile("^/mock$")
}

func (mock *Mock) ExpectedPayloadType() *string {
	return nil
}

func (mock *Mock) ResponseType() string {
	return "text/plain"
}

func (mock *Mock) Name() string {
	return "mock"
}

func (mock *Mock) Close() error {
	return nil
}
