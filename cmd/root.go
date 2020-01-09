package cmd

import (
	"fmt"
	"os"

	"novel/rdb"

	"github.com/go-redis/redis"
	"github.com/spf13/cobra"
)

var DBAddr, DBPassword string
var Verbose bool

var redisClient *redis.Client

func init() {
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().StringVar(&DBAddr, "addr", "localhost:6379", "address of Redis database")
	RootCmd.PersistentFlags().StringVar(&DBPassword, "pass", "", "password for Redis database")

	// RootCmd.AddCommand(listCmd)
	// RootCmd.AddCommand(clearCmd)
}

var RootCmd = &cobra.Command{
	Use:              "novel",
	PersistentPreRun: redisConnect,
}

func redisConnect(cmd *cobra.Command, args []string) {
	client, err := rdb.Connect(DBAddr, DBPassword)
	if err != nil {
		fmt.Println("novel failed to connect to redis, configuration is not correct", err.Error())
		os.Exit(1)
	}
	redisClient = client
	fmt.Println("%%%%%%%%", redisClient)
}
