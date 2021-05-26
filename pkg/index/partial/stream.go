package partial

import (
	"fmt"
	"io"
)

type ReaderAtWriterAt interface{
	io.ReaderAt
	io.WriterAt
}

type Stream struct{
	stream ReaderAtWriterAt
	streamSize int64
}

func NewStream(
	stream ReaderAtWriterAt,
	streamSize int64,
) *Stream {
	return &Stream{
		stream: stream,
		streamSize: streamSize,
	}
}

func (stream *Stream) Get(offset int64, bytesCount int) ([]byte, error) {
	bytes := make([]byte, 0, bytesCount)
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
	newOffset := stream.streamSize
	size, err := stream.stream.WriteAt(bytes, newOffset)
	if err != nil {
		return 0, err
	}
	if size != len(bytes) {
		return 0, fmt.Errorf("Expected to write %v bytes, wrote %v", len(bytes), size)
	}

	stream.streamSize += int64(len(bytes))
	return newOffset, nil
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
