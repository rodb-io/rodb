package source

import (
	"io/ioutil"
	"testing"
)

func TestMockOpen(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "Hello World!"
		mock := NewMock(data)
		reader, err := mock.Open("test")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		content, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if string(content) != data {
			t.Errorf("Expected to receive '%v', got '%+v'", data, string(content))
		}
	})
}

func TestMockSize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "Hello World!"
		mock := NewMock(data)
		size, err := mock.Size("test")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if size != int64(len(data)) {
			t.Errorf("Expected to get a size of '%v', got '%+v'", len(data), size)
		}
	})
}
