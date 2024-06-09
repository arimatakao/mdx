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
	mangaurldl string
	outputDir  string
	language   string
	chapter    int
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVarP(&mangaurldl,
		"url", "u", "", "specify the URL for the manga")
	downloadCmd.Flags().StringVarP(&outputDir,
		"output", "o", ".", "specify output directory for file")
	downloadCmd.Flags().StringVarP(&language,
		"language", "l", "en", "specify language")
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
			GetChaptersList("1", strconv.Itoa(chapter+i), mangaId, language)
		if err != nil {
			spinnerChInfo.Fail("Failed to fetch manga chapters")
			fmt.Printf("error while getting chapters: %v\n", err)
			os.Exit(1)
		}

		if len(chapterList.Data) != 1 {
			fmt.Println("no chapters to download")
			os.Exit(1)
		}

		checkChapter, _ := strconv.Atoi(chapterList.Data[0].Attributes.Chapter)
		if checkChapter > chapter+1 {
			fmt.Println("no chapters to download")
			os.Exit(1)
		}
		if checkChapter == chapter+1 {
			break
		}
	}
	spinnerChInfo.Success("Fetched manga chapters")

	mangaChapter := chapterList.Data[0].Attributes.Chapter
	mangaVolume := chapterList.Data[0].Attributes.Volume
	downloadedChapterId := chapterList.Data[0].ID

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

	dlbar, _ := pterm.DefaultProgressbar.WithMaxWidth(80).
		WithTotal(len(imageList.Chapter.Data)).
		WithTitle("Downloading images...").Start()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()
	for i, fileName := range imageList.Chapter.Data {
		insideFilename := fmt.Sprintf("%s_vol%s_ch%s_%d.png",
			strings.ReplaceAll(mangaTitle, " ", "_"),
			mangaVolume,
			mangaChapter,
			i+1)

		dlbar.UpdateTitle("Downloading " + insideFilename)

		w, err := zipWriter.Create(insideFilename)
		if err != nil {
			dlbar.Increment()
			fmt.Printf("\nerror while creating file in arhive: %v\n", err)
			continue
		}

		image, err := mangadexapi.
			DownloadImage(imageList.BaseURL, imageList.Chapter.Hash, fileName, false)
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
}
