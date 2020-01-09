package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: `find novel name`,
	Run:   FindCommand,
}

func FindCommand(cmd *cobra.Command, args []string) {
	for {
		reader := bufio.NewReader(os.Stdin)
		result, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("read err: ", err)
		}
		fmt.Println("result:::", result)

	}
}

func init() {
	RootCmd.AddCommand(findCmd)
}
