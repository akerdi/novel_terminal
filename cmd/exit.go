package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var exitCmd = &cobra.Command{
	Use:   "exit",
	Short: `Summary of cache contents`,
	Long:  `Displays a short summary of what is currently cached`,
	Args:  cobra.MaximumNArgs(1),
	Run:   RunCommand,
}

func RunCommand(cmd *cobra.Command, args []string) {
	fmt.Println("1111")
	for {
		break
	}
	fmt.Println("22222")
	os.Exit(1)
}

func init() {
	RootCmd.AddCommand(exitCmd)
}
