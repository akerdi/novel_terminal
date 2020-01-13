package main

import (
	"fmt"
	"novel/db"
	"os"
	"time"

	"novel/cmd"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db.SetUpdateDataBase()
	fmt.Println("Welcome to novel world!")
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	// setupSqlite3()
	// db.SetUpdateDataBase()
	// cmdDoFind()
}

var (
	CREATE_USERINFO   = ""
	CREATE_USERDETAIL = ""
)

func setupSqlite3() {
	db, err := sql.Open("sqlite3", "./foo.db")
	checkErr(err)
	stmt, err := db.Prepare(CREATE_USERINFO)
	checkErr(err)
	res, err := stmt.Exec()
	checkErr(err)
	fmt.Println("----", res)
	stmt, err = db.Prepare(CREATE_USERDETAIL)
	checkErr(err)
	res, err = stmt.Exec()
	checkErr(err)
	fmt.Println("----++++++ ------", res)

	//插入数据
	stmt, err = db.Prepare("INSERT INTO userinfo(username, departname, created) values(?,?,?)")
	checkErr(err)

	res, err = stmt.Exec("astaxie", "研发部门", "2012-12-09")
	checkErr(err)

	id, err := res.LastInsertId()
	checkErr(err)

	fmt.Println(id)
	//更新数据
	stmt, err = db.Prepare("update userinfo set username=? where uid=?")
	checkErr(err)

	res, err = stmt.Exec("astaxieupdate", id)
	checkErr(err)

	affect, err := res.RowsAffected()
	checkErr(err)

	fmt.Println(affect)

	//查询数据
	rows, err := db.Query("SELECT * FROM userinfo")
	checkErr(err)

	for rows.Next() {
		var uid int
		var username string
		var department string
		var created time.Time
		err = rows.Scan(&uid, &username, &department, &created)
		checkErr(err)
		fmt.Println(uid)
		fmt.Println(username)
		fmt.Println(department)
		fmt.Println(created)
	}

	//删除数据
	// stmt, err = db.Prepare("delete from userinfo where uid=?")
	// checkErr(err)

	// res, err = stmt.Exec(id)
	// checkErr(err)

	// affect, err = res.RowsAffected()
	// checkErr(err)

	// fmt.Println(affect)

	db.Close()
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func cmdDoFind() {
	cmd.GotoFind()
}
