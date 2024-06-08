package cmd

import (
	"fmt"

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
	c := mangadexapi.NewClient("")
	isAlive := c.Ping()

	if isAlive {
		fmt.Println("MangaDex API is alive")
	} else {

		fmt.Println("MangaDex API is NOT alive")
	}
}
