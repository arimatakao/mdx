package cmd

import (
	"os"

	"github.com/arimatakao/mdx/internal/mdx"
	"github.com/arimatakao/mdx/mangadexapi"
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
	mdx.NewInfoParams(mangaId).GetInfo()
}
