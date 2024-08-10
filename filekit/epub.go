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
	pageIndex  int
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
		pageIndex:  1,
	}, nil
}

func (e *epubArchive) WriteOnDiskAndClose(outputDir string, outputFileName string,
	m metadata.Metadata, chapterRange string) error {

	for i, filePath := range e.filesPaths {
		indexPage := fmt.Sprintf("%02d", i+1)
		imageEpubPath, err := e.b.AddImage(filePath, indexPage)
		if err != nil {
			return err
		}
		sectionStr := fmt.Sprintf(imageSectionTemplate, imageEpubPath, indexPage)
		_, err = e.b.AddSection(sectionStr, indexPage, "", "")
		if err != nil {
			return err
		}
	}

	bookTitle := fmt.Sprintf("%s vol%s ch%s", m.CI.Title, m.CI.Volume, m.CI.Number)
	if chapterRange != "" {
		bookTitle = fmt.Sprintf("%s ch%s", m.CI.Title, chapterRange)
	}
	e.b.SetTitle(bookTitle)

	authors := m.P.Authors + " | " + m.P.Artists
	e.b.SetAuthor(authors)

	e.b.SetLang(m.CI.LanguageISO)

	e.b.SetDescription(m.CI.Summary)

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return err
	}

	outputPath := safeOutputPath(outputDir, outputFileName, EPUB_EXT)

	err = e.b.Write(outputPath)
	if err != nil {
		return err
	}

	return os.RemoveAll(e.tempDir)
}

func (e *epubArchive) AddFile(fileExt string, imageBytes []byte) error {
	fileName := fmt.Sprintf("%02d.%s", e.pageIndex, fileExt)
	filePath := filepath.Join(e.tempDir, fileName)
	err := os.WriteFile(filePath, imageBytes, os.ModePerm)
	if err != nil {
		return err
	}

	e.filesPaths = append(e.filesPaths, filePath)

	e.pageIndex++
	return nil
}
