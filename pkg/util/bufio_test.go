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
		if _, err := buffer.Read(make([]byte, 3)); err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		offset, err := GetBufferedReaderOffset(reader, buffer)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		if expect := int64(3); offset != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, offset)
		}
	})
	t.Run("start", func(t *testing.T) {
		reader := strings.NewReader("abcdef")
		buffer := bufio.NewReader(reader)
		offset, err := GetBufferedReaderOffset(reader, buffer)
		if err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		if expect := int64(0); offset != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, offset)
		}
	})
}

func TestSetBufferedReaderOffset(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		reader := strings.NewReader("abcdef")
		buffer := bufio.NewReader(reader)
		if _, err := buffer.Read(make([]byte, 5)); err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		if err := SetBufferedReaderOffset(reader, buffer, 1); err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		data := make([]byte, 2)
		if _, err := buffer.Read(data); err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		if expect := "bc"; string(data) != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, string(data))
		}
	})
}
