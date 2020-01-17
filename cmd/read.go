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

type ChapterResultDB struct {
	Chapter      model.NovelChapter
	ID           int64 `json:"id"`
	CreateAt     int64 `json:"createAt"`
	NovelSite_ID int64
}
type ContentResultDB struct {
	Content       model.NovelContent
	ID            int64 `json:"id"`
	CreateAt      int64 `json:"createAt"`
	Chapter_ID    int64 `json:"chapter_id"`
	Site_ID       int64
	Chapter_INDEX int64
}

func ReadCommand(cmd *cobra.Command, args []string) {
	fmt.Println("novelname: ::: ", NovelName)
	if NovelName == "" {
		reader := bufio.NewReader(os.Stdin)
		kw, _ := reader.ReadString('\n')
		NovelName = kw
	}
	var segmenter sego.Segmenter
	segmenter.LoadDictionary("dictionary.txt")
	text := []byte(NovelName)
	segments := segmenter.Segment(text)
	likeString := fmt.Sprintf("n.title like '%%%s%%'", NovelName)
	for _, seg := range segments {
		fmt.Printf("%+v \n", seg.Token().Text())
		likeString = fmt.Sprintf("%s or n.title like '%%%s%%'", likeString, seg.Token().Text())
	}
	fmt.Println("likeString::: ", likeString)
	query := fmt.Sprintf("SELECT * FROM novelchapter as n WHERE (%s)", likeString)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	chapterResults := make([]*ChapterResultDB, 0)
	var askQs []string
	nextIndex := 0
	for rows.Next() {
		chapterResultDB := parseChapterResultDBByRows(rows)
		chapterResults = append(chapterResults, chapterResultDB)
		askQs = append(askQs, fmt.Sprintf("%d ||| %s %s", nextIndex, chapterResultDB.Chapter.Name, chapterResultDB.Chapter.OriginUrl))
		nextIndex++
	}

	index := askSearchSiteTitleSelect(askQs)
	fmt.Println("&&&&&&&&", index)
	askQs = []string{}
	nextIndex = 0
	if len(chapterResults[index].Chapter.Chapters) == 0 {
		log.Fatalf("小说 %s 没有找到章节。", chapterResults[index].Chapter.Name)
	}
	chapterResult := chapterResults[index]
	for _, chapterElement := range chapterResult.Chapter.Chapters {
		askQs = append(askQs, fmt.Sprintf("%d ||| %s %s", nextIndex, chapterElement.ChapterName, chapterElement.ChapterHref))
		nextIndex++
	}
	chapterIndex := askSearchSiteTitleSelect(askQs)
	Read(chapterResult, chapterIndex)
}

func Read(chapterResult *ChapterResultDB, chapterElementSelectIndex int64) {
	contentDBResult, err := getContentDBResult(chapterResult, chapterElementSelectIndex)
	htmlText, err := html2text.FromString(contentDBResult.Content.Content, html2text.Options{OmitLinks: true})
	if err != nil {
		log.Fatal("===++++++,,,, ", err)
	}
	fmt.Println(htmlText)
	// 保存当前
	// 读取前后章节，保存
	readyForNextAndPreviewNovelContent(chapterElementSelectIndex, chapterResult.Chapter.Chapters)
}
func getContentDBResult(chapterResult *ChapterResultDB, chapterElementSelectIndex int64) (*ContentResultDB, error) {
	var contentResultDB ContentResultDB
	var queryStr = fmt.Sprintf("SELECT * FROM novelcontent WHERE (chapter_index=%d AND novelsite_id=%d AND novelchapter_id=%d) LIMIT 1;", chapterElementSelectIndex, chapterResult.NovelSite_ID, chapterResult.ID)
	rows, err := db.Query(queryStr)
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
		log.Fatal("=========", err)
	}
	fmt.Printf("     &&&& %s", contentResult.Content)

	stmt, err := db.InsertQuery(db.InsertContent)
	if err != nil {
		log.Fatal("insert content err: ", err)
	}
	nowTime := time.Now().UnixNano() / 1e6
	_, err = db.ExecWithStmt(stmt, []interface{}{chapterResult.Chapter.Name, contentResult.Content, chapterElementSelectIndex, nowTime, chapterResult.NovelSite_ID, chapterResult.ID})
	if err != nil {
		log.Fatal("save content err:: ", err)
	}
	return getContentDBResult(chapterResult, chapterElementSelectIndex)
}

func readyForNextAndPreviewNovelContent(targetIndex int64, chapterResults []*model.NovelChapterElement) {

}

func parseContentResultDBByRows(rows *sql.Rows) *ContentResultDB {
	var id, novelsite_id, chapter_index, novelchapter_id, createAt int64
	var title, content string
	_ = rows.Scan(&id, &title, &content, &chapter_index, &createAt, &novelsite_id, &novelchapter_id)
	return &ContentResultDB{
		Content: model.NovelContent{
			NovelName:   "",
			Title:       title,
			ContentURL:  "",
			Content:     content,
			PreChapter:  "",
			NextChapter: "",
		},
		ID:            id,
		Site_ID:       novelsite_id,
		Chapter_ID:    novelchapter_id,
		Chapter_INDEX: chapter_index,
	}
}

func parseNovelContent(chapter *ChapterResultDB, chapterElement *model.NovelChapterElement) (*model.NovelContent, error) {
	var novelContent model.NovelContent
	var html string
	if chapter.Chapter.LinkPrefix == "1" {
		html = chapterElement.ChapterHref
	} else if chapter.Chapter.LinkPrefix == "-1" {
		// html = "www.baidu.com"
		html = common.UrlJoin(chapterElement.ChapterHref, chapter.Chapter.OriginUrl)
	} else if chapter.Chapter.LinkPrefix == "0" {
		html = common.UrlJoin(chapterElement.ChapterHref, chapter.Chapter.Domain)
	}
	fmt.Println("href html:: ", html)
	c := fetcher.NewCollector()
	requestURI, _ := url.ParseRequestURI(chapter.Chapter.Domain)
	host := requestURI.Host
	fmt.Println("--------------", host)
	contentSelector, _ := conf.RuleConfig.Rule[host]["content_selector"].(string)
	if contentSelector == "" {
		return &novelContent, fmt.Errorf("parseNovelContent %s", "contentSelector is empty")
	}
	fmt.Println("+++++++:", contentSelector)
	c.OnHTML(contentSelector, func(element *colly.HTMLElement) {
		html, err := element.DOM.Html()
		if err != nil {
			log.Fatal("=====111 : ", err)
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
