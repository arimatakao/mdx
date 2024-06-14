package cmd

import (
	"fmt"
	"os"

	"github.com/arimatakao/mdx/mangadexapi"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	findCmd = &cobra.Command{
		Use:     "find",
		Aliases: []string{"f", "search", "list"},
		Short:   "Find manga",
		Long:    "Search and print manga info. Sort by revelance asceding. Best results will be down",
		Run:     find,
	}
	title            string
	isDoujinshiAllow bool
)

func init() {
	rootCmd.AddCommand(findCmd)

	findCmd.Flags().StringVarP(&title,
		"title", "t", "", "specifies the title of the manga to search for")
	findCmd.Flags().BoolVarP(&isDoujinshiAllow,
		"doujinshi", "d", false, "show doujinshi in list")

	findCmd.MarkFlagRequired("title")
}

func find(cmd *cobra.Command, args []string) {
	c := mangadexapi.NewClient(MDX_USER_AGENT)

	spinner, _ := pterm.DefaultSpinner.Start("Searching manga...")
	printedCount := 25
	response, err := c.Find(title, printedCount, 0, isDoujinshiAllow)
	if err != nil {
		spinner.Fail("Failed to search manga")
		fmt.Printf("\nerror while search manga: %v\n", err)
		os.Exit(1)
	}

	if response.Total == 0 {
		spinner.Warning("Nothing found...")
		os.Exit(0)
	}
	spinner.Success("Manga found!")

	fmt.Printf("\nTotal found: %d\n", response.Total)

	for _, m := range response.List() {
		fmt.Println("------------------------------")
		printMangaInfo(m)
	}

	if response.Total > printedCount {
		fmt.Println("==============================")
		fmt.Printf("\nFull results: https://mangadex.org/search?q=%s\n", title)
	}
}
