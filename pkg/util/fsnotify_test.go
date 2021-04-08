package util

import (
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"testing"
)

func TestStartFilesystemWatchProcess(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		path := t.TempDir()
		fileName := "testOpen"

		file, err := os.Create(path + "/" + fileName)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		_, err = file.WriteString("initial content")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		logger := logrus.NewEntry(logrus.StandardLogger())

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		StartFilesystemWatchProcess(watcher, true, logger)

		err = watcher.Add(file.Name())
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		dieWaiter := &sync.WaitGroup{}
		dieCount := 0
		logger.Logger.ExitFunc = func(exitCode int) {
			dieCount++
			dieWaiter.Done()
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

		err = watcher.Close()
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
