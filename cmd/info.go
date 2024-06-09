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
		os.Exit(1)
	}
	paths := strings.Split(parsedUrl.Path, "/")
	if len(paths) < 3 {
		fmt.Println("error: Malformated URL")
		os.Exit(1)
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

	fmt.Println("Title: ", info.Attributes.Title["en"])
	fmt.Printf("Alternative titles: %s\n", info.GetAltTitles())
	fmt.Println("Type: ", info.Type)
	fmt.Println("Authors: ", info.GetAuthors())
	fmt.Println("Artists: ", info.GetArtists())
	fmt.Printf("Year: %d\n", info.Attributes.Year)
	fmt.Printf("Tags: %s\n", info.GetTags())
	fmt.Println("Status: ", info.Attributes.Status)
	fmt.Println("Last chapter: ", info.Attributes.LastChapter)
	fmt.Println("Original language: ", info.Attributes.OriginalLanguage)
	fmt.Printf("Translated: %v\n", info.Attributes.AvailableTranslatedLanguages)
	fmt.Printf("Link: https://mangadex.org/title/%s\n", info.ID)
	fmt.Println("Description:")
	fmt.Println(info.Attributes.Description["en"])
}
