package partial

import (
	"io/ioutil"
	"testing"
)

func createTestStream(t *testing.T) *Stream {
	file, err := ioutil.TempFile("/tmp", "test-index")
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}
	stream := NewStream(file, 0)

	// Dummy byte to avoid issues with the offset 0
	_, err = stream.Add([]byte{0})
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}

	return stream
}

func TestStreamGet(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "test-stream")
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}

	data := []byte("Hello World!")
	stream := NewStream(file, int64(len(data)))
	_, err = file.WriteAt(data, 0)
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}

	t.Run("normal", func(t *testing.T) {
		got, err := stream.Get(6, 5)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if expect := "World"; string(got) != expect {
			t.Errorf("Expected %v, got %v", expect, string(got))
		}
	})
	t.Run("beginning", func(t *testing.T) {
		got, err := stream.Get(0, 3)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if expect := "Hel"; string(got) != expect {
			t.Errorf("Expected %v, got %v", expect, string(got))
		}
	})
	t.Run("end", func(t *testing.T) {
		got, err := stream.Get(11, 1)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if expect := "!"; string(got) != expect {
			t.Errorf("Expected %v, got %v", expect, string(got))
		}
	})
	t.Run("too long", func(t *testing.T) {
		_, err := stream.Get(11, 2)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
	t.Run("negative offset", func(t *testing.T) {
		_, err := stream.Get(-1, 2)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestStreamAdd(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "test-stream")
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}

	data := []byte("Hello")
	stream := NewStream(file, int64(len(data)))
	_, err = file.WriteAt(data, 0)
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}

	t.Run("normal", func(t *testing.T) {
		gotOffset, err := stream.Add([]byte(" World!"))
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect := int64(5); gotOffset != expect {
			t.Errorf("Expected %v, got %v", expect, gotOffset)
		}

		gotFile := make([]byte, 12)
		_, err = file.ReadAt(gotFile, 0)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := int64(12), stream.streamSize; got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}
		if expect := "Hello World!"; string(gotFile) != expect {
			t.Errorf("Expected %v, got %v", expect, string(gotFile))
		}
	})
}

func TestStreamReplace(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "test-stream")
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}

	data := []byte("Hello xxxxx!")
	stream := NewStream(file, int64(len(data)))
	_, err = file.WriteAt(data, 0)
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}

	t.Run("normal", func(t *testing.T) {
		err := stream.Replace(6, []byte("World!"))
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		gotFile := make([]byte, 12)
		_, err = file.ReadAt(gotFile, 0)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := int64(12), stream.streamSize; got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}
		if expect := "Hello World!"; string(gotFile) != expect {
			t.Errorf("Expected %v, got %v", expect, string(gotFile))
		}
	})
	t.Run("longer", func(t *testing.T) {
		err := stream.Replace(6, []byte("World!!"))
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		gotFile := make([]byte, 13)
		_, err = file.ReadAt(gotFile, 0)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := int64(13), stream.streamSize; got != expect {
			t.Errorf("Expected %v, got %v", expect, got)
		}
		if expect := "Hello World!!"; string(gotFile) != expect {
			t.Errorf("Expected %v, got %v", expect, string(gotFile))
		}
	})
}
