package source

import (
	"github.com/sirupsen/logrus"
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

func TestFilesystemSize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		path := t.TempDir()
		fileName := "testSize"
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

		size, err := fs.Size(fileName)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if size != int64(len(data)) {
			t.Errorf("Expected to get a size of '%v', got '%+v'", len(data), size)
		}
	})
}

func TestFilesystemWatch(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		dir := t.TempDir()
		fileName := "testWatch"
		path := dir + "/" + fileName

		file, err := os.Create(path)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		_, err = file.WriteString("initial content")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		trueValue := true
		fs, err := NewFilesystem(&config.FilesystemSource{
			Logger:           logrus.NewEntry(logrus.StandardLogger()),
			Path:             dir,
			DieOnInputChange: &trueValue,
		})
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		dieWaiter := &sync.WaitGroup{}
		dieCount := 0
		fs.config.Logger.Logger.ExitFunc = func(exitCode int) {
			dieCount++
			dieWaiter.Done()
		}

		reader, err := fs.Open(fileName)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		dieWaiter.Add(1)
		_, err = file.WriteString("changed content")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		dieWaiter.Wait()
		if dieCount <= 0 {
			t.Errorf("Expected the process to exit, got '%v' calls to Exit", dieCount)
		}

		err = fs.CloseReader(reader)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		dieCount = 0
		dieWaiter.Add(1)
		_, err = file.WriteString("changed content again")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if dieCount != 0 {
			t.Errorf("Expected the process not to exit, got '%v' calls to Exit", dieCount)
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
