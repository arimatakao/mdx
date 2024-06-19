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
	optionPrint.Print("Link: ")
	dp.Printfln("https://mangadex.org/title/%s", i.ID)
	optionPrint.Print("Title: ")
	dp.Println(i.Title("en"))
	optionPrint.Print("Alternative titles: ")
	dp.Println(i.AltTitles())
	optionPrint.Print("Type: ")
	dp.Println(i.Type)
	optionPrint.Print("Authors: ")
	dp.Println(i.Authors())
	optionPrint.Print("Artists: ")
	dp.Println(i.Artists())
	optionPrint.Print("Year: ")
	dp.Println(i.Year())
	optionPrint.Print("Status: ")
	dp.Println(i.Status())
	optionPrint.Print("Original language: ")
	dp.Println(i.OriginalLanguage())
	optionPrint.Print("Translated: ")
	dp.Println(i.TranslatedLanguages())
	optionPrint.Print("Tags: ")
	dp.Println(i.Tags())
	optionPrint.Println("Description: ")
	dp.Println(i.Description("en"))
	optionPrint.Println("Read or Buy here:")
	dp.Println(i.Links())
}
