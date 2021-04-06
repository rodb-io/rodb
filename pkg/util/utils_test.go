package util

import (
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"sync"
	"testing"
)

func TestRemoveCharacters(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		if got, expect := RemoveCharacters("abcdef", "db"), "acef"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("nothing to replace", func(t *testing.T) {
		if got, expect := RemoveCharacters("abcdef", "ghi"), "abcdef"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("unicode character", func(t *testing.T) {
		if got, expect := RemoveCharacters("あいうえお", "うお"), "あいえ"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
}

func TestIsInArray(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if result := IsInArray("string", []string{"a", "string"}); !result {
			t.Fail()
		}
	})
	t.Run("invalid", func(t *testing.T) {
		if result := IsInArray("invalid", []string{"string"}); result {
			t.Fail()
		}
	})
	t.Run("empty value", func(t *testing.T) {
		if result := IsInArray("", []string{"string"}); result {
			t.Fail()
		}
	})
	t.Run("empty array", func(t *testing.T) {
		if result := IsInArray("string", []string{}); result {
			t.Fail()
		}
	})
}

func TestGetAddress(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		if got, expect := GetAddress(&net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 123}), "127.0.0.1:123"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'.", expect, got)
		}
		if got, expect := GetAddress(&net.TCPAddr{IP: net.IPv4(1, 2, 3, 4), Port: 123}), "1.2.3.4:123"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'.", expect, got)
		}
		if got, expect := GetAddress(&net.TCPAddr{IP: net.IPv4(100, 0, 0, 0), Port: 123}), "100.0.0.0:123"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'.", expect, got)
		}
		if got, expect := GetAddress(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 123}), "127.0.0.1:123"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'.", expect, got)
		}
	})
}

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
