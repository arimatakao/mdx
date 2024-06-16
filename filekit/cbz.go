package filekit

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"

	"github.com/arimatakao/mdx/filekit/metadata"
)

type cbzArchive struct {
	buf       *bytes.Buffer
	writer    *zip.Writer
	outputDir string
}

// fileName without extension
func NewCBZArchive(outputDir, fileName string, m metadata.Metadata) (cbzArchive, error) {
	fileName += ".cbz"

	buf := new(bytes.Buffer)

	zipWriter := zip.NewWriter(buf)

	c := cbzArchive{
		buf:       buf,
		writer:    zipWriter,
		outputDir: filepath.Join(outputDir, fileName),
	}

	// ComicBookInfo metadata
	comment, err := json.Marshal(m.CBI)
	if err != nil {
		return cbzArchive{}, err
	}
	err = c.writer.SetComment(string(comment))
	if err != nil {
		return cbzArchive{}, err
	}

	// ComicRack metadata
	w, err := c.writer.Create("ComicInfo.xml")
	if err != nil {
		return cbzArchive{}, err
	}

	comicInfoContent, err := xml.Marshal(m.CI)
	if err != nil {
		return cbzArchive{}, err
	}

	cireader := bytes.NewReader(comicInfoContent)
	if _, err := io.Copy(w, cireader); err != nil {
		return cbzArchive{}, err
	}

	return c, nil
}

// ALWAYS close archive after all operations
func (c cbzArchive) WriteOnDiskAndClose() error {
	err := c.writer.Close()
	if err != nil {
		return err
	}

	return os.WriteFile(c.outputDir, c.buf.Bytes(), os.ModePerm)
}

func (c cbzArchive) AddFile(fileName string, src []byte) error {
	buf := bytes.NewBuffer(src)
	w, err := c.writer.Create(fileName)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, buf); err != nil {
		return err
	}
	return nil
}
