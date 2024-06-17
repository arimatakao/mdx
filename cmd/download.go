package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

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
		"ext", "e", "cbz", "choose output file format: cbz pdf")
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
		fmt.Println("error: Malformated URL")
		os.Exit(0)
	}

	singleChapter, err := strconv.Atoi(chaptersRange)
	if err == nil {
		if singleChapter < 0 {
			fmt.Println("error: Malformated chapters format")
			os.Exit(0)
		}

		lowestChapter = singleChapter
		highestChapter = singleChapter
	} else if nums := strings.Split(chaptersRange, "-"); len(nums) == 2 {
		lowest, err := strconv.Atoi(nums[0])
		if err != nil {
			fmt.Println("error: Malformated chapters format")
			os.Exit(0)
		}

		highest, err := strconv.Atoi(nums[1])
		if err != nil {
			fmt.Println("error: Malformated chapters format")
			os.Exit(0)
		}

		if lowest >= highest {
			fmt.Println("error: Malformated chapters format")
			os.Exit(0)
		}

		lowestChapter = lowest
		highestChapter = highest

	} else {
		fmt.Println("error: Malformated chapters format")
		os.Exit(0)
	}

	if isJpgFileFormat {
		imgExt = "jpg"
	}

	if outputExt != filekit.CBZ_EXT && outputExt != filekit.PDF_EXT {
		fmt.Printf("error: %s format of file is not supported\n", outputExt)
		os.Exit(0)
	}
}

func downloadManga(cmd *cobra.Command, args []string) {
	c := mangadexapi.NewClient(MDX_USER_AGENT)

	spinnerMangaInfo, _ := pterm.DefaultSpinner.Start("Fetching manga info...")
	resp, err := c.GetMangaInfo(mangaId)
	if err != nil {
		spinnerMangaInfo.Fail("Failed to get manga info")
		fmt.Printf("error while getting manga info: %v\n", err)
		os.Exit(1)
	}
	mangaInfo := resp.MangaInfo()
	spinnerMangaInfo.Success("Fetched manga info")

	spinnerChapInfo, _ := pterm.DefaultSpinner.Start("Fetching chapters info...")
	chapters, err := c.GetFullChaptersInfo(mangaId, language, translateGroup,
		lowestChapter, highestChapter)
	if err != nil {
		spinnerChapInfo.Fail("Failed to get manga info")
		fmt.Printf("error while getting manga chapters: %v\n", err)
		os.Exit(1)
	}
	spinnerChapInfo.Success("Fetched chapters info")

	if len(chapters) == 0 {
		fmt.Printf("chapters %s not found, try another "+
			"range, language, translation group etc.\n", chaptersRange)
		os.Exit(0)
	}

	fmt.Println("Manga title: ", mangaInfo.Title("en"))
	fmt.Println("Alternative title: ", mangaInfo.AltTitles())
	fmt.Println("====")

	if isMergeChapters {
		downloadMergeChapters(c,
			mangaInfo, chapters, outputExt, MDX_USER_AGENT, isJpgFileFormat)
	} else {
		downloadChapters(c,
			mangaInfo, chapters, outputExt, MDX_USER_AGENT, isJpgFileFormat)
	}
}

func downloadMergeChapters(client mangadexapi.Clientapi,
	mangaInfo mangadexapi.MangaInfo,
	chapters []mangadexapi.ChapterFullInfo,
	outputExtension, userAgent string,
	isJpg bool) {

	containerFile, err := filekit.NewContainer(outputExtension)
	if err != nil {
		fmt.Printf("error while creating output file: %v\n", err)
	}

	for _, chapter := range chapters {
		printChapterInfo(chapter)

		err = downloadProcess(client, mangaInfo, chapter, containerFile, isJpg)
		if err != nil {
			fmt.Printf("\nerror while downloading chapter: %v\n", err)
			os.Exit(1)
		}
	}

	firstChapter := chapters[0].Number()
	lastChapter := chapters[len(chapters)-1].Number()

	filename := fmt.Sprintf("[%s] %s ch%s-%s",
		language, mangaInfo.Title("en"), firstChapter, lastChapter)
	metaInfo := metadata.NewMetadata(userAgent, mangaInfo, chapters[0])
	err = containerFile.WriteOnDiskAndClose(outputDir, filename, metaInfo)
	if err != nil {
		fmt.Printf("error while saving %s on disk: %v\n", filename, err)
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
			fmt.Printf("error while creating output file: %v\n", err)
		}

		err = downloadProcess(client, mangaInfo, chapter, containerFile, isJpg)
		if err != nil {
			fmt.Printf("error while downloading chapter: %v\n", err)
		}

		filename := fmt.Sprintf("[%s] %s vol%s ch%s",
			language, mangaInfo.Title("en"), chapter.Volume(), chapter.Number())
		metaInfo := metadata.NewMetadata(userAgent, mangaInfo, chapter)
		err = containerFile.WriteOnDiskAndClose(outputDir, filename, metaInfo)
		if err != nil {
			fmt.Printf("error while saving %s on disk: %v\n", filename, err)
		}
	}
}

func printChapterInfo(c mangadexapi.ChapterFullInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Chapter\t: %s\n", c.Number())
	fmt.Fprintf(w, "Chapter title\t: %s\n", c.Title())
	fmt.Fprintf(w, "Volume\t: %s\n", c.Volume())
	fmt.Fprintf(w, "Language\t: %s\n", c.Language())
	fmt.Fprintf(w, "Translated by\t: %s\n", c.Translator())
	fmt.Fprintf(w, "Uploaded by\t: %s\n", c.UploadedBy())
	w.Flush()
}

func downloadProcess(
	client mangadexapi.Clientapi,
	mangaInfo mangadexapi.MangaInfo,
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

		insideFilename := fmt.Sprintf("%s_vol%s_ch%s_%d.%s",
			strings.ReplaceAll(mangaInfo.Title("en"), " ", "_"),
			chapter.Volume(),
			strings.ReplaceAll(chapter.Number(), ".", "_"),
			i+1,
			imgExt)
		if err := outputFile.AddFile(insideFilename, outputImage); err != nil {
			dlbar.UpdateTitle("Failed downloading").Stop()
			return err
		}
		dlbar.Increment()
	}
	return nil
}
