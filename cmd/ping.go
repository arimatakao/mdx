package cmd

import (
	"github.com/arimatakao/mdx/internal/mdx"
	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Healthcheck of MangaDex API",
	Long:  "Check connection to MangaDex API",
	Run:   ping,
}

func init() {
	rootCmd.AddCommand(pingCmd)
}

func ping(cmd *cobra.Command, args []string) {
	mdx.Ping()
}
