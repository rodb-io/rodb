package wildcard

import (
	"fmt"
	"io"
	"os"
)

const STREAM_BUFFER_SIZE = 10_000

type Stream struct {
	stream        *os.File
	streamSize    int64
	bufferOffset  int64
	bufferMaxSize int64
	buffer        []byte
	readerOffset  int64
}

func NewStream(
	stream *os.File,
	streamSize int64,
) *Stream {
	return &Stream{
		stream:        stream,
		streamSize:    streamSize,
		bufferOffset:  streamSize,
		bufferMaxSize: STREAM_BUFFER_SIZE,
		buffer:        make([]byte, 0, STREAM_BUFFER_SIZE),
	}
}

// Forces the internal buffer to be written to the file
func (stream *Stream) Flush() error {
	if len(stream.buffer) == 0 {
		return nil
	}

	size, err := stream.stream.WriteAt(stream.buffer, stream.bufferOffset)
	if err != nil {
		return err
	}
	if size != len(stream.buffer) {
		return fmt.Errorf("Expected to write %v bytes, wrote %v", len(stream.buffer), size)
	}

	stream.bufferOffset = stream.streamSize
	stream.buffer = stream.buffer[:0]

	return nil
}

func (stream *Stream) Get(offset int64, bytesCount int) ([]byte, error) {
	if offset < stream.bufferOffset && (offset+int64(bytesCount)) > stream.bufferOffset {
		// Since it would make the process way more complex,
		// and this case is not expected at all, we just flush
		// the buffer if it happens
		if err := stream.Flush(); err != nil {
			return nil, err
		}
	}

	if offset < stream.bufferOffset {
		bytes := make([]byte, bytesCount)
		size, err := stream.stream.ReadAt(bytes, offset)
		if err != nil {
			return nil, err
		}
		if size != bytesCount {
			return nil, fmt.Errorf("Expected to read %v bytes, got %v", bytesCount, size)
		}

		return bytes, nil
	} else {
		start := offset - stream.bufferOffset
		end := start + int64(bytesCount)
		if end > int64(len(stream.buffer)) {
			return nil, fmt.Errorf("Expected to read %v bytes, got %v", bytesCount, int64(len(stream.buffer))-start)
		}

		return stream.buffer[start:end], nil
	}
}

func (stream *Stream) Add(bytes []byte) (offset int64, err error) {
	offset = stream.streamSize
	if err := stream.Replace(offset, bytes); err != nil {
		return 0, err
	}

	return offset, nil
}

func (stream *Stream) Replace(offset int64, bytes []byte) error {
	if offset < stream.bufferOffset && (offset+int64(len(bytes))) > stream.bufferOffset {
		// Since it would make the process way more complex,
		// and this case is not expected at all, we just flush
		// the buffer if it happens
		if err := stream.Flush(); err != nil {
			return err
		}
	}

	// TODO will bug if replace a longer string that overflows in the buffer

	if offset < stream.bufferOffset {
		size, err := stream.stream.WriteAt(bytes, offset)
		if err != nil {
			return err
		}
		if size != len(bytes) {
			return fmt.Errorf("Expected to write %v bytes, wrote %v", len(bytes), size)
		}
	} else {
		start := int(offset - stream.bufferOffset)
		for i := 0; i < len(bytes) && (i+start) < len(stream.buffer); i++ {
			stream.buffer[i+start] = bytes[i]
		}

		remainingBytes := len(bytes) - (len(stream.buffer) - start)
		if remainingBytes > 0 {
			stream.buffer = append(stream.buffer, bytes[(len(bytes)-remainingBytes):]...)
		}

		if int64(len(stream.buffer)) > stream.bufferMaxSize {
			if err := stream.Flush(); err != nil {
				return err
			}
		}
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
	// TODO
	// We just assume there is no concurrent read or write to this stream
	// So by we can return the raw stream because nothing will be buffered
	if err := stream.Flush(); err != nil {
		return nil, err
	}

	if _, err := stream.stream.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	return stream.stream, nil
}
