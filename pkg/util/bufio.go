package util

import (
	"bufio"
	"io"
)

func GetBufferedReaderOffset(
	reader io.ReadSeeker,
	buffer *bufio.Reader,
) (int64, error) {
	offset, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	bufferSize := int64(buffer.Buffered())

	return offset - bufferSize, nil
}

func SetBufferedReaderOffset(
	reader io.ReadSeeker,
	buffer *bufio.Reader,
	offset int64,
) {
	reader.Seek(offset, io.SeekStart)
	buffer.Reset(reader)
}
