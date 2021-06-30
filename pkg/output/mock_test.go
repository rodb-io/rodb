package output

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"rodb.io/pkg/parser"
	"testing"
)

func TestMock(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		parser := parser.NewMock()
		mock := NewMock(parser)
		_ = mock
	})
}

func TestMockExpectedPayloadType(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		expectedPayloadType := "image/png"
		parser := parser.NewMock()
		mock := NewMock(parser)
		mock.MockPayloadType = &expectedPayloadType

		if got := mock.ExpectedPayloadType(); *got != expectedPayloadType {
			t.Fatalf("Expected to get the payload type '%+v', got: '%+v'", expectedPayloadType, got)
		}
	})
}

func TestMockHandle(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "output test"
		parser := parser.NewMock()
		mock := NewMock(parser)
		mock.MockOutput = func(params map[string]string) ([]byte, error) {
			return []byte(data), nil
		}

		var gotErr error = nil
		output := bytes.NewBufferString("")
		err := mock.Handle(
			map[string]string{},
			[]byte{},
			func(err error) error {
				gotErr = err
				return nil
			},
			func() io.Writer {
				return output
			},
		)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		gotData, err := ioutil.ReadAll(output)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if gotErr != nil {
			t.Fatalf("Handler sent an unexpected error: '%+v'", gotErr)
		}

		if string(gotData) != data {
			t.Fatalf("Expected to get the data '%+v', got: '%+v'", data, string(gotData))
		}
	})
	t.Run("error", func(t *testing.T) {
		data := errors.New("test error")
		parser := parser.NewMock()
		mock := NewMock(parser)
		mock.MockOutput = func(params map[string]string) ([]byte, error) {
			return nil, data
		}

		var gotErr error = nil
		output := bytes.NewBufferString("")
		err := mock.Handle(
			map[string]string{},
			[]byte{},
			func(err error) error {
				gotErr = err
				return nil
			},
			func() io.Writer {
				return output
			},
		)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if gotErr != data {
			t.Fatalf("Expected to get the error '%+v', got: '%+v'", data, gotErr)
		}
	})
}

func TestMockClose(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		parser := parser.NewMock()
		mock := NewMock(parser)

		if err := mock.Close(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
	})
}
