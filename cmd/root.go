package cmd

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/spf13/cobra"
)

// NovelName 用户要搜索的小说名 可以是某个文本，然后采用分词搜索
var NovelName string

var redisClient *redis.Client

func init() {
	RootCmd.PersistentFlags().StringVar(&NovelName, "novelname", "", "搜索的小说名")
}

// RootCmd cobra 对象
var RootCmd = &cobra.Command{
	Use:              "novel",
	PersistentPreRun: dbConnect,
}

func dbConnect(cmd *cobra.Command, args []string) {
	fmt.Println(". . .")
}
