package mdx

import (
	"errors"
	"fmt"
	"os"

	"github.com/arimatakao/mdx/app"
	"github.com/arimatakao/mdx/filekit"
	"github.com/arimatakao/mdx/filekit/metadata"
	"github.com/arimatakao/mdx/mangadexapi"
	"github.com/pterm/pterm"
)

var (
	ErrEmptyChapters = errors.New("empty chapters")
)

type dlParam struct {
	mangaInfo      mangadexapi.MangaInfo
	chapters       []mangadexapi.ChapterFullInfo
	lowestChapter  int
	highestChapter int
	chaptersRange  string
	language       string
	translateGroup string
	outputDir      string
	outputExt      string
	isJpg          bool
	isMerge        bool
}

func NewDownloadParam(chaptersRange string, lowestChapter, highestChapter int,
	language, translateGroup, outputDir, outputExt string, isJpg, isMerge bool) dlParam {

	return dlParam{
		mangaInfo:      mangadexapi.MangaInfo{},
		chapters:       []mangadexapi.ChapterFullInfo{},
		lowestChapter:  lowestChapter,
		highestChapter: highestChapter,
		chaptersRange:  chaptersRange,
		language:       language,
		translateGroup: translateGroup,
		outputDir:      outputDir,
		outputExt:      outputExt,
		isJpg:          isJpg,
		isMerge:        isMerge,
	}
}

func (p dlParam) getMangaInfo(mangaId string) (mangadexapi.MangaInfo, error) {
	resp, err := client.GetMangaInfo(mangaId)
	if err != nil {
		return mangadexapi.MangaInfo{}, err
	}
	return resp.MangaInfo(), nil
}

func (p dlParam) downloadMergeChapters() {
	containerFile, err := filekit.NewContainer(p.outputExt)
	if err != nil {
		e.Printf("While creating output file: %v\n", err)
		os.Exit(1)
	}

	for _, chapter := range p.chapters {
		printChapterInfo(chapter)

		err = p.downloadProcess(containerFile, chapter)
		if err != nil {
			e.Printf("While downloading chapter: %v\n", err)
			os.Exit(1)
		}
	}

	filename := fmt.Sprintf("[%s] %s ch%s",
		p.language, p.mangaInfo.Title("en"), p.chaptersRange)
	metaInfo := metadata.NewMetadata(app.USER_AGENT, p.mangaInfo, p.chapters[0])
	err = containerFile.WriteOnDiskAndClose(p.outputDir, filename, metaInfo, p.chaptersRange)
	if err != nil {
		e.Printf("While saving %s on disk: %v\n", filename, err)
		os.Exit(1)
	}
}

func (p dlParam) downloadChapters() {
	for _, chapter := range p.chapters {
		printChapterInfo(chapter)

		containerFile, err := filekit.NewContainer(p.outputExt)
		if err != nil {
			e.Printf("While creating output file: %v\n", err)
			os.Exit(1)
		}

		err = p.downloadProcess(containerFile, chapter)
		if err != nil {
			e.Printf("While downloading chapter: %v\n", err)
			os.Exit(1)
		}

		filename := fmt.Sprintf("[%s] %s vol%s ch%s",
			p.language, p.mangaInfo.Title("en"), chapter.Volume(), chapter.Number())
		metaInfo := metadata.NewMetadata(app.USER_AGENT, p.mangaInfo, chapter)
		err = containerFile.WriteOnDiskAndClose(p.outputDir, filename, metaInfo, "")
		if err != nil {
			e.Printf("While saving %s on disk: %v\n", filename, err)
			os.Exit(1)
		}
	}
}

func (p dlParam) downloadProcess(outputFile filekit.Container,
	chapter mangadexapi.ChapterFullInfo) error {
	if len(p.chapters) == 0 {
		return ErrEmptyChapters
	}

	files := chapter.PngFiles
	imgExt := "png"
	if p.isJpg {
		files = chapter.JpgFiles
		imgExt = "jpg"
	}

	dlbar, _ := pterm.DefaultProgressbar.WithTotal(len(files)).
		WithTitle("Downloading pages...").
		WithBarStyle(pterm.NewStyle(pterm.FgGreen)).Start()
	defer dlbar.Stop()

	for _, imageFile := range files {
		outputImage, err := client.DownloadImage(chapter.DownloadBaseURL,
			chapter.HashId, imageFile, p.isJpg)
		if err != nil {
			dlbar.WithBarStyle(pterm.NewStyle(pterm.FgRed)).
				UpdateTitle("Failed downloading").Stop()
			return err
		}

		if err := outputFile.AddFile(imgExt, outputImage); err != nil {
			dlbar.WithBarStyle(pterm.NewStyle(pterm.FgRed)).
				UpdateTitle("Failed downloading").Stop()
			return err
		}
		dlbar.Increment()
	}
	dp.Println("")
	return nil
}

