package cmd

import (
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

	infoCmd.Flags().StringVarP(&mangaUrl, "url", "u", "", "specify the URL for the manga")
}

func checkInfoArgs(cmd *cobra.Command, args []string) {
	if len(args) == 0 && mangaUrl == "" {
		cmd.Help()
		os.Exit(0)
	}

	if mangaUrl == "" {
		mangaId = mangadexapi.GetMangaIdFromArgs(args)
	} else {
		mangaId = mangadexapi.GetMangaIdFromUrl(mangaUrl)
	}

	if mangaId == "" {
		e.Printfln("Malformated URL")
		os.Exit(0)
	}
}

func getInfo(cmd *cobra.Command, args []string) {
	c := mangadexapi.NewClient(MDX_USER_AGENT)

	spinner, _ := pterm.DefaultSpinner.Start("Fetching info...")
	resp, err := c.GetMangaInfo(mangaId)
	if err != nil {
		spinner.Fail("Failed to fetch manga info")
		e.Printfln("While getting manga information: %v\n", err)
		os.Exit(1)
	}
	spinner.Success("Fetched info")
	printMangaInfo(resp.MangaInfo())
}

func printMangaInfo(i mangadexapi.MangaInfo) {
	dp.Println(field.Sprint("Link: "), dp.Sprintf("https://mangadex.org/title/%s", i.ID))
	dp.Println(field.Sprint("Title: "), i.Title("en"))
	dp.Println(field.Sprint("Alternative titles: "), i.Title("en"))
	dp.Println(field.Sprint("Type: "), i.Type)
	dp.Println(field.Sprint("Authors: "), i.Authors())
	dp.Println(field.Sprint("Artists: "), i.Artists())
	dp.Println(field.Sprint("Year: "), i.Year())
	dp.Println(field.Sprint("Status: "), i.Status())
	dp.Println(field.Sprint("Original language: "), i.OriginalLanguage())
	dp.Println(field.Sprint("Translated: "), i.TranslatedLanguages())
	dp.Println(field.Sprint("Tags: "), i.Tags())
	dp.Println(field.Sprint("Description:\n"), i.Description("en"))
	dp.Println(field.Sprint("Read or Buy here:\n"), i.Links())
}
