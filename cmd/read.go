package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"novel/common"
	"novel/conf"
	"novel/db"
	"novel/fetcher"
	"novel/model"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"jaytaylor.com/html2text"

	"github.com/huichen/sego"

	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "select list novels or search novel to read",
	Run:   ReadCommand,
}

// ChapterResultDB 章节数据结果对象
type ChapterResultDB struct {
	Chapter     model.NovelChapter
	ID          int64 `json:"id"`
	CreateAt    int64 `json:"createAt"`
	NovelSiteID int64
}

// ContentResultDB 内容数据结果对象
type ContentResultDB struct {
	Content      model.NovelContent
	ID           int64 `json:"id"`
	CreateAt     int64 `json:"createAt"`
	ChapterID    int64 `json:"chapter_id"`
	SiteID       int64
	ChapterINDEX int64
}

// ReadCommand hmmm
func ReadCommand(cmd *cobra.Command, args []string) {
	if NovelName == "" {
		fmt.Println("请输入小说名+Enter键: ")
		reader := bufio.NewReader(os.Stdin)
		kw, _ := reader.ReadString('\n')
		NovelName = kw
	}
	fmt.Println("您要找的小说是: ", NovelName)
	var segmenter sego.Segmenter
	segmenter.LoadDictionary("dictionary.txt")
	text := []byte(NovelName)
	segments := segmenter.Segment(text)
	// 生成like 搜索语句
	likeArray := []string{NovelName}
	likeString := fmt.Sprintf("n.title like '%%%s%%'", NovelName)
	for _, seg := range segments {
		likeArray = append(likeArray, seg.Token().Text())
		fmt.Printf("%+v \n", seg.Token().Text())
		likeString = fmt.Sprintf("%s or n.title like '%%%s%%'", likeString, seg.Token().Text())
	}
	query := fmt.Sprintf("SELECT * FROM novelchapter as n WHERE (%s)", likeString)
	rows, err := db.Query(query)
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	chapterResults := make([]*ChapterResultDB, 0)
	var askQs []string
	nextIndex := 0
	for rows.Next() {
		chapterResultDB := parseChapterResultDBByRows(rows)
		askQs = append(askQs, fmt.Sprintf("%d ||| %s %s", nextIndex, chapterResultDB.Chapter.Name, chapterResultDB.Chapter.OriginURL))
		nextIndex++
		chapterResults = append(chapterResults, chapterResultDB)
	}
	if len(chapterResults) == 0 {
		log.Fatal("没有找到可用的书本: ", strings.Join(likeArray, " "))
	}
	// 选取想要的网站和小说
	index := askSearchSiteTitleSelect(askQs)
	chapterResult := chapterResults[index]
	selectChapterToRead(chapterResult)
}

// 选取章节
func selectChapterToRead(chapterResult *ChapterResultDB) {
	var askQs []string
	nextIndex := 0
	if len(chapterResult.Chapter.Chapters) == 0 {
		log.Fatalf("小说 %s 没有找到章节。", chapterResult.Chapter.Name)
	}
	for _, chapterElement := range chapterResult.Chapter.Chapters {
		askQs = append(askQs, fmt.Sprintf("%d ||| %s %s", nextIndex, chapterElement.ChapterName, chapterElement.ChapterHref))
		nextIndex++
	}
	chapterIndex := askSearchSiteTitleSelect(askQs)
	Read(chapterResult, chapterIndex)
}

// Read 开始读取该章节
func Read(chapterResult *ChapterResultDB, chapterElementSelectIndex int64) {
	chapterElement := chapterResult.Chapter.Chapters[chapterElementSelectIndex]
	fmt.Printf("\n-----------------------------------------\n   ****%s****   \n-----------------------------------------\n", chapterElement.ChapterName)
	contentDBResult, err := getContentDBResult(chapterResult, chapterElementSelectIndex)
	htmlText, err := html2text.FromString(contentDBResult.Content.Content, html2text.Options{OmitLinks: true})
	if err != nil {
		log.Fatal("文章转义出错 ", err)
	}
	fmt.Println(htmlText)
outerloop:
	for {
		reader := bufio.NewReader(os.Stdin)
		readStr, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("dododo err: ", err)
		}
		readStrByte := []byte(readStr)
		switch readStrByte[0] {
		case 'q':
			// 返回上一层
			selectChapterToRead(chapterResult)
			break outerloop
		case 'a':
			// 选取上一页
			chapterElementSelectIndex--
			if chapterElementSelectIndex <= 0 {
				log.Println("已经是第一章/页,不能再往前了")
				return
			}
			Read(chapterResult, chapterElementSelectIndex)
		case 'd':
			// 选取下一页
			chapterElementSelectIndex++
			chaptersLen := int64(len(chapterResult.Chapter.Chapters))
			if chapterElementSelectIndex >= chaptersLen {
				log.Println("已经是最后一章/页, 不能再往后了")
				return
			}
			Read(chapterResult, chapterElementSelectIndex)
		}
	}
}

