package filekit

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/arimatakao/mdx/filekit/metadata"
	"github.com/pterm/pterm"
)

const (
	CBZ_EXT  = "cbz"
	PDF_EXT  = "pdf"
	EPUB_EXT = "epub"
)

var ErrExtensionNotSupport = errors.New("extension container is not supported")

type Container interface {
	WriteOnDiskAndClose(outputDir string, outputFileName string, m metadata.Metadata, chapterRange string) error
	AddFile(fileExt string, imageBytes []byte) error
}

func NewContainer(extension string) (Container, error) {

	switch extension {
	case CBZ_EXT:
		return newCBZArchive()
	case PDF_EXT:
		return newPdfFile()
	case EPUB_EXT:
		return newEpubArchive()
	}

	return nil, ErrExtensionNotSupport
}

func safeOutputPath(outputDir, outputFileName, extension string) string {
	outputFileName = strings.ReplaceAll(outputFileName, "/", "_")
	outputFileName = strings.ReplaceAll(outputFileName, `\`, "_")

	outputPath := filepath.Join(outputDir, outputFileName+"."+extension)

	for count := 1; ; count++ {
		_, err := os.Stat(outputPath)
		if errors.Is(err, os.ErrNotExist) {
			break
		}
		outputPath = filepath.Join(outputDir,
			pterm.Sprintf("%s (%d).%s", outputFileName, count, extension))
	}
	return outputPath
}
