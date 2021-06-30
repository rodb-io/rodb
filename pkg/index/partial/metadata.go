package partial

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"rodb.io/pkg/input"
	"time"
)

// Current version of the indexing protocol
const CurrentVersion = uint16(1)

// Default magic bytes
const ExpectedMagicBytes = "RODB/INDEX/PARTIAL"

type Metadata struct {
	stream                    *Stream
	magicBytes                []byte
	version                   uint16
	inputFileModificationTime time.Time
	inputFileSize             int64
	ignoreCase                bool
	completed                 bool
	rootNodeOffsets           []TreeNodeOffset
}

type MetadataInput struct {
	Input          input.Input
	IgnoreCase     bool
	RootNodesCount int
}

func NewMetadata(stream *Stream, input MetadataInput) (*Metadata, error) {
	size, err := input.Input.Size()
	if err != nil {
		return nil, err
	}

	modTime, err := input.Input.ModTime()
	if err != nil {
		return nil, err
	}

	metadata := &Metadata{
		stream:                    stream,
		magicBytes:                []byte(ExpectedMagicBytes),
		version:                   CurrentVersion,
		inputFileModificationTime: modTime,
		inputFileSize:             size,
		ignoreCase:                input.IgnoreCase,
		rootNodeOffsets:           make([]TreeNodeOffset, input.RootNodesCount),
		completed:                 false,
	}

	// Saving it first to allocate the required space
	if err := metadata.Save(); err != nil {
		return nil, err
	}

	return metadata, nil
}

func LoadMetadata(stream *Stream) (*Metadata, error) {
	metadata := &Metadata{
		stream: stream,
	}

	reader, err := stream.GetReaderFrom(0)
	if err != nil {
		return nil, err
	}

	err = metadata.Unserialize(reader)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (metadata *Metadata) SetRootNode(index int, node *TreeNode) {
	metadata.rootNodeOffsets[index] = node.offset
}

// Sets the completed flag, which records wether or not the index
// generation has been finished
func (metadata *Metadata) SetCompleted(completed bool) {
	metadata.completed = completed
}

func (metadata *Metadata) GetRootNode(index int) (*TreeNode, error) {
	return GetTreeNode(metadata.stream, metadata.rootNodeOffsets[index])
}

func (metadata *Metadata) Serialize() ([]byte, error) {
	var err error
	buffer := &bytes.Buffer{}

	if err = binary.Write(buffer, binary.BigEndian, metadata.magicBytes); err != nil {
		return nil, err
	}
	if err = binary.Write(buffer, binary.BigEndian, metadata.version); err != nil {
		return nil, err
	}
	if err = binary.Write(buffer, binary.BigEndian, int64(metadata.inputFileModificationTime.Unix())); err != nil {
		return nil, err
	}
	if err = binary.Write(buffer, binary.BigEndian, metadata.inputFileSize); err != nil {
		return nil, err
	}
	if err = binary.Write(buffer, binary.BigEndian, metadata.ignoreCase); err != nil {
		return nil, err
	}
	if err = binary.Write(buffer, binary.BigEndian, metadata.completed); err != nil {
		return nil, err
	}
	if err = binary.Write(buffer, binary.BigEndian, int64(len(metadata.rootNodeOffsets))); err != nil {
		return nil, err
	}
	for _, offset := range metadata.rootNodeOffsets {
		if err = binary.Write(buffer, binary.BigEndian, offset); err != nil {
			return nil, err
		}
	}

	return buffer.Bytes(), nil
}

func (metadata *Metadata) Unserialize(data io.Reader) error {
	var err error

	metadata.magicBytes = make([]byte, len(ExpectedMagicBytes))
	if err = binary.Read(data, binary.BigEndian, &metadata.magicBytes); err != nil {
		return err
	}

	if err = binary.Read(data, binary.BigEndian, &metadata.version); err != nil {
		return err
	}

	var inputFileModificationTimeUnix int64
	if err = binary.Read(data, binary.BigEndian, &inputFileModificationTimeUnix); err != nil {
		return err
	}
	metadata.inputFileModificationTime = time.Unix(inputFileModificationTimeUnix, 0)

	if err = binary.Read(data, binary.BigEndian, &metadata.inputFileSize); err != nil {
		return err
	}
	if err = binary.Read(data, binary.BigEndian, &metadata.ignoreCase); err != nil {
		return err
	}
	if err = binary.Read(data, binary.BigEndian, &metadata.completed); err != nil {
		return err
	}

	var rootNodeOffsetCount int64
	if err = binary.Read(data, binary.BigEndian, &rootNodeOffsetCount); err != nil {
		return err
	}
	metadata.rootNodeOffsets = make([]TreeNodeOffset, int(rootNodeOffsetCount))
	for i := int64(0); i < rootNodeOffsetCount; i++ {
		if err = binary.Read(data, binary.BigEndian, &metadata.rootNodeOffsets[i]); err != nil {
			return err
		}
	}

	return nil
}

func (metadata *Metadata) Save() error {
	serialized, err := metadata.Serialize()
	if err != nil {
		return err
	}

	err = metadata.stream.Replace(0, serialized)
	if err != nil {
		return err
	}

	return nil
}

// Validates that the metadata of the file is an RODB partial index
// and matches the given configuration as well as the current version
func (metadata *Metadata) AssertValid(expect MetadataInput) error {
	if metadata.version != CurrentVersion {
		return fmt.Errorf("The index file is not compatible with the current version of this software.")
	}

	if string(metadata.magicBytes) != ExpectedMagicBytes {
		return fmt.Errorf("The given file is not a partial index.")
	}

	modTime, err := expect.Input.ModTime()
	if err != nil {
		return err
	}
	if metadata.inputFileModificationTime.Unix() != modTime.Unix() {
		return fmt.Errorf("The input file has been modified since the index generation.")
	}

	size, err := expect.Input.Size()
	if err != nil {
		return err
	}
	if metadata.inputFileSize != size {
		return fmt.Errorf("The input file size has changed since the index generation.")
	}

	if metadata.ignoreCase != expect.IgnoreCase {
		return fmt.Errorf("The configured ignoreCase value does not match the index file contents.")
	}
	if !metadata.completed {
		return fmt.Errorf("The previous indexing process has not ended properly. Please remove the corrupted file and try again.")
	}

	if len(metadata.rootNodeOffsets) != expect.RootNodesCount {
		return fmt.Errorf("The configured properties does not match the index file contents.")
	}

	return nil
}
