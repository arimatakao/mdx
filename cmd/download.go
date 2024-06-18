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

	downloadCmd.Flags().StringVarP(&outputExt,
		"ext", "e", "cbz", "choose output file format: cbz pdf epub")
	downloadCmd.Flags().StringVarP(&mangaurl,
		"url", "u", "", "specify the URL for the manga")
	downloadCmd.Flags().StringVarP(&outputDir,
		"output", "o", ".", "specify output directory for file")
	downloadCmd.Flags().StringVarP(&language,
		"language", "l", "en", "specify language")
	downloadCmd.Flags().StringVarP(&translateGroup,
		"translated-by", "t", "", "specify part of name translation group")
	downloadCmd.Flags().StringVarP(&chaptersRange,
		"chapter", "c", "1", "specify chapters")
	downloadCmd.Flags().BoolVarP(&isJpgFileFormat,
		"jpg", "j", false, "download compressed images for small output file size")
	downloadCmd.Flags().BoolVarP(&isMergeChapters,
		"merge", "m", false, "merge downloaded chapters into one file")

	e = pterm.Error
	optionPrint = pterm.NewStyle(pterm.FgGreen, pterm.Bold)
	dp = pterm.NewStyle(pterm.FgDefault, pterm.BgDefault)
}

func checkDownloadArgs(cmd *cobra.Command, args []string) {
	if len(args) == 0 && mangaurl == "" {
		cmd.Help()
		os.Exit(0)
	}

	if mangaurl == "" {
		mangaId = mangadexapi.GetMangaIdFromArg(args)
	} else {
		mangaId = mangadexapi.GetMangaIdFromUrl(mangaurl)
	}

	if mangaId == "" {
		e.Println("Malformated URL")
		os.Exit(0)
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

	if isJpgFileFormat {
		imgExt = "jpg"
	}

	if outputExt != filekit.CBZ_EXT &&
		outputExt != filekit.PDF_EXT &&
		outputExt != filekit.EPUB_EXT {
		e.Printf("%s format of file is not supported\n", outputExt)
		os.Exit(0)
	}
}

func downloadManga(cmd *cobra.Command, args []string) {
	c := mangadexapi.NewClient(MDX_USER_AGENT)

	spinnerMangaInfo, _ := pterm.DefaultSpinner.Start("Fetching manga info...")
	resp, err := c.GetMangaInfo(mangaId)
	if err != nil {
		spinnerMangaInfo.Fail("Failed to get manga info")
		e.Println("While getting manga info, maybe you get malformated link")
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

func printShortMangaInfo(i mangadexapi.MangaInfo) {
	optionPrint.Print("Manga title: ")
	dp.Println(i.Title("en"))
	optionPrint.Print("Alt titles: ")
	dp.Println(i.AltTitles())
	optionPrint.Println("Read or Buy here:")
	dp.Println(i.Links())
	dp.Println("==============")
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

func printChapterInfo(c mangadexapi.ChapterFullInfo) {
	tableData := pterm.TableData{
		{optionPrint.Sprint("Chapter"), dp.Sprint(c.Number())},
		{optionPrint.Sprint("Chapter title"), dp.Sprint(c.Title())},
		{optionPrint.Sprint("Volume"), dp.Sprint(c.Volume())},
		{optionPrint.Sprint("Language"), dp.Sprint(c.Language())},
		{optionPrint.Sprint("Translated by"), dp.Sprint(c.Translator())},
		{optionPrint.Sprint("Uploaded by"), dp.Sprint(c.UploadedBy())},
	}
	pterm.DefaultTable.WithData(tableData).Render()
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
		WithTitle("Downloading pages...").Start()
	defer dlbar.Stop()

	for i, imageFile := range files {
		outputImage, err := client.DownloadImage(chapter.DownloadBaseURL,
			chapter.HashId, imageFile, isJpgFileFormat)
		if err != nil {
			dlbar.UpdateTitle("Failed downloading").Stop()
			return err
		}

		pageIndex := i + 1

		insideFilename := fmt.Sprintf("vol%s_ch%s_%d.%s",
			chapter.Volume(),
			strings.ReplaceAll(chapter.Number(), ".", "_"),
			pageIndex,
			imgExt)
		if pageIndex < 10 {
			insideFilename = fmt.Sprintf("vol%s_ch%s_0%d.%s",
				chapter.Volume(),
				strings.ReplaceAll(chapter.Number(), ".", "_"),
				pageIndex,
				imgExt)
		}
		if err := outputFile.AddFile(insideFilename, outputImage); err != nil {
			dlbar.UpdateTitle("Failed downloading").Stop()
			return err
		}
		dlbar.Increment()
	}
	return nil
}
