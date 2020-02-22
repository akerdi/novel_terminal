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

// SearchResultDB 搜索数据结果对象
type SearchResultDB struct {
	SearchResult model.SearchResult
	ID           int64 `json:"id"`
}

// ListCommand hmmm
func ListCommand(cmd *cobra.Command, args []string) {
	var query string
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
	likeArray := []string{NovelName}
	var likeString = fmt.Sprintf("n.title like '%%%s%%'", NovelName)
	for _, seg := range segments {
		likeArray = append(likeArray, seg.Token().Text())
		likeString = fmt.Sprintf("%s or n.title like '%%%s%%'", likeString, seg.Token().Text())
	}
	query = fmt.Sprintf("SELECT * FROM novelsite as n WHERE (%s)", likeString)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("list 查询小说站点出错: ", err)
	}
	defer rows.Close()

	searchResults := make([]*SearchResultDB, 0)
	for rows.Next() {
		var id, createAt int64
		var href, title, host, kw string
		var isParse bool
		_ = rows.Scan(&id, &href, &title, &isParse, &host, &kw, &createAt)
		searchResults = append(searchResults, &SearchResultDB{
			SearchResult: model.SearchResult{
				Href:    href,
				Title:   title,
				IsParse: isParse,
				Host:    host,
			},
			ID: id,
		})
	}
	if len(searchResults) == 0 {
		log.Fatal("没有找到可用的书本: ", strings.Join(likeArray, " "))
	}
	ToReadBySearchResults(searchResults)
}

// ToReadBySearchResults 根据searchResults 去选取网站对应小说目录
func ToReadBySearchResults(searchResults []*SearchResultDB) {
	var askQs []string
	nextIndex := 0
	for _, searchResult := range searchResults {
		askQs = append(askQs, fmt.Sprintf("%d ||| %s %s", nextIndex, searchResult.SearchResult.Title, searchResult.SearchResult.Host))
		nextIndex++
	}
	selectIndex := askSearchSiteTitleSelect(askQs)
	searchResult := searchResults[selectIndex]
	askQs = askQs[:0]
	nextIndex = 0
	chapterDBResult, err := getChapterDBBySearchResult(searchResult)
	if err != nil {
		log.Fatal("获取本地书本失败: ", err)
	}
	for _, chapterElement := range chapterDBResult.Chapter.Chapters {
		askQs = append(askQs, fmt.Sprintf("%d ||| %s %s", nextIndex, chapterElement.ChapterName, chapterElement.ChapterHref))
		nextIndex++
	}
	chapterIndex := askSearchSiteTitleSelect(askQs)

	Read(chapterDBResult, chapterIndex)
}

// 根据网站对应小说目录，先查询本地是否有缓存，否则网络获取
func parseNovelChapter(searchResult *SearchResultDB) (*model.NovelChapter, error) {
	var novelChapter model.NovelChapter
	c := fetcher.NewCollector()
	requestURI, err := url.ParseRequestURI(searchResult.SearchResult.Href)
	if err != nil {
		return &novelChapter, err
	}
	host := requestURI.Host
	chapterSelector, ok := conf.RuleConfig.Rule[host]["chapter_selector"].(string)
	if !ok {
		return &novelChapter, err
	}
	chapterLinkPrefix, ok := conf.RuleConfig.Rule[host]["link_prefix"].(string)
	if !ok {
		return &novelChapter, err
	}
	var chapterElements []*model.NovelChapterElement
	c.OnHTML(chapterSelector, func(element *colly.HTMLElement) {
		html := element.Attr("href")
		if html == "" {
			fmt.Println("无效dom")
			return
		}
		var chapterElement model.NovelChapterElement
		chapterElement.ChapterName = element.Text
		chapterElement.ChapterHref = html
		chapterElements = append(chapterElements, &chapterElement)
	})
	var searchHref string = searchResult.SearchResult.Href
	if chapterTail, ok := conf.RuleConfig.Rule[host]["chapter_tail"].(string); ok {
		if containTail := strings.Contains(searchHref, chapterTail); !containTail {
			searchHref = fmt.Sprintf("%s%s", searchHref, chapterTail)
		}
	}
	fmt.Println("parseNovelChapter href: ", searchHref)
	err = c.Visit(searchHref)
	novelChapter.Chapters = chapterElements
	novelChapter.Name = searchResult.SearchResult.Title
	novelChapter.OriginUrl = searchResult.SearchResult.Href
	novelChapter.LinkPrefix = chapterLinkPrefix
	novelChapter.Domain = fmt.Sprintf("%s://%s", requestURI.Scheme, requestURI.Host)
	return &novelChapter, err
}

