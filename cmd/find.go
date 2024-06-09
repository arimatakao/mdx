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
		Aliases: []string{"f", "search"},
		Short:   "Find manga",
		Long:    "Search and print manga info",
		Run:     find,
	}
	title string
)

func init() {
	rootCmd.AddCommand(findCmd)

	findCmd.Flags().StringVarP(&title,
		"title", "t", "", "specifies the title of the manga to search for")

	findCmd.MarkFlagRequired("title")
}

func find(cmd *cobra.Command, args []string) {
	c := mangadexapi.NewClient("")

	spinner, _ := pterm.DefaultSpinner.Start("Searching manga...")

	response, err := c.Find(title, "15", "0")
	if err != nil {
		spinner.Fail("Failed to search manga")
		fmt.Printf("error while search manga: %v\n", err)
		os.Exit(1)
	}

	if response.Total == 0 {
		spinner.Warning("Nothing found...")
		os.Exit(0)
	}

	spinner.Success("Manga found!")
	fmt.Printf("\nTotal found: %d\n", response.Total)

	for _, m := range response.Data {
		fmt.Println("------------------------------")
		fmt.Println("Title: ", m.Attributes.Title["en"])
		fmt.Println("Type: ", m.Type)
		fmt.Printf("Year: %d\n", m.Attributes.Year)
		fmt.Println("Last chapter: ", m.Attributes.LastChapter)
		fmt.Println("Status: ", m.Attributes.Status)
		fmt.Printf("Translated: %v\n", m.Attributes.AvailableTranslatedLanguages)
		fmt.Printf("Link: https://mangadex.org/title/%s\n", m.ID)
		fmt.Println("Description:")
		fmt.Println(m.Attributes.Description["en"])
	}
}
