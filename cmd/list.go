package cmd

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"novel/conf"
	"novel/db"
	"novel/fetcher"
	"novel/model"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"

	"github.com/AlecAivazis/survey/v2"

	"github.com/huichen/sego"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: `list all or list novel name`,
	Run:   ListCommand,
}

type SearchResultDB struct {
	SearchResult model.SearchResult
	ID           int64 `json:"id"`
}

func ListCommand(cmd *cobra.Command, args []string) {
	fmt.Println("novelname :::", NovelName)
	var query string
	if NovelName == "" {
		reader := bufio.NewReader(os.Stdin)
		kw, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err.Error())
		}
		NovelName = kw
	}
	var segmenter sego.Segmenter
	segmenter.LoadDictionary("dictionary.txt")
	text := []byte(NovelName)
	segments := segmenter.Segment(text)
	var likeString = fmt.Sprintf("n.title like '%%%s%%'", NovelName)
	for _, seg := range segments {
		fmt.Printf("%+v \n", seg.Token().Text())
		likeString = fmt.Sprintf("%s or n.title like '%%%s%%'", likeString, seg.Token().Text())
	}
	fmt.Println("likeString:::", likeString)
	query = fmt.Sprintf("SELECT * FROM novelsite as n WHERE (%s)", likeString)
	fmt.Println("##########", query)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	searchResults := make([]*SearchResultDB, 0)
	var askQs []string
	nextIndex := 0
	for rows.Next() {
		var id, createAt int64
		var href, title, host, kw string
		var isParse bool
		_ = rows.Scan(&id, &href, &title, &isParse, &host, &kw, &createAt)
		fmt.Println(id)
		fmt.Println(createAt)
		fmt.Println(href)
		fmt.Println(title)
		fmt.Println(host)
		fmt.Println(kw)
		fmt.Println(isParse)
		searchResults = append(searchResults, &SearchResultDB{
			SearchResult: model.SearchResult{
				Href:    href,
				Title:   title,
				IsParse: isParse,
				Host:    host,
			},
			ID: id,
		})
		askQs = append(askQs, fmt.Sprintf("%d ||| %s %s", nextIndex, title, host))
		nextIndex++
	}
	fmt.Println("askQs", askQs)
	selectIndex := askSearchSiteTitleSelect(askQs)
	fmt.Println("-------------", selectIndex)

	searchResult := searchResults[selectIndex]
	// chapter, err := parseNovelChapter(searchResult.Href, searchResult.Title)
	chapterDBResult, err := getChapterDBBySearchResult(searchResult)
	if err != nil {
		log.Fatal("^^^^^^^^^", err)
	}
	for _, chapterElement := range chapterDBResult.Chapter.Chapters {
		askQs = append(askQs, fmt.Sprintf("%d ||| %s %s", nextIndex, chapterElement.ChapterName, chapterElement.ChapterHref))
		nextIndex++
	}
	chapterIndex := askSearchSiteTitleSelect(askQs)

	Read(chapterDBResult, chapterIndex)
}
func parseNovelChapter(searchResult *SearchResultDB) (*model.NovelChapter, error) {
	var novelChapter model.NovelChapter
	c := fetcher.NewCollector()
	requestURI, err := url.ParseRequestURI(searchResult.SearchResult.Href)
	if err != nil {
		fmt.Println("111111", err)
		return &novelChapter, err
	}
	host := requestURI.Host
	chapterSelector, ok := conf.RuleConfig.Rule[host]["chapter_selector"].(string)
	if !ok {
		fmt.Println("22222", ok)
		return &novelChapter, err
	}
	chapterSelector = chapterSelector + " a"
	var chapterElements []*model.NovelChapterElement
	c.OnHTML(chapterSelector, func(element *colly.HTMLElement) {
		html := element.Attr("href")
		if html == "" {
			fmt.Println("33333333", "无效dom")
			return
		}
		var chapterElement model.NovelChapterElement
		chapterElement.ChapterName = element.Text
		chapterElement.ChapterHref = html
		chapterElements = append(chapterElements, &chapterElement)
	})
	err = c.Visit(searchResult.SearchResult.Href)
	novelChapter.Chapters = chapterElements
	novelChapter.Name = searchResult.SearchResult.Title
	novelChapter.OriginUrl = searchResult.SearchResult.Href
	novelChapter.LinkPrefix = conf.RuleConfig.Rule[host]["link_prefix"].(string)
	novelChapter.Domain = fmt.Sprintf("%s://%s", requestURI.Scheme, requestURI.Host)
	return &novelChapter, err
}

