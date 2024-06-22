package cmd

import (
	"os"

	"github.com/arimatakao/mdx/app"
	"github.com/arimatakao/mdx/internal/mdx"
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
		Short: app.SHORT_DESCRIPTION,
		Long:  app.LONG_DESCRIPTION,
		Run: func(cmd *cobra.Command, args []string) {
			if versionApp {
				mdx.PrintVersion()
			}

			if versionAPI {
				mdx.PrintMangaDexAPIVersion()
			}

			cmd.Help()
		},
	}

	// for error print
	e = pterm.Error
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
