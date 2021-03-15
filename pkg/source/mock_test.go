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

func TestMockWatch(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "Hello World!"
		mock := NewMock(data)
		watcher := &Watcher{
			OnChange: func() {
			},
		}

		if expect, got := 0, len(mock.watchers); got != expect {
			t.Errorf("Expected the array to contain %+v elements, got %v.", expect, got)
		}

		err := mock.Watch("test", watcher)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := 1, len(mock.watchers); got != expect {
			t.Errorf("Expected the array to contain %+v elements, got %v.", expect, got)
		}

		err = mock.CloseWatcher("test", watcher)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := 0, len(mock.watchers); got != expect {
			t.Errorf("Expected the array to contain %+v elements, got %v.", expect, got)
		}
	})
}
