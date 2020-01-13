package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: `list all or list novel name`,
	Run:   ListCommand,
}

func ListCommand(cmd *cobra.Command, args []string) {
	fmt.Println("novelname :::", NovelName)
	if NovelName != "" {
		// 检查db.search site title
	} else {
		// db.search all title
	}
}

func init() {
	RootCmd.AddCommand(listCmd)
}
