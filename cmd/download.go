package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/arimatakao/mdx/filekit"
	"github.com/arimatakao/mdx/internal/mdx"
	"github.com/arimatakao/mdx/mangadexapi"
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
	isJpgFileFormat bool
	outputDir       string
	language        string
	translateGroup  string
	chaptersRange   string
	lowestChapter   int
	highestChapter  int
	isMergeChapters bool
	outputExt       string
	isLastChapter   bool
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
	downloadCmd.Flags().BoolVarP(&isLastChapter,
		"last", "", false, "download last chapter")
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

	if isLastChapter && mangaId == "" {
		e.Println("Malformated URL")
		os.Exit(0)
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

	if outputExt != filekit.CBZ_EXT &&
		outputExt != filekit.PDF_EXT &&
		outputExt != filekit.EPUB_EXT {
		e.Printfln("%s format of file is not supported", outputExt)
		os.Exit(0)
	}

	if mangaChapterId != "" || isLastChapter {
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
	params := mdx.NewDownloadParam(chaptersRange, lowestChapter, highestChapter, language,
		translateGroup, outputDir, outputExt, isJpgFileFormat, isMergeChapters)
	if mangaChapterId != "" {
		params.DownloadSpecificChapter(mangaChapterId)
	} else if isLastChapter {
		params.DownloadLastChapter(mangaId)
	} else {
		params.DownloadChapters(mangaId)
	}
}
