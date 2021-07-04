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
	bytes := make([]byte, 0, bytesCount)
	remainingBytesCount := bytesCount
	currentOffset := offset

	if currentOffset < stream.bufferOffset {
		bytesToReadFromFile := int64(remainingBytesCount)
		if currentOffset+bytesToReadFromFile > stream.bufferOffset {
			bytesToReadFromFile = stream.bufferOffset - currentOffset
		}

		bytesFromFile := make([]byte, bytesToReadFromFile)
		size, err := stream.stream.ReadAt(bytesFromFile, currentOffset)
		if err != nil {
			return nil, err
		}
		if int64(size) != bytesToReadFromFile {
			return nil, fmt.Errorf("Expected to read %v bytes from file, got %v", remainingBytesCount, size)
		}

		if bytesToReadFromFile < int64(remainingBytesCount) {
			for _, currentByte := range bytesFromFile {
				bytes = append(bytes, currentByte)
			}

			// Updating values so that we can handle both cases in the next condition
			currentOffset = stream.bufferOffset
			remainingBytesCount -= int(bytesToReadFromFile)
		} else {
			return bytesFromFile, nil
		}
	}

	if currentOffset >= stream.bufferOffset {
		start := currentOffset - stream.bufferOffset
		end := start + int64(remainingBytesCount)
		if end > int64(len(stream.buffer)) {
			return nil, fmt.Errorf("Expected to read %v bytes from buffer, got %v", remainingBytesCount, int64(len(stream.buffer))-start)
		}

		if len(bytes) > 0 {
			// Adding to the previous part we got from the file
			for _, currentByte := range stream.buffer[start:end] {
				bytes = append(bytes, currentByte)
			}
		} else {
			return stream.buffer[start:end], nil
		}
	}

	if len(bytes) != bytesCount {
		return nil, fmt.Errorf("Expected to end-up with %v bytes, got %v", bytesCount, len(bytes))
	}

	return bytes, nil
}

func (stream *Stream) Add(bytes []byte) (offset int64, err error) {
	offset = stream.streamSize
	if err := stream.Replace(offset, bytes); err != nil {
		return 0, err
	}

	return offset, nil
}

func (stream *Stream) Replace(offset int64, bytes []byte) error {
	currentOffset := offset
	remainingBytes := bytes

	if currentOffset < stream.bufferOffset {
		bytesToWriteToFile := int64(len(remainingBytes))
		if currentOffset+bytesToWriteToFile > stream.bufferOffset {
			bytesToWriteToFile = stream.bufferOffset - currentOffset
		}

		size, err := stream.stream.WriteAt(remainingBytes[:bytesToWriteToFile], currentOffset)
		if err != nil {
			return err
		}
		if int64(size) != bytesToWriteToFile {
			return fmt.Errorf("Expected to write %v bytes, wrote %v", len(remainingBytes), size)
		}

		if bytesToWriteToFile < int64(len(remainingBytes)) {
			// Updating values so that we can handle both cases in the next condition
			currentOffset = stream.bufferOffset
			remainingBytes = remainingBytes[bytesToWriteToFile:]
		}
	}

	if currentOffset >= stream.bufferOffset {
		start := int(currentOffset - stream.bufferOffset)
		for i := 0; i < len(remainingBytes) && (i+start) < len(stream.buffer); i++ {
			stream.buffer[i+start] = remainingBytes[i]
		}

		remainingBytesCount := len(remainingBytes) - (len(stream.buffer) - start)
		if remainingBytesCount > 0 {
			stream.buffer = append(stream.buffer, remainingBytes[(len(remainingBytes)-remainingBytesCount):]...)
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
