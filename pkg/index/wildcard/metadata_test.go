package wildcard

import (
	"bytes"
	"fmt"
	"rodb.io/pkg/input"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/input/record"
	"testing"
	"time"
)

func TestMetadataNewMetadata(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		stream := createTestStream(t)

		input := input.NewMock(parser.NewMock(), make([]record.Record, 42))
		input.SetModTime(time.Unix(1234, 0))

		metadata, err := NewMetadata(stream, MetadataInput{
			Input:          input,
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
			0,                        // completed
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
		if expect, got := false, metadata.completed; expect != got {
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
		data := append([]byte(ExpectedMagicBytes), []byte{
			0, 0x1, // version
			0, 0, 0, 0, 0, 0, 0x4, 0xD2, // inputFileModificationTime
			0, 0, 0, 0, 0, 0, 0, 0x2A, // inputFileSize
			1,                        // ignoreCase
			1,                        // completed
			0, 0, 0, 0, 0, 0, 0, 0x3, // rootNodeOffsetCount
			0, 0, 0, 0, 0, 0, 0, 0x1, // rootNodeOffsets[0]
			0, 0, 0, 0, 0, 0, 0, 0x2, // rootNodeOffsets[1]
			0, 0, 0, 0, 0, 0, 0, 0x3, // rootNodeOffsets[2]
		}...)
		if err := stream.Replace(0, data); err != nil {
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
		if expect, got := true, metadata.completed; expect != got {
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
			completed:                 false,
			rootNodeOffsets:           []TreeNodeOffset{1, 2, 3},
		}

		expectBytes := []byte{
			0x41, 0x42, 0x43, // magicBytes
			0, 0x1, // version
			0, 0, 0, 0, 0, 0, 0x4, 0xD2, // inputFileModificationTime
			0, 0, 0, 0, 0, 0, 0, 0x2A, // inputFileSize
			1,                        // ignoreCase
			0,                        // completed
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

		data := bytes.NewReader(append([]byte(ExpectedMagicBytes), []byte{
			0, 0x1, // version
			0, 0, 0, 0, 0, 0, 0x4, 0xD2, // inputFileModificationTime
			0, 0, 0, 0, 0, 0, 0, 0x2A, // inputFileSize
			1,                        // ignoreCase
			1,                        // completed
			0, 0, 0, 0, 0, 0, 0, 0x3, // rootNodeOffsetCount
			0, 0, 0, 0, 0, 0, 0, 0x1, // rootNodeOffsets[0]
			0, 0, 0, 0, 0, 0, 0, 0x2, // rootNodeOffsets[1]
			0, 0, 0, 0, 0, 0, 0, 0x3, // rootNodeOffsets[2]
		}...))
		if err := metadata.Unserialize(data); err != nil {
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
		if expect, got := true, metadata.completed; expect != got {
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
		serialized, err := (&Metadata{
			magicBytes:                []byte(ExpectedMagicBytes),
			version:                   1,
			inputFileModificationTime: time.Unix(1234, 0),
			inputFileSize:             42,
			ignoreCase:                true,
			completed:                 false,
			rootNodeOffsets:           []TreeNodeOffset{1, 2, 3},
		}).Serialize()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		metadata := Metadata{}
		if err := metadata.Unserialize(bytes.NewReader(serialized)); err != nil {
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
		if expect, got := false, metadata.completed; expect != got {
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
			completed:                 false,
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
			0,                        // completed
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
	modTime := time.Now()
	data := make([]record.Record, 42)
	input := input.NewMock(parser.NewMock(), data)
	input.SetModTime(modTime)

	metadata := Metadata{
		magicBytes:                []byte(ExpectedMagicBytes),
		version:                   CurrentVersion,
		inputFileModificationTime: modTime,
		inputFileSize:             int64(len(data)),
		ignoreCase:                true,
		completed:                 true,
		rootNodeOffsets:           []TreeNodeOffset{1, 2},
	}
	metadataInput := MetadataInput{
		Input:          input,
		IgnoreCase:     true,
		RootNodesCount: 2,
	}

	t.Run("valid", func(t *testing.T) {
		if err := metadata.AssertValid(metadataInput); err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
	t.Run("wrong version", func(t *testing.T) {
		metadata.version = CurrentVersion + 1
		if metadata.AssertValid(metadataInput) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("wrong magic bytes", func(t *testing.T) {
		metadata.magicBytes = []byte("wrong bytes")
		if metadata.AssertValid(metadataInput) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("wrong time", func(t *testing.T) {
		metadata.inputFileModificationTime = time.Unix(1234, 0)
		if metadata.AssertValid(metadataInput) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("wrong size", func(t *testing.T) {
		metadata.inputFileSize = int64(len(data) + 1)
		if metadata.AssertValid(metadataInput) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("wrong ignoreCase", func(t *testing.T) {
		metadata.ignoreCase = !metadataInput.IgnoreCase
		if metadata.AssertValid(metadataInput) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("not completed", func(t *testing.T) {
		metadata.completed = false
		if metadata.AssertValid(metadataInput) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
	t.Run("wrong rootNodes count", func(t *testing.T) {
		metadata.rootNodeOffsets = []TreeNodeOffset{1}
		if metadata.AssertValid(metadataInput) == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
}
