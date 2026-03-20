package filekit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/arimatakao/mdx/filekit/metadata"
)

type dirContainer struct {
	tempDir   string
	pageIndex int
}

func newDirContainer() (*dirContainer, error) {
	tempDir, err := os.MkdirTemp("", "mdxdirfiles")
	if err != nil {
		return nil, err
	}

	return &dirContainer{
		tempDir:   tempDir,
		pageIndex: 1,
	}, nil
}

func (d *dirContainer) WriteOnDiskAndClose(outputDir, outputFileName string,
	m metadata.Metadata, chapterRange string) error {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return err
	}

	outputPath := safeOutputDirPath(outputDir, outputFileName)
	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		return err
	}

	entries, err := os.ReadDir(d.tempDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(d.tempDir, entry.Name())
		dstPath := filepath.Join(outputPath, entry.Name())
		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	return os.RemoveAll(d.tempDir)
}

func (d *dirContainer) AddFile(fileExt string, imageBytes []byte) error {
	fileName := fmt.Sprintf("%02d.%s", d.pageIndex, fileExt)
	filePath := filepath.Join(d.tempDir, fileName)
	if err := os.WriteFile(filePath, imageBytes, os.ModePerm); err != nil {
		return err
	}

	d.pageIndex++
	return nil
}

func copyFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		_ = dstFile.Close()
		return err
	}

	if err := dstFile.Close(); err != nil {
		return err
	}

	return nil
}
