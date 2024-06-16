package filekit

import (
	"errors"

	"github.com/arimatakao/mdx/filekit/metadata"
)

const (
	CBZ_EXT = "cbz"
	PDF_EXT = "pdf"
)

var ErrExtensionNotSupport = errors.New("extension container is not supported")

type Container interface {
	WriteOnDiskAndClose() error
	AddFile(fileName string, imageBytes []byte) error
}

func NewContainer(extension string, outputDir, fileName string,
	m metadata.Metadata) (Container, error) {

	switch extension {
	case CBZ_EXT:
		return NewCBZArchive(outputDir, fileName, m)
	case PDF_EXT:
		return NewPdfFile(outputDir, fileName, m)
	}

	return nil, ErrExtensionNotSupport
}
