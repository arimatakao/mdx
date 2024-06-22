package cmd

import (
	"github.com/arimatakao/mdx/internal/mdx"
	"github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Check for updates to application",
		Run:   checkUpdate,
	}
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

func checkUpdate(cmd *cobra.Command, args []string) {
	mdx.CheckUpdate()
}
