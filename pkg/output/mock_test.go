package output

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"
)

func TestMock(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		mock := NewMock("/mock")
		_ = mock
	})
}

func TestMockEndpoint(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		mock := NewMock("/hello")

		if got, expect := mock.Endpoint().String(), "^/hello$"; got != expect {
			t.Errorf("Expected to get the endpoint '%+v', got: '%+v'", expect, got)
		}
	})
}

func TestMockExpectedPayloadType(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		expectedPayloadType := "image/png"
		mock := NewMock("/mock")
		mock.MockPayloadType = &expectedPayloadType

		if got := mock.ExpectedPayloadType(); *got != expectedPayloadType {
			t.Errorf("Expected to get the endpoint '%+v', got: '%+v'", expectedPayloadType, got)
		}
	})
}

func TestMockHandle(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "output test"
		mock := NewMock("/mock")
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
			t.Errorf("Unexpected error: '%+v'", err)
		}

		gotData, err := ioutil.ReadAll(output)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if gotErr != nil {
			t.Errorf("Handler sent an unexpected error: '%+v'", gotErr)
		}

		if string(gotData) != data {
			t.Errorf("Expected to get the data '%+v', got: '%+v'", data, string(gotData))
		}
	})
	t.Run("error", func(t *testing.T) {
		data := errors.New("test error")
		mock := NewMock("/mock")
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
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if gotErr != data {
			t.Errorf("Expected to get the error '%+v', got: '%+v'", data, gotErr)
		}
	})
}

func TestMockClose(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		mock := NewMock("/mock")

		err := mock.Close()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
	})
}
