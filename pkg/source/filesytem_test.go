package source

import (
	"io/ioutil"
	"os"
	"rods/pkg/config"
	"sync"
	"testing"
)

func TestFilesystemOpen(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		path := t.TempDir()
		fileName := "testOpen"
		data := "Hello World!"

		file, err := os.Create(path + "/" + fileName)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		_, err = file.WriteString(data)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		fs, err := NewFilesystem(&config.FilesystemSource{
			Path: path,
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		reader, err := fs.Open(fileName)
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

func TestFilesystemWatch(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		path := t.TempDir()
		fileName := "testWatch"

		file, err := os.Create(path + "/" + fileName)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		_, err = file.WriteString("initial content")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		fs, err := NewFilesystem(&config.FilesystemSource{
			Path: path,
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		onChangeWaiter := &sync.WaitGroup{}
		callCount := 0
		watcher := &Watcher{
			OnChange: func() {
				callCount++
				onChangeWaiter.Done()
			},
		}
		err = fs.Watch(fileName, watcher)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		onChangeWaiter.Add(1)
		_, err = file.WriteString("changed content")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		onChangeWaiter.Wait()
		if callCount != 1 {
			t.Errorf("Expected the function to be called once, got '%v'", callCount)
		}

		err = fs.CloseWatcher(fileName, watcher)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		callCount = 0
		onChangeWaiter.Add(1)
		_, err = file.WriteString("changed content again")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		// Closing to test the gorouting end and delay the next assertion
		err = fs.Close()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if callCount != 0 {
			t.Errorf("Expected the function to not be called, got '%v'", callCount)
		}
	})
}

func TestFilesystemGetFilePath(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		path := "/tmp"
		fileName := "file"

		fs, err := NewFilesystem(&config.FilesystemSource{
			Path: path,
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := "/tmp/file", fs.getFilePath(fileName); expect != got {
		}
	})
}
