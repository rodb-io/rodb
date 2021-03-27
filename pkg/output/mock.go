package output

import (
	"io"
	"regexp"
)

type Mock struct {
	endpoint *regexp.Regexp
}

func NewMock() *Mock {
	mock := &Mock{
		endpoint: regexp.MustCompile("^/mock$"),
	}

	return mock
}

func (mock *Mock) Endpoint() *regexp.Regexp {
	return mock.endpoint
}

func (mock *Mock) ExpectedPayloadType() *string {
	return nil
}

func (mock *Mock) ResponseType() string {
	return "text/plain"
}

func (mock *Mock) Handle(
	params map[string]string,
	payload []byte,
	sendError func(err error) error,
	sendSucces func() io.Writer,
) error {
	return nil
}

func (mock *Mock) Name() string {
	return "mock"
}

func (mock *Mock) Close() error {
	return nil
}
