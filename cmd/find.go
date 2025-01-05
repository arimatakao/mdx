package cmd

import (
	"github.com/arimatakao/mdx/internal/mdx"
	"github.com/spf13/cobra"
)

var (
	findCmd = &cobra.Command{
		Use:     "find",
		Aliases: []string{"f", "search", "list", "ls"},
		Short:   "Find manga",
		Long:    "Search and print manga info. Sort by revelance asceding. Best results will be down",
		Run:     find,
	}
	title            string
	isDoujinshiAllow bool
	outputToFile     bool
)

func init() {
	rootCmd.AddCommand(findCmd)

	findCmd.Flags().StringVarP(&title,
		"title", "t", "", "specifies the title of the manga to search for")
	findCmd.Flags().BoolVarP(&isDoujinshiAllow,
		"doujinshi", "d", false, "show doujinshi in list")
	findCmd.Flags().BoolVarP(&outputToFile,
		"outputToFile", "o", false, "Save the search results to a json file.")

	findCmd.MarkFlagRequired("title")
}

func find(cmd *cobra.Command, args []string) {
	mdx.NewFindParams(title, isDoujinshiAllow, outputToFile).Find()
}
