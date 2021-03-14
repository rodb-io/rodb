package source

import (
	"io/ioutil"
	"os"
	"rods/pkg/config"
	"testing"
)

func TestFilesystemOpen(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		path := "/tmp"
		fileName := "rods_TestFilesystemOpen"
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
