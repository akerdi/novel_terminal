package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var (
	DBdf              *sql.DB
	create_NOVEL_SITE = `
	CREATE TABLE IF NOT EXISTS novelsite (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		href VARCHAR(64) NULL,
		title VARCHAR(64) NULL,
		isParse NOT NULL DEFAULT False,
		host VARCHAR(64),
		kw VARCHAR(64),
		createAt INTEGER NOT NULL
		);`
	/*
		title 小说名
		chapters 包含章节链接、章节名字的json text
		link_prefix 章节跳转路径拼接逻辑
		origin_url 原始小说链接
		domain 该小说主域名
	*/
	create_NOVEL_CHAPTER = `
	CREATE TABLE IF NOT EXISTS novelchapter (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title VARCHAR(64) NULL,
		chapters TEXT NOT NULL,
		origin_url VARCHAR(64) NOT NULL,
		link_prefix VARCHAR(32) NOT NULL,
		domain VARCHAR(64) NOT NULL,
		createAt INTEGER NOT NULL,
		novelsite_id INTEGER,
		FOREIGN KEY (novelsite_id) REFERENCES novelsite(id)
		);
		CREATE UNIQUE INDEX IF NOT EXISTS origin_url_unique
		ON novelchapter (origin_url);
		`
	/*
		chapter_index 章节索引unique
	*/
	create_NOVEL_CONTENT = `
	CREATE TABLE IF NOT EXISTS novelcontent (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title VARCHAR(64) NOT NULL,
		content TEXT NULL,
		chapter_index INTEGER NOT NULL,
		createAt INTEGER NOT NULL,
		novelsite_id INTEGER ,
		novelchapter_id INTEGER,
		FOREIGN KEY (novelsite_id) REFERENCES novcelsite(id),
		FOREIGN KEY (novelchapter_id) REFERENCES novelchapter(id)
		);`
	createNovelContentIndex = `
	CREATE UNIQUE INDEX IF NOT EXISTS site_chapter_chapterIndex
	ON novelcontent (novelsite_id, novelchapter_id, chapter_index)
	`
	InsertChapter = "INSERT INTO novelchapter(title, chapters, origin_url, link_prefix, domain, novelsite_id, createAt) values(?,?,?,?,?,?,?)"
	InsertSite    = "INSERT INTO novelsite(href, title, isParse, host, kw, createAt) values(?,?,?,?,?,?)"
	InsertContent = "INSERT INTO novelcontent(title, content, chapter_index, createAt, novelsite_id, novelchapter_id) values(?,?,?,?,?,?)"
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
	// 索引
	stmt, err = db.Prepare(createNovelContentIndex)
	checkErr(err)
	res, err = stmt.Exec()
	checkErr(err)
	fmt.Println("create novel content compose index success")
}
func InsertQuery(query string) (*sql.Stmt, error) {
	stmt, err := DBdf.Prepare(query)
	return stmt, err
}
func ExecWithStmt(stmt *sql.Stmt, param []interface{}) (sql.Result, error) {
	res, err := stmt.Exec(param...)
	return res, err
}
func Query(queryString string) (*sql.Rows, error) {
	return DBdf.Query(queryString)
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
