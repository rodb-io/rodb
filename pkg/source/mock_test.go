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

func TestMockWatch(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "Hello World!"
		mock := NewMock(data)

		callCount := 0
		watcher := &Watcher{
			OnChange: func() {
				callCount++
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

		if callCount != 0 {
			t.Errorf("Expected the function to not be called, got '%v'", callCount)
		}

		mock.TriggerWatchers()
		if callCount != 1 {
			t.Errorf("Expected the function to be called once, got '%v'", callCount)
		}

		callCount = 0
		err = mock.CloseWatcher("test", watcher)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := 0, len(mock.watchers); got != expect {
			t.Errorf("Expected the array to contain %+v elements, got %v.", expect, got)
		}

		mock.TriggerWatchers()
		if callCount != 0 {
			t.Errorf("Expected the function to not be called, got '%v'", callCount)
		}
	})
}
