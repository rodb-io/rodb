package util

import (
	"bufio"
	"strings"
	"testing"
)

func TestGetBufferedReaderOffset(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		reader := strings.NewReader("abcdef")
		buffer := bufio.NewReader(reader)
		buffer.Read(make([]byte, 3))
		offset, err := GetBufferedReaderOffset(reader, buffer)
		if err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}

		if expect := int64(3); offset != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, offset)
		}
	})
	t.Run("start", func(t *testing.T) {
		reader := strings.NewReader("abcdef")
		buffer := bufio.NewReader(reader)
		offset, err := GetBufferedReaderOffset(reader, buffer)
		if err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}

		if expect := int64(0); offset != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, offset)
		}
	})
}

func TestSetBufferedReaderOffset(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		reader := strings.NewReader("abcdef")
		buffer := bufio.NewReader(reader)
		buffer.Read(make([]byte, 5))
		SetBufferedReaderOffset(reader, buffer, 1)

		data := make([]byte, 2)
		buffer.Read(data)

		if expect := "bc"; string(data) != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, string(data))
		}
	})
}
