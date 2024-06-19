package cmd

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	// general flags
	mangaUrl        string
	mangaId         string
	mangaChapterUrl string
	mangaChapterId  string
)

var (
	versionApp bool
	versionAPI bool

	rootCmd = &cobra.Command{
		Use:   "mdx",
		Short: "manga downloader from MangaDex website",
		Long: `mdx is a command-line interface program for downloading manga from the MangaDex - https://mangadex.org .
The program uses MangaDex API (https://api.mangadex.org/docs) to fetch manga content.`,
		Run: func(cmd *cobra.Command, args []string) {
			if versionApp {
				dp.Printfln(MDX_APP_VERSION)
				os.Exit(0)
			}

			if versionAPI {
				dp.Println(MANGADEX_API_VERSION)
				os.Exit(0)
			}

			cmd.Help()
		},
	}

	// for error print
	e = pterm.Error
	// default print
	dp = pterm.NewStyle(pterm.FgDefault, pterm.BgDefault)
	// for option print
	optionPrint = pterm.NewStyle(pterm.FgGreen, pterm.Bold)
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		e.Printf("While start execute root command: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.Flags().BoolP("help", "h", false, "Help message for toggle")
	rootCmd.Flags().BoolVarP(&versionApp, "version", "v", false, "version of application")
	rootCmd.Flags().BoolVarP(&versionAPI, "version-api", "a", false, "version of API")
}