func (p dlParam) DownloadSpecificChapter(chapterId string) {
	spinnerChapInfo, _ := pterm.DefaultSpinner.Start("Fetching chapter info...")
	resp, err := client.GetChapterInfo(chapterId)
	if err != nil {
		spinnerChapInfo.Fail("Failed to get chapter info")
		os.Exit(1)
	}
	chapterInfo := resp.GetChapterInfo()
	chapterFullInfo, err := client.GetChapterImagesInFullInfo(chapterInfo)
	if err != nil {
		spinnerChapInfo.Fail("Failed to get chapter info")
		os.Exit(1)
	}
	spinnerChapInfo.Success("Fetched chapter info")

	mangaId := chapterInfo.GetMangaId()
	spinnerMangaInfo, _ := pterm.DefaultSpinner.Start("Fetching manga info...")
	mangaInfo, err := p.getMangaInfo(mangaId)
	if err != nil {
		spinnerMangaInfo.Fail("Failed to get manga info")
		os.Exit(1)
	}
	spinnerMangaInfo.Success("Fetched manga info")
	p.chapters = []mangadexapi.ChapterFullInfo{chapterFullInfo}
	printShortMangaInfo(mangaInfo)
	p.downloadChapters()
}

func (p dlParam) DownloadLastChapter(mangaId string) {
	spinnerMangaInfo, _ := pterm.DefaultSpinner.Start("Fetching manga info...")
	mangaInfo, err := p.getMangaInfo(mangaId)
	if err != nil {
		spinnerMangaInfo.Fail("Failed to get manga info")
		os.Exit(1)
	}
	spinnerMangaInfo.Success("Fetched manga info")
	p.mangaInfo = mangaInfo

	printShortMangaInfo(mangaInfo)
	spinnerChapInfo, _ := pterm.DefaultSpinner.Start("Fetching chapter info...")
	chapterFullInfo, err := client.
		GetLastChapterFullInfo(mangaId, p.language, p.translateGroup)
	if err != nil {
		spinnerChapInfo.Fail("Failed to get chapter info")
		e.Printf("While getting manga chapters: %v\n", err)
		os.Exit(1)
	}
	spinnerChapInfo.Success("Fetched chapter info")

	p.chapters = []mangadexapi.ChapterFullInfo{chapterFullInfo}

	p.downloadChapters()
}

func (p dlParam) DownloadChapters(mangaId string) {
	spinnerMangaInfo, _ := pterm.DefaultSpinner.Start("Fetching manga info...")
	mangaInfo, err := p.getMangaInfo(mangaId)
	if err != nil {
		spinnerMangaInfo.Fail("Failed to get manga info")
		os.Exit(1)
	}
	spinnerMangaInfo.Success("Fetched manga info")
	p.mangaInfo = mangaInfo

	spinnerChapInfo, _ := pterm.DefaultSpinner.Start("Fetching chapters info...")
	p.chapters, err = client.GetFullChaptersInfo(mangaId, p.language, p.translateGroup,
		p.lowestChapter, p.highestChapter)
	if err != nil {
		spinnerChapInfo.Fail("Failed to get chapters info")
		e.Printf("While getting manga chapters: %v\n", err)
		os.Exit(1)
	}
	spinnerChapInfo.Success("Fetched chapters info")

	if len(p.chapters) == 0 {
		e.Printf("Chapters %s not found, try another "+
			"range, language, translation group etc.\n", p.chaptersRange)
		os.Exit(0)
	}

	printShortMangaInfo(mangaInfo)
	if p.isMerge {
		p.downloadMergeChapters()
	} else {
		p.downloadChapters()
	}
}
