package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/arimatakao/mdx/filekit"
	"github.com/arimatakao/mdx/filekit/metadata"
	"github.com/arimatakao/mdx/mangadexapi"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	downloadCmd = &cobra.Command{
		Use:     "download",
		Aliases: []string{"dl", "save"},
		Short:   "Download manga by URL",
		PreRun:  checkDownloadArgs,
		Run:     downloadManga,
	}
	imgExt          string = "png"
	isJpgFileFormat bool
	outputDir       string
	language        string
	translateGroup  string
	chaptersRange   string
	lowestChapter   int
	highestChapter  int
	isMergeChapters bool
	outputExt       string
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVarP(&mangaUrl,
		"url", "u", "", "specify the URL for the manga")
	downloadCmd.Flags().StringVarP(&mangaChapterUrl,
		"this", "s", "", "specify the direct URL to a specific chapter")
	downloadCmd.Flags().StringVarP(&outputExt,
		"ext", "e", "cbz", "choose output file format: cbz pdf epub")
	downloadCmd.Flags().StringVarP(&outputDir,
		"output", "o", ".", "specify output directory for file")
	downloadCmd.Flags().StringVarP(&language,
		"language", "l", "en", "specify language")
	downloadCmd.Flags().StringVarP(&translateGroup,
		"translated-by", "t", "", "specify a part of the translation group's name")
	downloadCmd.Flags().StringVarP(&chaptersRange,
		"chapter", "c", "1", "specify chapters")
	downloadCmd.Flags().BoolVarP(&isJpgFileFormat,
		"jpg", "j", false, "download compressed images for small output file size")
	downloadCmd.Flags().BoolVarP(&isMergeChapters,
		"merge", "m", false, "merge downloaded chapters into one file")

	e = pterm.Error
	dp = pterm.NewStyle(pterm.FgDefault, pterm.BgDefault)
	field = pterm.NewStyle(pterm.FgGreen, pterm.BgDefault, pterm.Bold)
}

func checkDownloadArgs(cmd *cobra.Command, args []string) {
	if len(args) == 0 && mangaUrl == "" && mangaChapterUrl == "" {
		cmd.Help()
		os.Exit(0)
	}

	if mangaUrl == "" {
		mangaId = mangadexapi.GetMangaIdFromArgs(args)
	} else {
		mangaId = mangadexapi.GetMangaIdFromUrl(mangaUrl)
	}

	if mangaChapterUrl == "" {
		mangaChapterId = mangadexapi.GetChapterIdFromArgs(args)
	} else {
		mangaChapterId = mangadexapi.GetChapterIdFromUrl(mangaChapterUrl)
	}

	if mangaId == "" && mangaChapterId == "" {
		e.Println("Malformated URL")
		os.Exit(0)
	}

	if isJpgFileFormat {
		imgExt = "jpg"
	}

	if outputExt != filekit.CBZ_EXT &&
		outputExt != filekit.PDF_EXT &&
		outputExt != filekit.EPUB_EXT {
		e.Printfln("%s format of file is not supported", outputExt)
		os.Exit(0)
	}

	if mangaChapterId != "" {
		return
	}

	singleChapter, err := strconv.Atoi(chaptersRange)
	if err == nil {
		if singleChapter < 0 {
			e.Println("Malformated chapters format")
			os.Exit(0)
		}

		lowestChapter = singleChapter
		highestChapter = singleChapter
	} else if nums := strings.Split(chaptersRange, "-"); len(nums) == 2 {
		lowest, err := strconv.Atoi(nums[0])
		if err != nil {
			e.Println("Malformated chapters format")
			os.Exit(0)
		}

		highest, err := strconv.Atoi(nums[1])
		if err != nil {
			e.Println("Malformated chapters format")
			os.Exit(0)
		}

		if lowest >= highest {
			e.Println("Malformated chapters format")
			os.Exit(0)
		}

		lowestChapter = lowest
		highestChapter = highest

	} else {
		e.Println("Malformated chapters format")
		os.Exit(0)
	}

}

func downloadManga(cmd *cobra.Command, args []string) {
	if mangaChapterId != "" {
		downloadSingleChapter()
		return
	}

	c := mangadexapi.NewClient(MDX_USER_AGENT)

	spinnerMangaInfo, _ := pterm.DefaultSpinner.Start("Fetching manga info...")
	resp, err := c.GetMangaInfo(mangaId)
	if err != nil {
		spinnerMangaInfo.Fail("Failed to get manga info")
		e.Println("While getting manga info, maybe you set malformated link")
		os.Exit(1)
	}
	mangaInfo := resp.MangaInfo()
	spinnerMangaInfo.Success("Fetched manga info")

	spinnerChapInfo, _ := pterm.DefaultSpinner.Start("Fetching chapters info...")
	chapters, err := c.GetFullChaptersInfo(mangaId, language, translateGroup,
		lowestChapter, highestChapter)
	if err != nil {
		spinnerChapInfo.Fail("Failed to get manga info")
		e.Printf("While getting manga chapters: %v\n", err)
		os.Exit(1)
	}
	spinnerChapInfo.Success("Fetched chapters info")

	if len(chapters) == 0 {
		e.Printf("Chapters %s not found, try another "+
			"range, language, translation group etc.\n", chaptersRange)
		os.Exit(0)
	}

	printShortMangaInfo(mangaInfo)

	if isMergeChapters {
		downloadMergeChapters(c,
			mangaInfo, chapters, outputExt, MDX_USER_AGENT, isJpgFileFormat)
	} else {
		downloadChapters(c,
			mangaInfo, chapters, outputExt, MDX_USER_AGENT, isJpgFileFormat)
	}
}

