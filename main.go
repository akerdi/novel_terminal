package main

import (
	"fmt"
	"novel/db"
	"os"

	"novel/cmd"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db.SetUpdateDataBase()
	fmt.Println("Welcome to novel world! \n  有疑问或者问题可以联系QQ mail: 767838865@qq.com")
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	// setupSqlite3()
	// db.SetUpdateDataBase()
	// cmdDoFind()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func cmdDoFind() {
	cmd.GotoFind()
}
