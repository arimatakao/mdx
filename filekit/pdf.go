package filekit

import (
	"bytes"
	"image"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/arimatakao/mdx/filekit/metadata"
	"github.com/signintech/gopdf"
)

type pdfFile struct {
	pdf        *gopdf.GoPdf
	outputDir  string
	outputFile string
}

func NewPdfFile(outputDir, fileName string, m metadata.Metadata) (pdfFile, error) {
	fileName += ".pdf"

	pdf := new(gopdf.GoPdf)
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeA4,
	})

	author := m.P.Authors + " | " + m.P.Artists

	pdf.SetInfo(gopdf.PdfInfo{
		Title:        m.CI.Title,
		Author:       author,
		Subject:      m.CBI.ComicBookInfoData.Title,
		Creator:      m.CBI.AppID,
		Producer:     m.CBI.AppID,
		CreationDate: time.Now(),
	})

	pdf.SetNoCompression()

	return pdfFile{
		pdf:        pdf,
		outputDir:  outputDir,
		outputFile: fileName,
	}, nil
}

func (p pdfFile) WriteOnDiskAndClose() error {

	err := os.MkdirAll(p.outputDir, os.ModePerm)
	if err != nil {
		return err
	}

	err = p.pdf.WritePdf(filepath.Join(p.outputDir, p.outputFile))
	if err != nil {
		return err
	}
	return p.pdf.Close()
}

func (p pdfFile) AddFile(fileName string, imageBytes []byte) error {
	buf := bytes.NewBuffer(imageBytes)
	imgWidth, imgHeight, err := getImageDimensions(buf)
	if err != nil {
		return err
	}

	imageReader := bytes.NewBuffer(imageBytes)

	p.pdf.AddPageWithOption(gopdf.PageOption{
		PageSize: &gopdf.Rect{
			W: imgWidth,
			H: imgHeight,
		},
	})

	imgH1, err := gopdf.ImageHolderByReader(imageReader)
	if err != nil {
		return err
	}
	if err := p.pdf.ImageByHolder(imgH1, 0, 0, &gopdf.Rect{
		W: float64(imgWidth),
		H: float64(imgHeight),
	}); err != nil {
		return err
	}

	return nil
}

func getImageDimensions(img io.Reader) (float64, float64, error) {
	config, _, err := image.DecodeConfig(img)
	if err != nil {
		return 0, 0, err
	}
	return float64(config.Width), float64(config.Height), nil
}