func downloadSingleChapter() {
	c := mangadexapi.NewClient(MDX_USER_AGENT)
	spinnerChapInfo, _ := pterm.DefaultSpinner.Start("Fetching chapter info...")
	resp, err := c.GetChapterInfo(mangaChapterId)
	if err != nil {
		spinnerChapInfo.Fail("Failed to get chapter info")
		os.Exit(1)
	}
	chapterInfo := resp.GetChapterInfo()
	chapterFullInfo, err := c.GetChapterImagesInFullInfo(chapterInfo)
	if err != nil {
		spinnerChapInfo.Fail("Failed to get chapter info")
		os.Exit(1)
	}
	spinnerChapInfo.Success("Fetched chapter info")

	mangaId := chapterInfo.GetMangaId()
	spinnerMangaInfo, _ := pterm.DefaultSpinner.Start("Fetching manga info...")
	respManga, err := c.GetMangaInfo(mangaId)
	if err != nil {
		spinnerMangaInfo.Fail("Failed to get manga info")
		os.Exit(1)
	}
	mangaInfo := respManga.MangaInfo()
	spinnerMangaInfo.Success("Fetched manga info")
	chapArr := []mangadexapi.ChapterFullInfo{chapterFullInfo}
	printShortMangaInfo(mangaInfo)
	downloadChapters(c, mangaInfo, chapArr, outputExt, MDX_USER_AGENT, isJpgFileFormat)
}

func printShortMangaInfo(i mangadexapi.MangaInfo) {
	dp.Println(field.Sprint("Manga title: "), i.Title("en"))
	dp.Println(field.Sprint("Alt titles: "), i.AltTitles())
	field.Println("Read or Buy here:")
	dp.Println(i.Links())
	dp.Println("==============")
}

func printChapterInfo(c mangadexapi.ChapterFullInfo) {
	tableData := pterm.TableData{
		{field.Sprint("Chapter"), dp.Sprint(c.Number())},
		{field.Sprint("Chapter title"), dp.Sprint(c.Title())},
		{field.Sprint("Volume"), dp.Sprint(c.Volume())},
		{field.Sprint("Language"), dp.Sprint(c.Language())},
		{field.Sprint("Translated by"), dp.Sprint(c.Translator())},
		{field.Sprint("Uploaded by"), dp.Sprint(c.UploadedBy())},
	}
	pterm.DefaultTable.WithData(tableData).Render()
}

func downloadMergeChapters(client mangadexapi.Clientapi,
	mangaInfo mangadexapi.MangaInfo,
	chapters []mangadexapi.ChapterFullInfo,
	outputExtension, userAgent string,
	isJpg bool) {

	containerFile, err := filekit.NewContainer(outputExtension)
	if err != nil {
		e.Printf("While creating output file: %v\n", err)
		os.Exit(1)
	}

	for _, chapter := range chapters {
		printChapterInfo(chapter)

		err = downloadProcess(client, chapter, containerFile, isJpg)
		if err != nil {
			e.Printf("While downloading chapter: %v\n", err)
			os.Exit(1)
		}
	}

	filename := fmt.Sprintf("[%s] %s ch%s",
		language, mangaInfo.Title("en"), chaptersRange)
	metaInfo := metadata.NewMetadata(userAgent, mangaInfo, chapters[0])
	err = containerFile.WriteOnDiskAndClose(outputDir, filename, metaInfo, chaptersRange)
	if err != nil {
		e.Printf("While saving %s on disk: %v\n", filename, err)
		os.Exit(1)
	}
}

func downloadChapters(client mangadexapi.Clientapi,
	mangaInfo mangadexapi.MangaInfo,
	chapters []mangadexapi.ChapterFullInfo,
	outputExtension, userAgent string,
	isJpg bool) {

	for _, chapter := range chapters {
		printChapterInfo(chapter)

		containerFile, err := filekit.NewContainer(outputExtension)
		if err != nil {
			e.Printf("While creating output file: %v\n", err)
			os.Exit(1)
		}

		err = downloadProcess(client, chapter, containerFile, isJpg)
		if err != nil {
			e.Printf("While downloading chapter: %v\n", err)
			os.Exit(1)
		}

		filename := fmt.Sprintf("[%s] %s vol%s ch%s",
			language, mangaInfo.Title("en"), chapter.Volume(), chapter.Number())
		metaInfo := metadata.NewMetadata(userAgent, mangaInfo, chapter)
		err = containerFile.WriteOnDiskAndClose(outputDir, filename, metaInfo, "")
		if err != nil {
			e.Printf("While saving %s on disk: %v\n", filename, err)
			os.Exit(1)
		}
	}
}

func downloadProcess(
	client mangadexapi.Clientapi,
	chapter mangadexapi.ChapterFullInfo,
	outputFile filekit.Container, isJpg bool) error {

	files := chapter.PngFiles
	if isJpg {
		files = chapter.JpgFiles
	}

	dlbar, _ := pterm.DefaultProgressbar.WithTotal(len(files)).
		WithTitle("Downloading pages...").
		WithBarStyle(pterm.NewStyle(pterm.FgGreen)).Start()
	defer dlbar.Stop()

	for _, imageFile := range files {
		outputImage, err := client.DownloadImage(chapter.DownloadBaseURL,
			chapter.HashId, imageFile, isJpgFileFormat)
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
	return nil
}
