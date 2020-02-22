package cmd

import (
	"bufio"
	"fmt"
	"log"
	"novel/db"
	"novel/model"
	"os"
	"strconv"
	"strings"
	"sync"

	"novel/service/searchengine"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: `find novel name`,
	Run:   FindCommand,
}

func FindCommand(cmd *cobra.Command, args []string) {
	DoFind()
}

func DoFind() {
	var keyWord string
	for {
		reader := bufio.NewReader(os.Stdin)
		kw, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("read err: ", err)
		}
		fmt.Println("result:::", kw)
		startSearchEngine(kw)
		keyWord = kw
		if len(searchResults) > 0 {
			break
		}
	}
	var searchSiteResultArray []string
	for index, result := range searchResults {
		log.Printf("4result::: %+v \n", result)
		askStr := fmt.Sprintf("%d ||| %s %s", index, result.Title, result.Host)
		log.Println("askStr:::", askStr)
		searchSiteResultArray = append(searchSiteResultArray, askStr)
		stmt, err := db.InsertQuery("INSERT INTO novelsite(href, title, isParse, host, kw) values(?,?,?,?,?)")
		if err != nil {
			log.Fatal("))))))))", err)
		}
		res, err := db.ExecWithStmt(stmt, []interface{}{result.Href, result.Title, true, result.Host, keyWord})
		if err != nil {
			log.Fatal("meet err: ", err)
		}
		log.Println("======", res)

	}
	fmt.Println("searchSiteResultArray:::", searchSiteResultArray)
	fmt.Println("searchSiteResultArray[0]", searchSiteResultArray[0])
	askSearchSiteToSelect(searchSiteResultArray)
}

func askSearchSiteToSelect(searchSiteResultArray []string) {
	qs := []*survey.Question{
		{
			Name: "title",
			Prompt: &survey.Select{
				Message: "Choose a title:",
				Options: searchSiteResultArray,
				Default: searchSiteResultArray[0],
			},
		},
	}
	answers := struct {
		ChooseTitle string `survey:"title"`
	}{}
	err := survey.Ask(qs, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%s chose %s. \n", "1111", answers.ChooseTitle)
	indexStr := strings.Split(answers.ChooseTitle, " ||| ")[0]
	index, _ := strconv.Atoi(indexStr)
	fmt.Printf("+++++++ %d", index)

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
	fmt.Println("------", novelName)
	group := sync.WaitGroup{}
	results := make([]*model.SearchResult, 0)
	group.Add(1)
	fmt.Println("22222222")
	searchEngine := searchengine.NewBaiduSearchEngine(func(result *model.SearchResult) {
		fmt.Println("333333")
		results = append(results, result)
	})
	fmt.Println("4444444")
	go searchEngine.EngineRun(novelName, &group)
	fmt.Println("5555555")
	group.Wait()
	fmt.Println("6666666")
	searchResults = results
	if len(results) > 0 {
		fmt.Printf("------%v\n ", results)
	}
	return results
}

func init() {
	RootCmd.AddCommand(findCmd)
}
