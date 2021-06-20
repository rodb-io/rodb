package partial

import (
	"bytes"
	"fmt"
	"testing"
	"testing/fstest"
	"time"
)

func TestMetadataNewMetadata(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		stream := createTestStream(t)

		file := &fstest.MapFile{
			Data:    make([]byte, 42),
			ModTime: time.Unix(1234, 0),
		}
		fileInfo, err := fstest.MapFS{"file": file}.Stat("file")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		metadata, err := NewMetadata(stream, MetadataInput{
			InputFileStats: fileInfo,
			IgnoreCase:     true,
			RootNodesCount: 3,
		})
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		expectBytes := append([]byte(ExpectedMagicBytes), []byte{
			0, 0x1, // version
			0, 0, 0, 0, 0, 0, 0x4, 0xD2, // inputFileModificationTime
			0, 0, 0, 0, 0, 0, 0, 0x2A, // inputFileSize
			1,                        // ignoreCase
			0, 0, 0, 0, 0, 0, 0, 0x3, // rootNodeOffsetCount
			0, 0, 0, 0, 0, 0, 0, 0, // rootNodeOffsets[0]
			0, 0, 0, 0, 0, 0, 0, 0, // rootNodeOffsets[1]
			0, 0, 0, 0, 0, 0, 0, 0, // rootNodeOffsets[2]
		}...)
		gotBytes, err := stream.Get(0, len(expectBytes))
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if expect, got := fmt.Sprintf("%x", expectBytes), fmt.Sprintf("%x", gotBytes); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		if expect, got := ExpectedMagicBytes, string(metadata.magicBytes); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := CurrentVersion, metadata.version; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(1234), metadata.inputFileModificationTime.Unix(); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(42), metadata.inputFileSize; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := true, metadata.ignoreCase; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		if expect, got := 3, len(metadata.rootNodeOffsets); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(0), metadata.rootNodeOffsets[0]; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(0), metadata.rootNodeOffsets[1]; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(0), metadata.rootNodeOffsets[2]; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestMetadataLoadMetadata(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		stream := createTestStream(t)
		_, err := stream.Add(append([]byte(ExpectedMagicBytes), []byte{
			0x41, 0x42, 0x43, // magicBytes
			0, 0x1, // version
			0, 0, 0, 0, 0, 0, 0x4, 0xD2, // inputFileModificationTime
			0, 0, 0, 0, 0, 0, 0, 0x2A, // inputFileSize
			1,                        // ignoreCase
			0, 0, 0, 0, 0, 0, 0, 0x3, // rootNodeOffsetCount
			0, 0, 0, 0, 0, 0, 0, 0x1, // rootNodeOffsets[0]
			0, 0, 0, 0, 0, 0, 0, 0x2, // rootNodeOffsets[1]
			0, 0, 0, 0, 0, 0, 0, 0x3, // rootNodeOffsets[2]
		}...)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		metadata, err := LoadMetadata(stream)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := ExpectedMagicBytes, string(metadata.magicBytes); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := uint16(1), metadata.version; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(1234), metadata.inputFileModificationTime.Unix(); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(42), metadata.inputFileSize; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := true, metadata.ignoreCase; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		if expect, got := 3, len(metadata.rootNodeOffsets); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(1), metadata.rootNodeOffsets[0]; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(2), metadata.rootNodeOffsets[1]; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(3), metadata.rootNodeOffsets[2]; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestMetadataSerialize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		metadata := Metadata{
			magicBytes:                []byte("ABC"),
			version:                   1,
			inputFileModificationTime: time.Unix(1234, 0),
			inputFileSize:             42,
			ignoreCase:                true,
			rootNodeOffsets:           []TreeNodeOffset{1, 2, 3},
		}

		expectBytes := []byte{
			0x41, 0x42, 0x43, // magicBytes
			0, 0x1, // version
			0, 0, 0, 0, 0, 0, 0x4, 0xD2, // inputFileModificationTime
			0, 0, 0, 0, 0, 0, 0, 0x2A, // inputFileSize
			1,                        // ignoreCase
			0, 0, 0, 0, 0, 0, 0, 0x3, // rootNodeOffsetCount
			0, 0, 0, 0, 0, 0, 0, 0x1, // rootNodeOffsets[0]
			0, 0, 0, 0, 0, 0, 0, 0x2, // rootNodeOffsets[1]
			0, 0, 0, 0, 0, 0, 0, 0x3, // rootNodeOffsets[2]
		}

		gotBytes, err := metadata.Serialize()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := fmt.Sprintf("%x", expectBytes), fmt.Sprintf("%x", gotBytes); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestMetadataUnserialize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		metadata := Metadata{}

		err := metadata.Unserialize(bytes.NewReader([]byte{
			0x41, 0x42, 0x43, // magicBytes
			0, 0x1, // version
			0, 0, 0, 0, 0, 0, 0x4, 0xD2, // inputFileModificationTime
			0, 0, 0, 0, 0, 0, 0, 0x2A, // inputFileSize
			1,                        // ignoreCase
			0, 0, 0, 0, 0, 0, 0, 0x3, // rootNodeOffsetCount
			0, 0, 0, 0, 0, 0, 0, 0x1, // rootNodeOffsets[0]
			0, 0, 0, 0, 0, 0, 0, 0x2, // rootNodeOffsets[1]
			0, 0, 0, 0, 0, 0, 0, 0x3, // rootNodeOffsets[2]
		}))
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := "ABC", string(metadata.magicBytes); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := uint16(1), metadata.version; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(1234), metadata.inputFileModificationTime.Unix(); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(42), metadata.inputFileSize; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := true, metadata.ignoreCase; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		if expect, got := 3, len(metadata.rootNodeOffsets); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(1), metadata.rootNodeOffsets[0]; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(2), metadata.rootNodeOffsets[1]; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(3), metadata.rootNodeOffsets[2]; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
	t.Run("from serialize", func(t *testing.T) {
		serialized, err := Metadata{
			magicBytes:                []byte("ABC"),
			version:                   1,
			inputFileModificationTime: time.Unix(1234, 0),
			inputFileSize:             42,
			ignoreCase:                true,
			rootNodeOffsets:           []TreeNodeOffset{1, 2, 3},
		}.Serialize()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		metadata := Metadata{}
		err := metadata.Unserialize(bytes.NewReader(serialized))
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := "ABC", string(metadata.magicBytes); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := uint16(1), metadata.version; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(1234), metadata.inputFileModificationTime.Unix(); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := int64(42), metadata.inputFileSize; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := true, metadata.ignoreCase; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}

		if expect, got := 3, len(metadata.rootNodeOffsets); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(1), metadata.rootNodeOffsets[0]; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(2), metadata.rootNodeOffsets[1]; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
		if expect, got := TreeNodeOffset(3), metadata.rootNodeOffsets[2]; expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestMetadataSave(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		stream := createTestStream(t)

		metadata := Metadata{
			stream:                    stream,
			magicBytes:                []byte("ABC"),
			version:                   1,
			inputFileModificationTime: time.Unix(1234, 0),
			inputFileSize:             42,
			ignoreCase:                true,
			rootNodeOffsets:           []TreeNodeOffset{1, 2, 3},
		}
		if err := metadata.Save(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		expectBytes := []byte{
			0x41, 0x42, 0x43, // magicBytes
			0, 0x1, // version
			0, 0, 0, 0, 0, 0, 0x4, 0xD2, // inputFileModificationTime
			0, 0, 0, 0, 0, 0, 0, 0x2A, // inputFileSize
			1,                        // ignoreCase
			0, 0, 0, 0, 0, 0, 0, 0x3, // rootNodeOffsetCount
			0, 0, 0, 0, 0, 0, 0, 0x1, // rootNodeOffsets[0]
			0, 0, 0, 0, 0, 0, 0, 0x2, // rootNodeOffsets[1]
			0, 0, 0, 0, 0, 0, 0, 0x3, // rootNodeOffsets[2]
		}
		gotBytes, err := stream.Get(0, len(expectBytes))
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := fmt.Sprintf("%x", expectBytes), fmt.Sprintf("%x", gotBytes); expect != got {
			t.Fatalf("Expected %v, got %v", expect, got)
		}
	})
}

func TestMetadataAssertValid(t *testing.T) {
	file := &fstest.MapFile{
		Data:    make([]byte, 42),
		ModTime: time.Now(),
	}
	fileInfo, err := fstest.MapFS{"file": file}.Stat("file")
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	metadata := Metadata{
		magicBytes:                []byte(ExpectedMagicBytes),
		version:                   CurrentVersion,
		inputFileModificationTime: file.ModTime,
		inputFileSize:             int64(len(file.Data)),
		ignoreCase:                true,
		rootNodeOffsets:           []TreeNodeOffset{1, 2},
	}
	input := MetadataInput{
		InputFileStats: fileInfo,
		IgnoreCase:     true,
		RootNodesCount: 2,
	}

	t.Run("valid", func(t *testing.T) {
		if err := metadata.AssertValid(input); err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
	t.Run("wrong version", func(t *testing.T) {
		metadata.version = CurrentVersion + 1
		if metadata.AssertValid(input) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("wrong magic bytes", func(t *testing.T) {
		metadata.magicBytes = []byte("wrong bytes")
		if metadata.AssertValid(input) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("wrong time", func(t *testing.T) {
		metadata.inputFileModificationTime = time.Unix(1234, 0)
		if metadata.AssertValid(input) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("wrong size", func(t *testing.T) {
		metadata.inputFileSize = int64(len(file.Data) + 1)
		if metadata.AssertValid(input) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("wrong ignoreCase", func(t *testing.T) {
		metadata.ignoreCase = !input.IgnoreCase
		if metadata.AssertValid(input) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("wrong rootNodes count", func(t *testing.T) {
		metadata.rootNodeOffsets = []TreeNodeOffset{1}
		if metadata.AssertValid(input) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
}
