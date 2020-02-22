package cmd

import (
	"bufio"
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
	fmt.Println("searchResult:::: ", searchResults)
	selectIndex := askSearchSiteTitleSelect(askQs)
	fmt.Println("-------------", selectIndex)

	searchResult := searchResults[selectIndex]
	// chapter, err := parseNovelChapter(searchResult.Href, searchResult.Title)
	chapter, err := parseNovelChapter(searchResult)
	if err != nil {
		log.Fatal(")))))))))", err)
	}
	fmt.Printf("(((((((%s ", chapter)
	saveNovelChapter(chapter, searchResult)
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
	fmt.Println("++mmm", host)
	chapterSelector, ok := conf.RuleConfig.Rule[host]["chapter_selector"].(string)
	chapterSelector = chapterSelector + " a"
	if !ok {
		fmt.Println("22222", ok)
		return &novelChapter, err
	}
	var chapterElements []*model.NovelChapterElement
	c.OnHTML(chapterSelector, func(element *colly.HTMLElement) {
		html := element.Attr("href")
		fmt.Println(" text:: ", element.Text, "  html ::::", html)
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

func saveNovelChapter(novelChapter *model.NovelChapter, searchResult *SearchResultDB) {
	stmt, err := db.InsertQuery("INSERT INTO novelchapter(title, chapters, origin_url, link_prefix, domain, novelsite_id, createAt) values(?,?,?,?,?,?,?)")
	if err != nil {
		log.Fatal("saveNovelChapter)))))))", err)
	}
	nowTime := time.Now().UnixNano() / 1e6
	chapters, err := json.Marshal(novelChapter.Chapters)
	if err != nil {
		log.Fatalf("JSON MARSHALING failed: %s", err)
	}
	log.Println("---", chapters)
	_, err = db.ExecWithStmt(stmt, []interface{}{novelChapter.Name, chapters, novelChapter.OriginUrl, novelChapter.LinkPrefix, novelChapter.Domain, searchResult.ID, nowTime})
	if err != nil {
		log.Fatal("----", err)
	}
}

func askSearchSiteTitleSelect(searchTitleResultArray []string) int {
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
	index, _ := strconv.Atoi(indexStr)
	fmt.Printf("+++++++ %d\n", index)
	return index
}

func init() {
	RootCmd.AddCommand(listCmd)
}
