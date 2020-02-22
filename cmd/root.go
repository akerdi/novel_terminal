package cmd

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/spf13/cobra"
)

var DBAddr, DBPassword string
var Verbose bool
var NovelName string

var redisClient *redis.Client

func init() {
	RootCmd.PersistentFlags().StringVar(&NovelName, "novelname", "", "搜索的小说名")
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().StringVar(&DBAddr, "addr", "localhost:6379", "address of Redis database")
	RootCmd.PersistentFlags().StringVar(&DBPassword, "pass", "", "password for Redis database")
}

var RootCmd = &cobra.Command{
	Use:              "novel",
	PersistentPreRun: dbConnect,
}

func dbConnect(cmd *cobra.Command, args []string) {
	// client, err := rdb.Connect(DBAddr, DBPassword)
	// if err != nil {
	// 	fmt.Println("novel failed to connect to redis, configuration is not correct", err.Error())
	// 	os.Exit(1)
	// }
	// redisClient = client
	// fmt.Println("%%%%%%%%", redisClient)
	fmt.Println("database setup")
}
