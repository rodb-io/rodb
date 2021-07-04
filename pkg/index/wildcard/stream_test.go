package wildcard

import (
	"io/ioutil"
	"testing"
)

func createTestStream(t *testing.T) *Stream {
	file, err := ioutil.TempFile("/tmp", "test-index")
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}
	stream := NewStream(file, 0)

	// Dummy byte to avoid issues with the offset 0
	_, err = stream.Add([]byte{0})
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	return stream
}

func TestStreamFlush(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "test-stream")
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	t.Run("normal", func(t *testing.T) {
		fileData := []byte("Hello ")
		bufferData := []byte("World!")

		if _, err = file.WriteAt(fileData, 0); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		stream := NewStream(file, int64(len(fileData)))
		stream.bufferOffset = int64(len(fileData))
		stream.buffer = append([]byte{}, bufferData...)

		if err := stream.Flush(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		got := make([]byte, len(fileData)+len(bufferData))
		if _, err := file.ReadAt(got, 0); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := 0, len(stream.buffer); got != expect {
			t.Fatalf("Expected the buffer to have len %v, got %v", expect, got)
		}

		if expect := string(fileData) + string(bufferData); string(got) != expect {
			t.Fatalf("Expected %v, got %v", expect, string(got))
		}
	})
	t.Run("empty file", func(t *testing.T) {
		bufferData := []byte("Hello World!")

		stream := NewStream(file, 0)
		stream.bufferOffset = 0
		stream.buffer = append([]byte{}, bufferData...)

		if err := stream.Flush(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		got := make([]byte, len(bufferData))
		if _, err := file.ReadAt(got, 0); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := 0, len(stream.buffer); got != expect {
			t.Fatalf("Expected the buffer to have len %v, got %v", expect, got)
		}

		if expect := string(bufferData); string(got) != expect {
			t.Fatalf("Expected %v, got %v", expect, string(got))
		}
	})
	t.Run("empty buffer", func(t *testing.T) {
		fileData := []byte("Hello World!")

		if _, err = file.WriteAt(fileData, 0); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		stream := NewStream(file, int64(len(fileData)))
		stream.bufferOffset = int64(len(fileData))
		stream.buffer = []byte{}

		if err := stream.Flush(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		got := make([]byte, len(fileData))
		if _, err := file.ReadAt(got, 0); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := 0, len(stream.buffer); got != expect {
			t.Fatalf("Expected the buffer to have len %v, got %v", expect, got)
		}

		if expect := string(fileData); string(got) != expect {
			t.Fatalf("Expected %v, got %v", expect, string(got))
		}
	})
}

func TestStreamGet(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "test-stream")
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	data := []byte("Hello World!")
	stream := NewStream(file, int64(len(data)))
	_, err = file.WriteAt(data, 0)
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	t.Run("normal", func(t *testing.T) {
		got, err := stream.Get(6, 5)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect := "World"; string(got) != expect {
			t.Fatalf("Expected %v, got %v", expect, string(got))
		}
	})
	t.Run("beginning", func(t *testing.T) {
		got, err := stream.Get(0, 3)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect := "Hel"; string(got) != expect {
			t.Fatalf("Expected %v, got %v", expect, string(got))
		}
	})
	t.Run("end", func(t *testing.T) {
		got, err := stream.Get(11, 1)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect := "!"; string(got) != expect {
			t.Fatalf("Expected %v, got %v", expect, string(got))
		}
	})
	t.Run("too long", func(t *testing.T) {
		_, err := stream.Get(11, 2)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})
	t.Run("negative offset", func(t *testing.T) {
		_, err := stream.Get(-1, 2)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})
	t.Run("from buffer", func(t *testing.T) {
		stream.bufferOffset = stream.streamSize
		stream.buffer = []byte("Hello World 2!")

		gotBytes, err := stream.Get(18, 7)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if expect, got := "World 2", string(gotBytes); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
	t.Run("from file and buffer", func(t *testing.T) {
		stream.bufferOffset = stream.streamSize
		stream.buffer = []byte("Hello World 2!")

		gotBytes, err := stream.Get(6, 11)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if expect, got := "World!Hello", string(gotBytes); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestStreamAdd(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		file, err := ioutil.TempFile("/tmp", "test-stream")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		data := []byte("Hello")
		stream := NewStream(file, int64(len(data)))
		_, err = file.WriteAt(data, 0)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		gotOffset, err := stream.Add([]byte(" World!"))
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect := int64(5); gotOffset != expect {
			t.Fatalf("Expected %v, got %v", expect, gotOffset)
		}
		if expect, got := int64(12), stream.streamSize; got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		gotFile := make([]byte, 5)
		_, err = file.ReadAt(gotFile, 0)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect := "Hello"; string(gotFile) != expect {
			t.Fatalf("Expected %v, got %v", expect, string(gotFile))
		}

		if expect, got := " World!", string(stream.buffer); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
	t.Run("exceeds buffer size", func(t *testing.T) {
		file, err := ioutil.TempFile("/tmp", "test-stream")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		data := []byte("Hello")
		stream := NewStream(file, int64(len(data)))
		stream.bufferMaxSize = 1
		_, err = file.WriteAt(data, 0)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		gotOffset, err := stream.Add([]byte(" World!"))
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect := int64(5); gotOffset != expect {
			t.Fatalf("Expected %v, got %v", expect, gotOffset)
		}
		if expect, got := int64(12), stream.streamSize; got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		gotFile := make([]byte, 12)
		_, err = file.ReadAt(gotFile, 0)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect := "Hello World!"; string(gotFile) != expect {
			t.Fatalf("Expected %v, got %v", expect, string(gotFile))
		}
	})
}

func TestStreamReplace(t *testing.T) {
	t.Run("in file", func(t *testing.T) {
		file, err := ioutil.TempFile("/tmp", "test-stream")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		data := []byte("Hello xxxxx!")
		stream := NewStream(file, int64(len(data)))
		_, err = file.WriteAt(data, 0)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if err := stream.Replace(6, []byte("World!")); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		gotFile := make([]byte, 12)
		_, err = file.ReadAt(gotFile, 0)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := int64(12), stream.streamSize; got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect := "Hello World!"; string(gotFile) != expect {
			t.Fatalf("Expected %v, got %v", expect, string(gotFile))
		}
	})
	t.Run("in buffer", func(t *testing.T) {
		file, err := ioutil.TempFile("/tmp", "test-stream")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		data := []byte("Hello xxxxx!")
		stream := NewStream(file, int64(len(data)))
		stream.buffer = data
		stream.bufferOffset = 0
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if err := stream.Replace(6, []byte("World!")); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := int64(12), stream.streamSize; got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := "Hello World!", string(stream.buffer); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
	t.Run("in file and buffer, extends buffer", func(t *testing.T) {
		file, err := ioutil.TempFile("/tmp", "test-stream")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		data := []byte("Hello xxxxx!")
		stream := NewStream(file, int64(len(data)))
		_, err = file.WriteAt(data, 0)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if err := stream.Replace(6, []byte("World!!")); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		gotFile := make([]byte, 12)
		_, err = file.ReadAt(gotFile, 0)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := int64(13), stream.streamSize; got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect := "Hello World!"; string(gotFile) != expect {
			t.Fatalf("Expected %v, got %v", expect, string(gotFile))
		}

		if expect, got := 1, len(stream.buffer); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
	t.Run("in file and buffer, replaces buffer", func(t *testing.T) {
		file, err := ioutil.TempFile("/tmp", "test-stream")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		fileData := []byte("Hello xxxxx!")
		bufferData := []byte("?")

		stream := NewStream(file, int64(len(fileData)+len(bufferData)))
		stream.bufferOffset = int64(len(fileData))
		stream.buffer = append([]byte{}, bufferData...)
		_, err = file.WriteAt(fileData, 0)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if err := stream.Replace(6, []byte("World!!")); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		gotFile := make([]byte, 12)
		_, err = file.ReadAt(gotFile, 0)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := int64(13), stream.streamSize; got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect := "Hello World!"; string(gotFile) != expect {
			t.Fatalf("Expected %v, got %v", expect, string(gotFile))
		}

		if expect, got := 1, len(stream.buffer); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
	t.Run("in file and buffer, triggers flush", func(t *testing.T) {
		file, err := ioutil.TempFile("/tmp", "test-stream")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		fileData := []byte("Hello")
		bufferData := []byte("!")

		stream := NewStream(file, int64(len(fileData)+len(bufferData)))
		stream.buffer = append([]byte{}, bufferData...)
		stream.bufferMaxSize = 1
		_, err = file.WriteAt(fileData, 0)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if err := stream.Replace(0, []byte("Hello World!")); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		gotFile := make([]byte, 12)
		_, err = file.ReadAt(gotFile, 0)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := int64(12), stream.streamSize; got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect := "Hello World!"; string(gotFile) != expect {
			t.Fatalf("Expected %v, got %v", expect, string(gotFile))
		}

		if expect, got := 0, len(stream.buffer); got != expect {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}
