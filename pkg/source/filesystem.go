package source

import (
	"io"
	"os"
	"path/filepath"
	"rods/pkg/config"
)

type Filesystem struct{
	config *config.FilesystemSourceConfig
}

func NewFilesystem(config *config.FilesystemSourceConfig) *Filesystem {
	return &Filesystem{
		config: config,
	}
}

func (fs *Filesystem) Open(filePath string) (io.ReadSeeker, error) {
	path := filepath.Join(fs.config.Path, filePath)
	file, err := os.Open(path)
	return io.ReadSeeker(file), err
}
