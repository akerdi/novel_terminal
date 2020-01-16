package cmd

import (
	"bufio"
	"encoding/json"
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
		var id, novelsite_id, createAt int64
		var title, chapters, origin_url, link_prefix, domain string
		_ = rows.Scan(&id, &title, &chapters, &origin_url, &link_prefix, &domain, &createAt, &novelsite_id)
		var chapterElements []*model.NovelChapterElement
		byteData := []byte(chapters)
		if err := json.Unmarshal(byteData, &chapterElements); err != nil {
			log.Fatal("JSON UNMarshaling failed:: ", err)
		}
		chapterResults = append(chapterResults, &ChapterResultDB{
			Chapter: model.NovelChapter{
				Name:       title,
				OriginUrl:  origin_url,
				Chapters:   chapterElements,
				LinkPrefix: link_prefix,
				Domain:     domain,
			},
			ID:           id,
			NovelSite_ID: novelsite_id,
			CreateAt:     createAt,
		})
		askQs = append(askQs, fmt.Sprintf("%d ||| %s %s", nextIndex, title, origin_url))
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

	chapterElement := chapterResult.Chapter.Chapters[chapterIndex]
	Read(chapterResult, chapterElement)
}

func Read(chapterResult *ChapterResultDB, chapterElement *model.NovelChapterElement) {
	contentResult, err := parseNovelContent(chapterResult, chapterElement)
	if err != nil {
		log.Fatal("=========", err)
	}
	// fmt.Printf("     &&&& %s", contentResult)
	htmlText, err := html2text.FromString(contentResult.Content, html2text.Options{OmitLinks: true})
	if err != nil {
		log.Fatal("===++++++,,,, ", err)
	}
	stmt, err := db.InsertQuery(db.InsertContent)
	if err != nil {
		log.Fatal("insert content err: ", err)
	}
	nowTime := time.Now().UnixNano() / 1e6
	_, err = db.ExecWithStmt(stmt, []interface{}{chapterResult.Chapter.Name, htmlText, nowTime, chapterResult.NovelSite_ID, chapterResult.ID})
	if err != nil {
		log.Fatal("save content err:: ", err)
	}
	// 保存当前
	// 读取前后章节，保存
	go fmt.Println(htmlText)
}

func readyForNextAndPreviewNovelContent(targetIndex int64, chapterResults []*ChapterResultDB) {

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
	fmt.Println("+++++++", contentSelector)
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
