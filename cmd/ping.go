package cmd

import (
	"github.com/arimatakao/mdx/mangadexapi"
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
	c := mangadexapi.NewClient(MDX_USER_AGENT)
	isAlive := c.Ping()

	if isAlive {
		dp.Println("MangaDex API is alive")
	} else {
		dp.Println("MangaDex API is NOT alive")
	}
}
