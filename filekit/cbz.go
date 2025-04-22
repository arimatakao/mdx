package filekit

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"

	"github.com/pterm/pterm"

	"github.com/arimatakao/mdx/filekit/metadata"
)

type cbzArchive struct {
	buf         *bytes.Buffer
	writer      *zip.Writer
	pageCounter int
}

// fileName without extension
func newCBZArchive() (*cbzArchive, error) {
	buf := new(bytes.Buffer)

	zipWriter := zip.NewWriter(buf)

	c := cbzArchive{
		buf:         buf,
		writer:      zipWriter,
		pageCounter: 1,
	}

	return &c, nil
}

// ALWAYS close archive after all operations
func (c *cbzArchive) WriteOnDiskAndClose(outputDir, outputFileName string,
	m metadata.Metadata, chapterRange string) error {
	// ComicBookInfo metadata
	comment, err := json.Marshal(m.CBI)
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

	comicInfoContent, err := xml.Marshal(m.CI)
	if err != nil {
		return err
	}

	cireader := bytes.NewReader(comicInfoContent)
	if _, err := io.Copy(w, cireader); err != nil {
		return err
	}

	outputPath := safeOutputPath(outputDir, outputFileName, CBZ_EXT)

	err = c.writer.Close()
	if err != nil {
		return err
	}

	err = os.WriteFile(outputPath, c.buf.Bytes(), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (c *cbzArchive) AddFile(fileExt string, src []byte) error {
	fileName := pterm.Sprintf("%02d.%s", c.pageCounter, fileExt)
	buf := bytes.NewBuffer(src)
	w, err := c.writer.Create(fileName)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, buf); err != nil {
		return err
	}
	c.pageCounter++
	return nil
}
