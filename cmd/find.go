package cmd

import (
	"bufio"
	"fmt"
	"log"
	"novel/db"
	"novel/model"
	"os"
	"sync"
	"time"

	"novel/service/searchengine"

	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: `find novel name`,
	Run:   FindCommand,
}

func FindCommand(cmd *cobra.Command, args []string) {
	if NovelName != "" {
		startSearchEngine(NovelName)
		afterSearchNovel(NovelName)
	} else {
		GotoFind()
	}
}

func GotoFind() {
	fmt.Println("请输入小说名+Enter键: ")
	var keyword string
	for {
		reader := bufio.NewReader(os.Stdin)
		kw, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("read err: ", err)
		}
		startSearchEngine(kw)
		keyword = kw
		if len(searchResults) > 0 {
			break
		}
	}
	afterSearchNovel(keyword)
}

func afterSearchNovel(keyword string) {
	var searchSiteResultArray []string
	var searchSiteResults []*SearchResultDB
	for index, result := range searchResults {
		askStr := fmt.Sprintf("%d ||| %s %s", index, result.Title, result.Host)
		searchSiteResultArray = append(searchSiteResultArray, askStr)
		stmt, err := db.InsertQuery(db.InsertSite)
		if err != nil {
			log.Fatal(err)
		}
		nowTime := time.Now().UnixNano() / 1e6
		res, err := db.ExecWithStmt(stmt, []interface{}{result.Href, result.Title, true, result.Host, keyword, nowTime})
		if err != nil {
			log.Fatal("database exec meet err: ", err)
		}
		id, _ := res.LastInsertId()
		searchSiteResults = append(searchSiteResults, &SearchResultDB{
			ID:           id,
			SearchResult: *result,
		})
	}
	ToReadBySearchResults(searchSiteResults)
}

func readRuneFunc() rune {
	char, _, err := bufio.NewReader(os.Stdin).ReadRune()
	if err != nil {
		fmt.Println("error reading char: ", err)
	}
	return char
}

var searchResults []*model.SearchResult
var searchResultIndex = 0

type EngineSearch interface {
	EngineRun(string, *sync.WaitGroup)
}

func startSearchEngine(novelName string) []*model.SearchResult {
	fmt.Println("您要找的小说是: ", NovelName)
	group := sync.WaitGroup{}
	results := make([]*model.SearchResult, 0)
	group.Add(1)
	searchEngine := searchengine.NewBaiduSearchEngine(func(result *model.SearchResult) {
		results = append(results, result)
	})
	go searchEngine.EngineRun(novelName, &group)
	group.Wait()
	searchResults = results
	if len(results) == 0 {
		fmt.Println("当前没有找到被解析的小说网站, 请联系aker QQ mail:767838865@qq.com")
		os.Exit(1)
	}
	return results
}

func init() {
	RootCmd.AddCommand(findCmd)
}
