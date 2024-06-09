package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"net/url"
	"os"
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
	outputFile string
	language   string
	chapter    string
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVarP(&mangaurldl,
		"url", "u", "", "specify the URL for the manga")
	downloadCmd.Flags().StringVarP(&outputFile,
		"output", "o", "", "specify output cbz file of manga")
	downloadCmd.Flags().StringVarP(&language,
		"language", "l", "en", "specify language")
	downloadCmd.Flags().StringVarP(&chapter,
		"chapter", "c", "1", "specify chapter")

	downloadCmd.MarkFlagRequired("url")
	downloadCmd.MarkFlagRequired("output")
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

	c := mangadexapi.NewClient("")

	mangaInfo, err := c.GetMangaInfo(mangaId)
	if err != nil {
		fmt.Printf("error while getting manga info: %v\n", err)
		os.Exit(1)
	}
	mangaTitle := mangaInfo.Attributes.Title["en"]

	spinner, _ := pterm.DefaultSpinner.Start("Fetching chapter...")
	list, err := c.GetChaptersList(mangaId, language)
	if err != nil {
		spinner.Fail("Failed to fetch manga chapters")
		fmt.Printf("error while getting chapters: %v\n", err)
		os.Exit(1)
	}
	spinner.Success("Fetched chapters")

	if len(list.Data) == 0 {
		fmt.Println("no chapters to download")
		os.Exit(1)
	}

	firstChapterId := list.Data[0].ID

	imageList, err := c.GetChapterImageList(firstChapterId)
	if err != nil {
		fmt.Printf("error while getting images of chapter: %v\n", err)
		os.Exit(1)
	}
	archive, err := os.Create(outputFile)
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
			list.Data[0].Attributes.Volume,
			list.Data[0].Attributes.Chapter,
			i+1)

		barTitle := fmt.Sprintf("Downloading %s", insideFilename)
		dlbar.UpdateTitle(barTitle)

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
