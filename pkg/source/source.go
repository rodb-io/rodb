package source

import (
	"io"
)

type Source interface {
	Open(filePath string) (io.ReadSeeker, error)
}
