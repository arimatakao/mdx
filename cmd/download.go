package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/arimatakao/mdx/filekit"
	"github.com/arimatakao/mdx/internal/mdx"
	"github.com/arimatakao/mdx/mangadexapi"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	downloadCmd = &cobra.Command{
		Use:     "download",
		Aliases: []string{"dl", "save", "sv"},
		Short:   "Download manga by URL",
		PreRun:  checkDownloadArgs,
		Run:     downloadManga,
	}
	isJpgFileFormat   bool
	outputDir         string
	language          string
	translateGroup    string
	volumesRange      string
	chaptersRange     string
	lowestChapter     int
	highestChapter    int
	lowestVolume      int
	highestVolume     int
	isMergeChapters   bool
	outputExt         string
	isLastChapter     bool
	isAllChapters     bool
	isVolume          bool
	isInteractiveMode bool
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVarP(&mangaUrl,
		"url", "u", "", "specify the URL for the manga")
	downloadCmd.Flags().StringVarP(&mangaChapterUrl,
		"this", "s", "", "specify the direct URL to a specific chapter")
	downloadCmd.Flags().StringVarP(&outputExt,
		"ext", "e", "pdf", "choose output file format: pdf cbz epub")
	downloadCmd.Flags().StringVarP(&outputDir,
		"output", "o", ".", "specify output directory for file")
	downloadCmd.Flags().StringVarP(&language,
		"language", "l", "en", "specify language")
	downloadCmd.Flags().StringVarP(&translateGroup,
		"translated-by", "t", "", "specify a part of the translation group's name")
	downloadCmd.Flags().StringVarP(&chaptersRange,
		"chapter", "c", "1", "specify chapters")
	downloadCmd.Flags().StringVarP(&volumesRange,
		"volume", "v", "", "specify volumes")
	downloadCmd.Flags().BoolVarP(&isAllChapters,
		"all", "a", false, "download all chapters")
	downloadCmd.Flags().BoolVarP(&isJpgFileFormat,
		"jpg", "j", false, "download compressed images for small output file size")
	downloadCmd.Flags().BoolVarP(&isMergeChapters,
		"merge", "m", false, "merge downloaded chapters into one file. If used with `--volume` or `-v,` it will merge the chapters into their volumes")
	downloadCmd.Flags().BoolVarP(&isLastChapter,
		"last", "", false, "download last chapter")
	downloadCmd.Flags().BoolVarP(&isInteractiveMode,
		"interactive", "i", false, "interactive download mode")
}

func checkDownloadArgs(cmd *cobra.Command, args []string) {
	urlErrorMessage := "Malformatted URL."
	if isInteractiveMode {
		return
	}

	if len(args) == 0 && mangaUrl == "" && mangaChapterUrl == "" {
		cmd.Help()
		os.Exit(0)
	}

	if mangaUrl == "" {
		mangaId = mangadexapi.GetMangaIdFromArgs(args)
	} else {
		mangaId = mangadexapi.GetMangaIdFromUrl(mangaUrl)
	}

	if isLastChapter && mangaId == "" {
		e.Println(urlErrorMessage)
		os.Exit(0)
	}

	if isAllChapters && mangaId == "" {
		e.Println(urlErrorMessage)
		os.Exit(0)
	}

	if mangaChapterUrl == "" {
		mangaChapterId = mangadexapi.GetChapterIdFromArgs(args)
	} else {
		mangaChapterId = mangadexapi.GetChapterIdFromUrl(mangaChapterUrl)
	}

	if mangaId == "" && mangaChapterId == "" {
		e.Println(urlErrorMessage)
		os.Exit(0)
	}

	if filekit.IsNotSupported(outputExt) {
		e.Printfln("%s format of file is not supported", outputExt)
		os.Exit(0)
	}

	if mangaChapterId != "" || isLastChapter {
		return
	}

	if volumesRange != "" {
		isVolume = true
		lowestVolume, highestVolume = parseRange(volumesRange)
		return
	}

	if chaptersRange != "" {
		lowestChapter, highestChapter = parseRange(chaptersRange)
		return
	}
}

func parseRange(rangeStr string) (low, high int) {
	errorMsg := pterm.Sprintf("Malformatted downloading range format %s", rangeStr)

	single, err := strconv.Atoi(rangeStr)
	if err == nil {
		if single < 0 {
			e.Println(errorMsg)
			os.Exit(0)
		}

		return single, single
	}

	nums := strings.Split(rangeStr, "-")
	if len(nums) != 2 {
		e.Println(errorMsg)
		os.Exit(0)
	}

	lowest, err := strconv.Atoi(nums[0])
	if err != nil {
		e.Println(errorMsg)
		os.Exit(0)
	}
	highest, err := strconv.Atoi(nums[1])
	if err != nil {
		e.Println(errorMsg)
		os.Exit(0)
	}

	if lowest >= highest {
		e.Println(errorMsg)
		os.Exit(0)
	}

	return lowest, highest
}

func downloadManga(cmd *cobra.Command, args []string) {
	params := mdx.NewDownloadParam(
		chaptersRange, volumesRange, lowestChapter, highestChapter, lowestVolume, highestVolume,
		language, translateGroup, outputDir, outputExt,
		isJpgFileFormat, isMergeChapters, isVolume, isAllChapters, isLastChapter)

	if isInteractiveMode {
		params.RunInteractiveDownload()
	} else {
		params.RunDownload(mangaId, mangaChapterId)
	}
}