// 根据选择的网站对应的小说条目, 用户前往选取小说
func getChapterDBBySearchResult(searchResult *SearchResultDB) (*ChapterResultDB, error) {
	var chapterDBResult ChapterResultDB
	queryStr := fmt.Sprintf("SELECT * FROM novelchapter WHERE (novelsite_id=%d AND title like '%%%s%%') LIMIT 1;", searchResult.ID, searchResult.SearchResult.Title)
	rows, err := db.Query(queryStr)
	defer rows.Close()
	if err != nil {
		log.Fatal("getChapterDBBySearchResult err: ", err)
	}
	if rows.Next() {
		chapterDBResult = *parseChapterResultDBByRows(rows)
		return &chapterDBResult, nil
	}
	// 去网络获取，同时生成ChapterResultDB
	chapterResult, err := parseNovelChapter(searchResult)
	if err != nil {
		log.Fatal(err)
	}
	if len(chapterResult.Chapters) == 0 {
		fmt.Println("获取章节失败, 请试用其他网站")
		return &chapterDBResult, fmt.Errorf("获取章节失败 %d章节", 0)
	}
	id, err := saveNovelChapter(chapterResult, searchResult)
	if id < 0 || err != nil {
		log.Fatal("保存章节失败")
		return &chapterDBResult, nil
	}
	chapterDBResult = ChapterResultDB{
		ID:          id,
		CreateAt:    0,
		NovelSiteID: searchResult.ID,
		Chapter:     *chapterResult,
	}
	return &chapterDBResult, nil
}

// 根据sql.Rows 得到ChapterResultDB
func parseChapterResultDBByRows(rows *sql.Rows) *ChapterResultDB {
	var id, novelsiteID, createAt int64
	var title, chapters, originURL, linkPrefix, domain string
	_ = rows.Scan(&id, &title, &chapters, &originURL, &linkPrefix, &domain, &createAt, &novelsiteID)
	var chapterElements []*model.NovelChapterElement
	byteData := []byte(chapters)
	if err := json.Unmarshal(byteData, &chapterElements); err != nil {
		log.Fatal(err)
	}
	return &ChapterResultDB{
		Chapter: model.NovelChapter{
			Name:       title,
			OriginUrl:  originURL,
			Chapters:   chapterElements,
			LinkPrefix: linkPrefix,
			Domain:     domain,
		},
		ID:          id,
		NovelSiteID: novelsiteID,
		CreateAt:    createAt,
	}
}

// 保存选取的网站，保存该网站里面的对应小说完整数据: 章节信息
func saveNovelChapter(novelChapter *model.NovelChapter, searchResult *SearchResultDB) (int64, error) {
	var saveID = int64(-1)
	stmt, err := db.InsertQuery(db.InsertChapter)
	if err != nil {
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
		return saveID, err
	}
	saveID, err = chapterResult.LastInsertId()
	return saveID, err
}

// 输入字符串数据得到用户选取的index
func askSearchSiteTitleSelect(searchTitleResultArray []string) int64 {
	qs := []*survey.Question{
		{
			Name: "site",
			Prompt: &survey.Select{
				Message: "请您作出选择:",
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
		log.Fatal("survey meet error: ", err)
	}
	indexStr := strings.Split(answers.ChooseSite, " ||| ")[0]
	index, err := strconv.ParseInt(indexStr, 10, 64)
	if err != nil {
		fmt.Println("strconv parseInt meet error: ", err)
	}
	fmt.Printf("您选择了 %s", answers.ChooseSite)
	return index
}

func init() {
	RootCmd.AddCommand(listCmd)
}
