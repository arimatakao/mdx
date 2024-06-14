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
	resp, err := c.GetMangaInfo(mangaId)
	if err != nil {
		spinner.Fail("Failed to fetch manga info")
		fmt.Printf("error while getting info: %v\n", err)
		os.Exit(1)
	}
	spinner.Success("Fetched info")
	printMangaInfo(resp.MangaInfo())
}

func printMangaInfo(i mangadexapi.MangaInfo) {
	fmt.Println("Title: ", i.Title("en"))
	fmt.Printf("Alternative titles: %s\n", i.AltTitles())
	fmt.Println("Type: ", i.Type)
	fmt.Println("Authors: ", i.Authors())
	fmt.Println("Artists: ", i.Artists())
	fmt.Printf("Year: %d\n", i.Year())
	fmt.Println("Status: ", i.Status())
	fmt.Println("Original language: ", i.OriginalLanguage())
	fmt.Printf("Translated: %v\n", i.TranslatedLanguages())
	fmt.Printf("Tags: %s\n", i.Tags())
	fmt.Println("Description:")
	fmt.Println(i.Description("en"))
	fmt.Println("---")
	fmt.Println("Read or Buy here:")
	for _, v := range i.Links() {
		fmt.Println(v)
	}
}
