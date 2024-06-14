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
		Aliases: []string{"dl"},
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
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().BoolVarP(&isJpgFileFormat,
		"jpg", "j", false, "download compressed images for small archive size (default: false)")
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

	for _, chapter := range chapters {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintf(w, "Chapter\t: %s\n", chapter.Number())
		fmt.Fprintf(w, "Chapter title\t: %s\n", chapter.Title())
		fmt.Fprintf(w, "Volume\t: %s\n", chapter.Volume())
		fmt.Fprintf(w, "Language\t: %s\n", chapter.Language())
		fmt.Fprintf(w, "Translated by\t: %s\n", chapter.Translator())
		fmt.Fprintf(w, "Uploaded by\t: %s\n", chapter.UploadedBy())
		w.Flush()

		dlbar, _ := pterm.DefaultProgressbar.WithTotal(len(chapter.PngFiles)).
			WithTitle("Downloading pages").Start()

		filename := fmt.Sprintf("[%s] %s vol%s ch%s",
			language, mangaInfo.Title("en"), chapter.Volume(), chapter.Number())
		cbzFile, err := filekit.NewCBZFile(outputDir, filename)
		if err != nil {
			dlbar.Increment()
			fmt.Printf("error while creating arhive: %v\n", err)
		}
		defer cbzFile.Close()

		files := chapter.PngFiles
		if isJpgFileFormat {
			files = chapter.JpgFiles
		}
		for i, imageFile := range files {
			outputImage, err := c.DownloadImage(chapter.DownloadBaseURL,
				chapter.HashId, imageFile, isJpgFileFormat)
			if err != nil {
				dlbar.Increment()
				fmt.Printf("\nfailed to download image: %v\n", err)
				continue
			}

			insideFilename := fmt.Sprintf("%s_vol%s_ch%s_%d.%s",
				strings.ReplaceAll(mangaInfo.Title("en"), " ", "_"),
				chapter.Volume(),
				strings.ReplaceAll(chapter.Number(), ".", "_"),
				i+1,
				imgExt)
			if err := cbzFile.AddFile(insideFilename, outputImage); err != nil {
				dlbar.Increment()
				fmt.Printf("\nfailed to copy image in archive: %v\n", err)
				continue
			}
			dlbar.Increment()
		}
		err = cbzFile.
			AddMetadata(metadata.NewCBZMetadata(MDX_USER_AGENT, mangaInfo, chapter))
		if err != nil {
			fmt.Printf("error while adding metadata to file: %v\n", err)
		}
	}
}
