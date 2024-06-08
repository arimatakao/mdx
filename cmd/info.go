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
		Use:   "info",
		Short: "Print detailed information about manga",
		Run:   getInfo,
	}
	mangaurl string
)

func init() {
	rootCmd.AddCommand(infoCmd)

	infoCmd.PersistentFlags().StringVarP(&mangaurl,
		"url", "u", "", "specify the URL for the manga")

	infoCmd.MarkPersistentFlagRequired("url")
}

func getInfo(cmd *cobra.Command, args []string) {
	parsedUrl, err := url.Parse(mangaurl)
	if err != nil {
		fmt.Println("error: Malfomated URL")
		os.Exit(1)
	}
	paths := strings.Split(parsedUrl.Path, "/")
	if len(paths) < 3 {
		fmt.Println("error: Malfomated URL")
		os.Exit(1)
	}

	mangaid := paths[2]

	c := mangadexapi.NewClient("")

	spinner, _ := pterm.DefaultSpinner.Start("Fetching info...")
	info, err := c.GetMangaInfo(mangaid)
	if err != nil {
		spinner.Fail("Failed to fetch manga info")
		fmt.Printf("error while getting info: %v", err)
		os.Exit(1)
	}
	spinner.Success("Fetched info")

	altTitles := []string{}
	for _, m := range info.Attributes.AltTitles {
		for language, title := range m {
			altTitles = append(altTitles, fmt.Sprintf("%s (%s)", title, language))
		}
	}
	joinedAltTitles := strings.Join(altTitles, " | ")

	tags := []string{}
	for _, tagEntity := range info.Attributes.Tags {
		if tagEntity.Type == "tag" {
			tags = append(tags, tagEntity.Attributes.Name["en"])
		}
	}
	joinedTags := strings.Join(tags, ", ")

	authors := []string{}
	artists := []string{}
	for _, relation := range info.Relationships {
		if relation.Type == "author" {
			authors = append(authors, relation.Attributes.Name)
		} else if relation.Type == "artist" {
			artists = append(artists, relation.Attributes.Name)
		}
	}
	joinedAuthors := strings.Join(authors, ", ")
	joinedArtists := strings.Join(artists, ", ")

	fmt.Println("Title: ", info.Attributes.Title["en"])
	fmt.Printf("Alternative titles: %s\n", joinedAltTitles)
	fmt.Println("Type: ", info.Type)
	fmt.Println("Authors: ", joinedAuthors)
	fmt.Println("Artists: ", joinedArtists)
	fmt.Printf("Year: %d\n", info.Attributes.Year)
	fmt.Printf("Tags: %s\n", joinedTags)
	fmt.Println("Status: ", info.Attributes.Status)
	fmt.Println("Last chapter: ", info.Attributes.LastChapter)
	fmt.Println("Original language: ", info.Attributes.OriginalLanguage)
	fmt.Printf("Translated: %v\n", info.Attributes.AvailableTranslatedLanguages)
	fmt.Printf("Link: https://mangadex.org/title/%s\n", info.ID)
	fmt.Println("Description:")
	fmt.Println(info.Attributes.Description["en"])
}
