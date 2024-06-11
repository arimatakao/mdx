package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

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
	chapter         int
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
	downloadCmd.Flags().IntVarP(&chapter,
		"chapter", "c", 1, "specify chapter")
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

	if chapter < 0 {
		fmt.Println("error: Malformated chapter")
		os.Exit(0)
	}

	if chapter != 0 {
		chapter -= 1
	}

	if isJpgFileFormat {
		imgExt = "jpg"
	}
}

func downloadManga(cmd *cobra.Command, args []string) {
	c := mangadexapi.NewClient(MDX_USER_AGENT)

	spinnerMangaInfo, _ := pterm.DefaultSpinner.Start("Fetching manga info...")
	mangaInfo, err := c.GetMangaInfo(mangaId)
	if err != nil {
		spinnerMangaInfo.Fail("Failed to get manga info")
		fmt.Printf("error while getting manga info: %v\n", err)
		os.Exit(1)
	}
	mangaTitle := mangaInfo.Attributes.Title["en"]
	spinnerMangaInfo.Success("Fetched manga info")

	chapterList := mangadexapi.ResponseChapterList{}
	spinnerChInfo, _ := pterm.DefaultSpinner.Start("Fetching manga chapter...")
	for i := 0; ; i++ {
		chapterList, err = c.
			GetChaptersList("1", strconv.Itoa(chapter+i),
				mangaId, language)
		if err != nil {
			spinnerChInfo.Fail("Failed to fetch manga chapters")
			fmt.Printf("error while getting chapters: %v\n", err)
			os.Exit(0)
		}

		if len(chapterList.Data) != 1 {
			spinnerChInfo.Fail("Failed to fetch manga chapters")
			fmt.Println("no chapters to download")
			os.Exit(0)
		}

		checkChapter, _ := strconv.Atoi(chapterList.FirstChapter())
		if checkChapter > chapter+1 {
			spinnerChInfo.Fail("Failed to fetch manga chapters")
			fmt.Println("no chapters to download")
			os.Exit(0)
		}
		if checkChapter == chapter+1 &&
			(translateGroup == "" || chapterList.IsTranslateGroup(translateGroup)) {
			break
		}
	}
	spinnerChInfo.Success("Fetched manga chapters")

	mangaChapter := chapterList.FirstChapter()
	mangaVolume := chapterList.FirstVolume()
	downloadedChapterId := chapterList.FirstID()

	imageList, err := c.GetChapterImageList(downloadedChapterId)
	if err != nil {
		fmt.Printf("error while getting images of chapter: %v\n", err)
		os.Exit(1)
	}
	err = os.MkdirAll(filepath.Join("", outputDir), os.ModePerm)
	if err != nil {
		fmt.Printf("error while creating : %v\n", err)
		os.Exit(1)
	}

	filename := fmt.Sprintf("%s vol%s ch%s.cbz", mangaTitle, mangaVolume, mangaChapter)

	archive, err := os.Create(filepath.Join(outputDir, filename))
	if err != nil {
		fmt.Printf("error while creating arhive: %v\n", err)
		os.Exit(1)
	}
	defer archive.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Title\t: %s\n", mangaTitle)
	fmt.Fprintf(w, "Chapter title\t: %s\n", chapterList.FirstChapterTitle())
	fmt.Fprintf(w, "Chapter\t: %s\n", mangaChapter)
	fmt.Fprintf(w, "Volume\t: %s\n", mangaVolume)
	fmt.Fprintf(w, "Pages\t: %d\n", chapterList.FirstChapterPages())
	fmt.Fprintf(w, "Language\t: %s\n", chapterList.FirstTranslationLanguage())
	fmt.Fprintf(w, "Translated by\t: %s\n", chapterList.FirstTranslateGroup())
	fmt.Fprintf(w, "Translator description\t: %s\n", chapterList.FirstTranslateGroupDescription())
	w.Flush()

	dlbar, _ := pterm.DefaultProgressbar.
		WithTitle("Downloading pages").WithTotal(len(imageList.Chapter.Data)).Start()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()
	fileImages := imageList.Chapter.Data
	if isJpgFileFormat {
		fileImages = imageList.Chapter.DataSaver
	}
	for i, fileName := range fileImages {
		insideFilename := fmt.Sprintf("%s_vol%s_ch%s_%d.%s",
			strings.ReplaceAll(mangaTitle, " ", "_"),
			mangaVolume,
			mangaChapter,
			i+1,
			imgExt)

		w, err := zipWriter.Create(insideFilename)
		if err != nil {
			dlbar.Increment()
			fmt.Printf("\nerror while creating file in arhive: %v\n", err)
			continue
		}

		image, err := c.DownloadImage(imageList.BaseURL, imageList.Chapter.Hash, fileName,
			isJpgFileFormat)
		if err != nil {
			dlbar.Increment()
			fmt.Printf("\nfailed to download image: %v\n", err)
			continue
		}

		if _, err := io.Copy(w, image); err != nil {
			dlbar.Increment()
			fmt.Printf("\nfailed to copy image in archive: %v\n", err)
			continue
		}
		dlbar.Increment()
	}

	savedDir := ""
	if outputDir == "." {
		wd, _ := os.Getwd()
		savedDir = filepath.Join(wd, outputDir, archive.Name())
	} else {
		savedDir = archive.Name()
	}
	fmt.Printf("Saved in: %s\n", savedDir)
}
