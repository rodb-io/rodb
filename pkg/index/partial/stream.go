package partial

import (
	"fmt"
	"io"
	"os"
)

type Stream struct {
	stream     *os.File
	streamSize int64
}

func NewStream(
	stream *os.File,
	streamSize int64,
) *Stream {
	return &Stream{
		stream:     stream,
		streamSize: streamSize,
	}
}

func (stream *Stream) Get(offset int64, bytesCount int) ([]byte, error) {
	bytes := make([]byte, bytesCount)
	size, err := stream.stream.ReadAt(bytes, offset)
	if err != nil {
		return nil, err
	}
	if size != bytesCount {
		return nil, fmt.Errorf("Expected to read %v bytes, got %v", bytesCount, size)
	}

	return bytes, nil
}

func (stream *Stream) Add(bytes []byte) (offset int64, err error) {
	offset = stream.streamSize
	if err = stream.Replace(offset, bytes); err != nil {
		return 0, err
	}

	return offset, nil
}

func (stream *Stream) Replace(offset int64, bytes []byte) error {
	size, err := stream.stream.WriteAt(bytes, offset)
	if err != nil {
		return err
	}
	if size != len(bytes) {
		return fmt.Errorf("Expected to write %v bytes, wrote %v", len(bytes), size)
	}

	endOffset := offset + int64(len(bytes))
	if endOffset > stream.streamSize {
		stream.streamSize = endOffset
	}

	return nil
}

// Returns a reader from the given position
// Note: the current implementation is dumb and does not
// work with concurrent read or writes
func (stream *Stream) GetReaderFrom(offset int64) (io.Reader, error) {
	_, err := stream.stream.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return stream.stream, nil
}