func getChapterDBBySearchResult(searchResult *SearchResultDB) (*ChapterResultDB, error) {
	var chapterDBResult ChapterResultDB
	queryStr := fmt.Sprintf("SELECT * FROM novelchapter WHERE (novelsite_id=%d AND title like '%%%s%%') LIMIT 1;", searchResult.ID, searchResult.SearchResult.Title)
	rows, err := db.Query(queryStr)
	if err != nil {
		log.Fatal("---=======================", err)
	}
	if rows.Next() {
		fmt.Println("----")
		chapterDBResult = *parseChapterResultDBByRows(rows)
		return &chapterDBResult, nil
	}
	// 去网络获取，同时生成ChapterResultDB
	chapterResult, err := parseNovelChapter(searchResult)
	if err != nil {
		log.Fatal(")))))))))", err)
	}
	if len(chapterResult.Chapters) == 0 {
		fmt.Println("获取章节失败, 请试用其他网站")
		return &chapterDBResult, nil
	}
	id, err := saveNovelChapter(chapterResult, searchResult)
	if id < 0 || err != nil {
		log.Fatal("保存章节失败")
		return &chapterDBResult, nil
	}
	chapterDBResult = ChapterResultDB{
		ID:           id,
		CreateAt:     0,
		NovelSite_ID: searchResult.ID,
		Chapter:      *chapterResult,
	}
	return &chapterDBResult, nil
}

func parseChapterResultDBByRows(rows *sql.Rows) *ChapterResultDB {
	var id, novelsite_id, createAt int64
	var title, chapters, origin_url, link_prefix, domain string
	_ = rows.Scan(&id, &title, &chapters, &origin_url, &link_prefix, &domain, &createAt, &novelsite_id)
	var chapterElements []*model.NovelChapterElement
	byteData := []byte(chapters)
	if err := json.Unmarshal(byteData, &chapterElements); err != nil {
		log.Fatal("-----", err)
	}
	return &ChapterResultDB{
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
	}
}

func saveNovelChapter(novelChapter *model.NovelChapter, searchResult *SearchResultDB) (int64, error) {
	var saveID = int64(-1)
	stmt, err := db.InsertQuery(db.InsertChapter)
	if err != nil {
		log.Println("saveNovelChapter)))))))", err)
		return saveID, err
	}
	nowTime := time.Now().UnixNano() / 1e6
	chapterstr, err := json.Marshal(novelChapter.Chapters)
	if err != nil {
		log.Printf("JSON MARSHALING failed: %s \n", err)
		return saveID, err
	}
	chapterResult, err := db.ExecWithStmt(stmt, []interface{}{novelChapter.Name, chapterstr, novelChapter.OriginUrl, novelChapter.LinkPrefix, novelChapter.Domain, searchResult.ID, nowTime})
	if err != nil {
		log.Println("----", err)
		return saveID, err
	}
	fmt.Println("[list.saveNoivelChapter] chapterResult:: ", chapterResult)
	saveID, err = chapterResult.LastInsertId()
	return saveID, err
}

func askSearchSiteTitleSelect(searchTitleResultArray []string) int64 {
	qs := []*survey.Question{
		{
			Name: "site",
			Prompt: &survey.Select{
				Message: "Choose a site to scrawler:",
				Options: searchTitleResultArray,
				Default: searchTitleResultArray[0],
			},
		},
	}
	answers := struct {
		ChooseSite string `survey:"site"`
	}{}
	err := survey.Ask(qs, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}
	fmt.Printf("%s chose %s. \n", "1111", answers.ChooseSite)
	indexStr := strings.Split(answers.ChooseSite, " ||| ")[0]
	index, _ := strconv.ParseInt(indexStr, 10, 64)
	fmt.Printf("+++++++ %d\n", index)
	return index
}

func init() {
	RootCmd.AddCommand(listCmd)
}
