package cmd

import (
	"fmt"
	"os"

	"github.com/arimatakao/mdx/mangadexapi"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	infoCmd = &cobra.Command{
		Use:    "info",
		Short:  "Print detailed information about manga",
		PreRun: checkInfoArgs,
		Run:    getInfo,
	}
)

func init() {
	rootCmd.AddCommand(infoCmd)

	infoCmd.Flags().StringVarP(&mangaurl, "url", "u", "", "specify the URL for the manga")
}

func checkInfoArgs(cmd *cobra.Command, args []string) {
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
}

func getInfo(cmd *cobra.Command, args []string) {
	fmt.Println(mangaId)
	c := mangadexapi.NewClient(MDX_USER_AGENT)

	spinner, _ := pterm.DefaultSpinner.Start("Fetching info...")
	info, err := c.GetMangaInfo(mangaId)
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
