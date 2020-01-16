package searchengine

import (
	"fmt"
	"net/url"
	"novel/conf"
	"novel/fetcher"
	"novel/model"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

type BaiDuSearchEngine struct {
	parseRule       string
	searchRule      string
	domain          string
	parseResultFunc func(searchResult *model.SearchResult)
}

func NewBaiduSearchEngine(parseResultFunc func(result *model.SearchResult)) *BaiDuSearchEngine {
	return &BaiDuSearchEngine{
		parseRule:       "#content_left h3.t a",
		searchRule:      "intitle: %s 小说 阅读",
		domain:          "http://www.baidu.com/s?wd=%s&ie=utf-8&rn=15&vf_bl=1",
		parseResultFunc: parseResultFunc,
	}
}

func (engine *BaiDuSearchEngine) EngineRun(novelName string, group *sync.WaitGroup) {
	defer group.Done()
	searchKey := url.QueryEscape(fmt.Sprintf(engine.searchRule, novelName))
	requestUrl := fmt.Sprintf(engine.domain, searchKey)
	c := fetcher.NewCollector()
	fmt.Println("requestUrlrequestUrl: ", requestUrl)
	c.OnHTML(engine.parseRule, func(element *colly.HTMLElement) {
		group.Add(1)
		go engine.extractData(element, group)
	})
	err := c.Visit(requestUrl)
	if err != nil {
		fmt.Println(err)
	}
}
func (engine *BaiDuSearchEngine) extractData(element *colly.HTMLElement, group *sync.WaitGroup) {
	defer group.Done()
	href := element.Attr("href")
	title := element.Text
	// result := &model.SearchResult{Href: href, Title: title, IsParse: 1, Host: "www.baidu"}
	// engine.parseResultFunc(result)
	c := fetcher.NewCollector()
	c.OnResponse(func(response *colly.Response) {
		realURL := response.Request.URL.String()
		isContain := strings.Contains(realURL, "baidu")
		if isContain {
			return
		}
		host := response.Request.URL.Host
		_, ok := conf.RuleConfig.IgnoreDomain[host]
		if ok {
			return
		}
		isParse := engine.CheckIsParse(host)
		result := &model.SearchResult{Href: realURL, Title: title, IsParse: isParse, Host: host}
		engine.parseResultFunc(result)
		// _, ok := conf.RuleConfig.IgnoreDomain[host]
	})
	err := c.Visit(href)
	if err != nil {
		fmt.Println(err)
	}
}

func (engine *BaiDuSearchEngine) CheckIsParse(host string) bool {
	isParse := false
	for key := range conf.RuleConfig.Rule {
		if host == key {
			isParse = true
			break
		}
	}
	return isParse
}
