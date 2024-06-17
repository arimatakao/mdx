package filekit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arimatakao/mdx/filekit/metadata"
	"github.com/go-shiori/go-epub"
)

const imageSectionTemplate = `<img src="%s" alt="%s" />`

type epubArchive struct {
	b          *epub.Epub
	tempDir    string
	filesPaths []string
}

func newEpubArchive() (*epubArchive, error) {
	book, err := epub.NewEpub("")
	if err != nil {
		return &epubArchive{}, err
	}

	dir, err := os.MkdirTemp("", "mdxepubfiles")
	if err != nil {
		return &epubArchive{}, err
	}

	return &epubArchive{
		b:          book,
		tempDir:    dir,
		filesPaths: []string{},
	}, nil
}

func (e epubArchive) WriteOnDiskAndClose(outputDir string, outputFileName string,
	m metadata.Metadata) error {

	for i, filePath := range e.filesPaths {
		indexStr := fmt.Sprint(i + 1)
		imageEpubPath, err := e.b.AddImage(filePath, indexStr)
		if err != nil {
			return err
		}
		sectionStr := fmt.Sprintf(imageSectionTemplate, imageEpubPath, indexStr)
		_, err = e.b.AddSection(sectionStr, indexStr, "", "")
		if err != nil {
			return err
		}
	}

	bookTitle := fmt.Sprintf("%s vol%s ch%s", m.CI.Title, m.CI.Volume, m.CI.Number)
	e.b.SetTitle(bookTitle)

	authors := m.P.Authors + " | " + m.P.Artists
	e.b.SetAuthor(authors)

	e.b.SetLang(m.CI.LanguageISO)

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return err
	}

	err = e.b.Write(filepath.Join(outputDir, outputFileName))
	if err != nil {
		return err
	}

	return os.RemoveAll(e.tempDir)
}

func (e *epubArchive) AddFile(fileName string, imageBytes []byte) error {
	filePath := filepath.Join(e.tempDir, fileName)
	err := os.WriteFile(filePath, imageBytes, os.ModePerm)
	if err != nil {
		return err
	}

	e.filesPaths = append(e.filesPaths, filePath)

	return nil
}
