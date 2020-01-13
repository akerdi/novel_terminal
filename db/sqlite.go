package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var (
	create_NOVEL_SITE = "CREATE TABLE IF NOT EXISTS `novelsite` (" +
		"`id` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"`href` VARCHAR(64) NULL," +
		"`title` VARCHAR(64) NULL," +
		"`isParse` NOT NULL DEFAULT False," +
		"`host` VARCHAR(64)," +
		"`kw` VARCHAR(64)," +
		"`createAt` INTEGER NOT NULL" +
		");"
	create_NOVEL_CHAPTER = "CREATE TABLE IF NOT EXISTS `novelchapter` (" +
		"`id` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"`title` VARCHAR(64) NULL," +
		"`host` VARCHAR(64) NOT NULL," +
		"`createAt` INTEGER NOT NULL," +
		"`novelsite_id` INTEGER," +
		"FOREIGN KEY (novelsite_id) REFERENCES novelsite(id)" +
		");"
	create_NOVEL_CONTENT = "CREATE TABLE IF NOT EXISTS `novelcontent` (" +
		"`id` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"`title` VARCHAR(64) NOT NULL," +
		"`content` TEXT NULL," +
		"`createAt` INTEGER NOT NULL," +
		"`novelsite_id` INTEGER ," +
		"`novelchapter_id` INTEGER," +
		"FOREIGN KEY (novelsite_id) REFERENCES novcelsite(id)," +
		"FOREIGN KEY (novelchapter_id) REFERENCES novelchapter(id)" +
		");"
	DBdf *sql.DB
)

func SetUpdateDataBase() {
	fmt.Println("SetUpdateDataBase:", "start")
	db, err := sql.Open("sqlite3", "./novel.db")
	DBdf = db
	checkErr(err)
	// enable foreign_keys
	tx, _ := db.Begin()
	tx.Exec("PRAGMA foreign_keys = ON")

	stmt, err := db.Prepare(create_NOVEL_SITE)
	checkErr(err)
	res, err := stmt.Exec()
	checkErr(err)
	fmt.Println("create novel_site success", res)
	stmt, err = db.Prepare(create_NOVEL_CHAPTER)
	checkErr(err)
	res, err = stmt.Exec()
	checkErr(err)
	fmt.Println("create novel chapter success")
	stmt, err = db.Prepare(create_NOVEL_CONTENT)
	checkErr(err)
	res, err = stmt.Exec()
	checkErr(err)
	fmt.Println("create novel content success")
}
func InsertQuery(query string) (*sql.Stmt, error) {
	stmt, err := DBdf.Prepare(query)
	return stmt, err
}
func ExecWithStmt(stmt *sql.Stmt, param []interface{}) (interface{}, error) {
	res, err := stmt.Exec(param...)
	return res, err
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
