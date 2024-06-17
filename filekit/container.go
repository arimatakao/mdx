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
	WriteOnDiskAndClose(outputDir string, outputFileName string, m metadata.Metadata) error
	AddFile(fileName string, imageBytes []byte) error
}

func NewContainer(extension string) (Container, error) {

	switch extension {
	case CBZ_EXT:
		return newCBZArchive()
	case PDF_EXT:
		return newPdfFile()
	}

	return nil, ErrExtensionNotSupport
}
