package filekit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arimatakao/mdx/filekit/metadata"
)

const (
	// CBZ_EXT is a CBZ archive container extension.
	CBZ_EXT = "cbz"
	// PDF_EXT is a PDF container extension.
	PDF_EXT = "pdf"
	// EPUB_EXT is an EPUB container extension.
	EPUB_EXT = "epub"
	// DIR_EXT stores chapter pages in a plain directory.
	DIR_EXT = "dir"
)

// ErrExtensionNotSupport is returned when a requested output container
// extension is unknown.
var ErrExtensionNotSupport = errors.New("extension container is not supported")

// IsNotSupported reports whether fileFormat is not one of the supported
// container extensions.
func IsNotSupported(fileFormat string) bool {
	return fileFormat != CBZ_EXT &&
		fileFormat != PDF_EXT &&
		fileFormat != EPUB_EXT &&
		fileFormat != DIR_EXT
}

// Container describes a writable chapter output that accepts page images and
// then persists itself to disk.
type Container interface {
	// WriteOnDiskAndClose finalizes container content and writes it into
	// outputDir using outputFileName as a base name.
	WriteOnDiskAndClose(outputDir string, outputFileName string, m metadata.Metadata, chapterRange string) error
	// AddFile appends a new page represented by imageBytes with fileExt format.
	AddFile(fileExt string, imageBytes []byte) error
}

// NewContainer creates a container by file extension.
//
// Supported extensions are CBZ_EXT, PDF_EXT, EPUB_EXT and DIR_EXT.
func NewContainer(extension string) (Container, error) {

	switch extension {
	case CBZ_EXT:
		return newCBZArchive()
	case PDF_EXT:
		return newPdfFile()
	case EPUB_EXT:
		return newEpubArchive()
	case DIR_EXT:
		return newDirContainer()
	}

	return nil, ErrExtensionNotSupport
}

func safeOutputPath(outputDir, outputFileName, extension string) string {
	outputFileName = safeOutputName(outputFileName)

	outputPath := filepath.Join(outputDir, outputFileName+"."+extension)

	for count := 1; ; count++ {
		_, err := os.Stat(outputPath)
		if errors.Is(err, os.ErrNotExist) {
			break
		}
		outputPath = filepath.Join(outputDir,
			fmt.Sprintf("%s (%d).%s", outputFileName, count, extension))
	}
	return outputPath
}

func safeOutputDirPath(outputDir, outputFileName string) string {
	outputFileName = safeOutputName(outputFileName)

	outputPath := filepath.Join(outputDir, outputFileName)

	for count := 1; ; count++ {
		_, err := os.Stat(outputPath)
		if errors.Is(err, os.ErrNotExist) {
			break
		}
		outputPath = filepath.Join(outputDir,
			fmt.Sprintf("%s (%d)", outputFileName, count))
	}
	return outputPath
}

func safeOutputName(outputFileName string) string {
	// unix
	outputFileName = strings.ReplaceAll(outputFileName, "/", "_")
	outputFileName = strings.ReplaceAll(outputFileName, `\`, "_")
	// windows
	outputFileName = strings.ReplaceAll(outputFileName, "<", "_")
	outputFileName = strings.ReplaceAll(outputFileName, ">", "_")
	outputFileName = strings.ReplaceAll(outputFileName, ":", "_")
	outputFileName = strings.ReplaceAll(outputFileName, `"`, "_")
	outputFileName = strings.ReplaceAll(outputFileName, "?", "_")
	outputFileName = strings.ReplaceAll(outputFileName, "*", "_")
	outputFileName = strings.ReplaceAll(outputFileName, "|", "-")
	return outputFileName
}
