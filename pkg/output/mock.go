package output

import (
	"io"
	"github.com/rodb-io/rodb/pkg/parser"
)

type Mock struct {
	MockOutput      func(params map[string]string) ([]byte, error)
	MockPayloadType *string
	parser          parser.Parser
}

func NewMock(
	parser parser.Parser,
) *Mock {
	return &Mock{
		parser: parser,
	}
}

func (mock *Mock) ExpectedPayloadType() *string {
	return mock.MockPayloadType
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
	data, err := mock.MockOutput(params)
	if err != nil {
		return sendError(err)
	}

	_, err = sendSucces().Write(data)
	return err
}

func (mock *Mock) Name() string {
	return "mock"
}

func (mock *Mock) HasParameter(paramName string) bool {
	return true
}

func (mock *Mock) GetParameterParser(paramName string) (parser.Parser, error) {
	return mock.parser, nil
}

func (mock *Mock) Close() error {
	return nil
}