// 根据用户选取文章对应章节，得到该章节数据
func getContentDBResult(chapterResult *ChapterResultDB, chapterElementSelectIndex int64) (*ContentResultDB, error) {
	var contentResultDB ContentResultDB
	var queryStr = fmt.Sprintf("SELECT * FROM novelcontent WHERE (chapter_index=%d AND novelsite_id=%d AND novelchapter_id=%d) LIMIT 1;", chapterElementSelectIndex, chapterResult.NovelSiteID, chapterResult.ID)
	rows, err := db.Query(queryStr)
	defer rows.Close()
	if err != nil {
		log.Fatal("query database content err:", err)
	}
	if rows.Next() {
		contentResultDB = *parseContentResultDBByRows(rows)
		return &contentResultDB, nil
	}
	// 数据库没有该数据，从网络获取
	chapterElement := chapterResult.Chapter.Chapters[chapterElementSelectIndex]
	contentResult, err := parseNovelContent(chapterResult, chapterElement)
	if err != nil {
		log.Fatal("解析文章正文出错: ", err)
	}

	stmt, err := db.InsertQuery(db.InsertContent)
	if err != nil {
		log.Fatal("insert content err: ", err)
	}
	nowTime := time.Now().UnixNano() / 1e6
	_, err = db.ExecWithStmt(stmt, []interface{}{chapterResult.Chapter.Name, contentResult.Content, chapterElementSelectIndex, nowTime, chapterResult.NovelSiteID, chapterResult.ID})
	if err != nil {
		log.Fatal("save content err: ", err)
	}
	return getContentDBResult(chapterResult, chapterElementSelectIndex)
}

// 根据sql.Rows 转换得到数据库对象
func parseContentResultDBByRows(rows *sql.Rows) *ContentResultDB {
	var id, novelsiteID, chapterIndex, novelchapterID, createAt int64
	var title, content string
	_ = rows.Scan(&id, &title, &content, &chapterIndex, &createAt, &novelsiteID, &novelchapterID)
	return &ContentResultDB{
		Content: model.NovelContent{
			NovelName:   "",
			Title:       title,
			ContentURL:  "",
			Content:     content,
			PreChapter:  "",
			NextChapter: "",
		},
		ID:           id,
		SiteID:       novelsiteID,
		ChapterID:    novelchapterID,
		ChapterINDEX: chapterIndex,
	}
}

// 解析文章content
func parseNovelContent(chapter *ChapterResultDB, chapterElement *model.NovelChapterElement) (*model.NovelContent, error) {
	var novelContent model.NovelContent
	var html string
	if chapter.Chapter.LinkPrefix == "1" {
		html = chapterElement.ChapterHref
	} else if chapter.Chapter.LinkPrefix == "-1" {
		// html = "www.baidu.com"
		html = common.UrlJoin(chapterElement.ChapterHref, chapter.Chapter.OriginURL)
	} else if chapter.Chapter.LinkPrefix == "0" {
		html = common.UrlJoin(chapterElement.ChapterHref, chapter.Chapter.Domain)
	}
	c := fetcher.NewCollector()
	requestURI, _ := url.ParseRequestURI(chapter.Chapter.Domain)
	host := requestURI.Host
	contentSelector, _ := conf.RuleConfig.Rule[host]["content_selector"].(string)
	if contentSelector == "" {
		return &novelContent, fmt.Errorf("parseNovelContent %s", "contentSelector is empty")
	}
	c.OnHTML(contentSelector, func(element *colly.HTMLElement) {
		html, err := element.DOM.Html()
		if err != nil {
			log.Fatal("解析文章正文遇到出错 : ", err)
		}
		novelContent.Content = html
		novelContent.Title = chapterElement.ChapterName
		novelContent.ContentURL = chapterElement.ChapterHref
		novelContent.NovelName = chapter.Chapter.Name
	})
	err := c.Visit(html)
	return &novelContent, err
}

func init() {
	RootCmd.AddCommand(readCmd)
}
