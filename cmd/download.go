package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/arimatakao/mdx/mangadexapi"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	downloadCmd = &cobra.Command{
		Use:     "download",
		Aliases: []string{"dl"},
		Short:   "Download manga by URL",
		Run:     downloadManga,
	}
	isJpgFileFormat bool
	mangaurldl      string
	outputDir       string
	language        string
	translateGroup  string
	chapter         int
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().BoolVarP(&isJpgFileFormat,
		"jpg", "j", false, "download compressed images for small archive size (default: false)")
	downloadCmd.Flags().StringVarP(&mangaurldl,
		"url", "u", "", "specify the URL for the manga")
	downloadCmd.Flags().StringVarP(&outputDir,
		"output", "o", ".", "specify output directory for file")
	downloadCmd.Flags().StringVarP(&language,
		"language", "l", "en", "specify language")
	downloadCmd.Flags().StringVarP(&translateGroup,
		"translated-by", "t", "", "specify part of name translation group")
	downloadCmd.Flags().IntVarP(&chapter,
		"chapter", "c", 1, "specify chapter")

	downloadCmd.MarkFlagRequired("url")
}

func downloadManga(cmd *cobra.Command, args []string) {
	parsedUrl, err := url.Parse(mangaurldl)
	if err != nil {
		fmt.Println("error: Malformated URL")
		os.Exit(1)
	}

	paths := strings.Split(parsedUrl.Path, "/")
	if len(paths) < 3 {
		fmt.Println("error: Malformated URL")
		os.Exit(1)
	}

	mangaId := paths[2]

	if chapter < 0 {
		fmt.Println("error: Malformated chapter")
		os.Exit(1)
	}

	if chapter != 0 {
		chapter -= 1
	}

	imgExt := "png"
	if isJpgFileFormat {
		imgExt = "jpg"
	}

	c := mangadexapi.NewClient("")

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

	fmt.Println("Information about manga chapter:")
	fmt.Printf("\tTitle: %s\n", mangaTitle)
	fmt.Printf("\tVolume: %s\n", mangaVolume)
	fmt.Printf("\tChapter: %s\n", mangaChapter)
	fmt.Printf("\tPages: %d\n", chapterList.FirstChapterPages())
	fmt.Printf("\tLanguage: %s\n", chapterList.FirstTranslationLanguage())
	fmt.Printf("\tTranslated by: %s\n", chapterList.FirstTranslateGroup())
	fmt.Printf("\tTranslator description: %s\n", chapterList.FirstTranslateGroupDescription())

	dlbar, _ := pterm.DefaultProgressbar.WithMaxWidth(80).
		WithTotal(len(imageList.Chapter.Data)).
		WithTitle("Downloading images...").Start()

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

		dlbar.UpdateTitle("Downloading " + insideFilename)

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
	fmt.Printf("Saved in : %s\n", archive.Name())
}
