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

type cbz struct {
	file   *os.File
	writer *zip.Writer
}

// fileName without extension
func NewCBZFile(outputDir, fileName string) (cbz, error) {
	fileName += ".cbz"

	err := os.MkdirAll(filepath.Join("", outputDir), os.ModePerm)
	if err != nil {
		return cbz{}, err
	}

	archive, err := os.Create(filepath.Join(outputDir, fileName))
	if err != nil {
		return cbz{}, err
	}

	zipWriter := zip.NewWriter(archive)

	return cbz{
		file:   archive,
		writer: zipWriter,
	}, nil
}

// ALWAYS close archive after all operations
func (c cbz) Close() error {
	err := c.writer.Close()
	if err != nil {
		return err
	}
	err = c.file.Close()
	if err != nil {
		return err
	}

	return nil
}

func (c cbz) AddFile(fileName string, src io.Reader) error {
	w, err := c.writer.Create(fileName)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, src); err != nil {
		return err
	}
	return nil
}

func (c cbz) AddMetadata(m metadata.CBZMetadata) error {
	// ComicBookInfo metadata
	comment, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = c.writer.SetComment(string(comment))
	if err != nil {
		return err
	}

	// ComicRack metadata
	w, err := c.writer.Create("ComicInfo.xml")
	if err != nil {
		return err
	}

	comicInfoContent, err := xml.Marshal(m.ComicInfoMetadata)
	if err != nil {
		return err
	}

	cireader := bytes.NewReader(comicInfoContent)
	if _, err := io.Copy(w, cireader); err != nil {
		return err
	}

	return nil
}
