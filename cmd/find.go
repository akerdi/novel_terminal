package cmd

import (
	"bufio"
	"fmt"
	"novel/model"
	"os"
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
	for {
		reader := bufio.NewReader(os.Stdin)
		result, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("read err: ", err)
		}
		fmt.Println("result:::", result)
		startSearchEngine(result)
		if len(searchResults) > 0 {
			break
		}
	}
	var hostArray []string
	for _, result := range searchResults {
		hostArray = append(hostArray, result.Href)
	}
	fmt.Println("hostArray:::", hostArray)
	fmt.Println("hostArrayhostArray", hostArray[0])
	qs := []*survey.Question{
		{
			Name: "site",
			Prompt: &survey.Select{
				Message: "Choose a site:",
				Options: hostArray,
				Default: hostArray[0],
			},
		},
	}
	ansers := struct {
		ChooseSite string `survey:"site"`
	}{}
	err := survey.Ask(qs, &ansers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%s chose %s.", "1111", ansers.ChooseSite)

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
