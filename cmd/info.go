package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/arimatakao/mdx/mangadexapi"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	infoCmd = &cobra.Command{
		Use:     "info",
		Aliases: []string{"list"},
		Short:   "Print detailed information about manga",
		Run:     getInfo,
	}
	mangaurl string
)

func init() {
	rootCmd.AddCommand(infoCmd)

	infoCmd.Flags().StringVarP(&mangaurl, "url", "u", "", "specify the URL for the manga")

	infoCmd.MarkFlagRequired("url")
}

func getInfo(cmd *cobra.Command, args []string) {
	parsedUrl, err := url.Parse(mangaurl)
	if err != nil {
		fmt.Println("error: Malformated URL")
		os.Exit(0)
	}
	paths := strings.Split(parsedUrl.Path, "/")
	if len(paths) < 3 {
		fmt.Println("error: Malformated URL")
		os.Exit(0)
	}

	mangaid := paths[2]

	c := mangadexapi.NewClient("")

	spinner, _ := pterm.DefaultSpinner.Start("Fetching info...")
	info, err := c.GetMangaInfo(mangaid)
	if err != nil {
		spinner.Fail("Failed to fetch manga info")
		fmt.Printf("error while getting info: %v\n", err)
		os.Exit(1)
	}
	spinner.Success("Fetched info")
	printMangaInfo(info)
}

func printMangaInfo(i mangadexapi.MangaInfo) {
	fmt.Println("Title: ", i.Attributes.Title["en"])
	fmt.Printf("Alternative titles: %s\n", i.GetAltTitles())
	fmt.Println("Type: ", i.Type)
	fmt.Println("Authors: ", i.GetAuthors())
	fmt.Println("Artists: ", i.GetArtists())
	fmt.Printf("Year: %d\n", i.Attributes.Year)
	fmt.Printf("Tags: %s\n", i.GetTags())
	fmt.Println("Status: ", i.Attributes.Status)
	fmt.Println("Last chapter: ", i.Attributes.LastChapter)
	fmt.Println("Original language: ", i.Attributes.OriginalLanguage)
	fmt.Printf("Translated: %v\n", i.Attributes.AvailableTranslatedLanguages)
	fmt.Printf("Link: https://mangadex.org/title/%s\n", i.ID)
	fmt.Println("Description:")
	fmt.Println(i.Attributes.Description["en"])
}
